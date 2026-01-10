package repository

import (
	"case-trading/app/helper/config"
	"case-trading/app/models"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func (s *Service) MatchOrderInRedis(tx *gorm.DB, order *models.Order, market models.Market) ([]models.MatchResult, error) {

	script := redis.NewScript(`
		local buys_key = KEYS[1]
		local sells_key = KEYS[2]
		local order_id = ARGV[1]
		local side = ARGV[2]
		local price = tonumber(ARGV[3])
		local remaining_qty = tonumber(ARGV[4])
		local matches = {}

		local function match_buy()
			local sells = redis.call('ZRANGE', sells_key, 0, -1, 'WITHSCORES')
			for i=1,#sells,2 do
				local counter_id = sells[i]
				local counter_price = tonumber(sells[i+1])
				if counter_price > price then
					break
				end
				-- Fetch counter order from hash
				local counter_hash = redis.call('HGETALL', 'order:' .. counter_id)
				local counter = {}
				for j=1,#counter_hash,2 do
					counter[counter_hash[j]] = counter_hash[j+1]
				end
				local counter_remaining = tonumber(counter.quantity) - tonumber(counter.filled_qty)
				if counter_remaining <= 0 then
					redis.call('ZREM', sells_key, counter_id)
					goto continue
				end
				local match_qty = math.min(remaining_qty, counter_remaining)
				-- Update incoming order
				remaining_qty = remaining_qty - match_qty
				redis.call('HINCRBYFLOAT', 'order:' .. order_id, 'filled_qty', match_qty)
				-- Update counter
				redis.call('HINCRBYFLOAT', 'order:' .. counter_id, 'filled_qty', match_qty)
				counter_remaining = counter_remaining - match_qty
				if counter_remaining == 0 then
					redis.call('ZREM', sells_key, counter_id)
					redis.call('HSET', 'order:' .. counter_id, 'status', 'FILLED')
				end
				-- Record match
				table.insert(matches, {counter_id, counter_price, match_qty})
				if remaining_qty == 0 then
					redis.call('ZREM', buys_key, order_id)
					redis.call('HSET', 'order:' .. order_id, 'status', 'FILLED')
					break
				end
				::continue::
			end
		end

		local function match_sell()
			local buys = redis.call('ZREVRANGE', buys_key, 0, -1, 'WITHSCORES')
			for i=1,#buys,2 do
				local counter_id = buys[i]
				local counter_price = tonumber(buys[i+1]) * -1  -- since buys score is -price
				if counter_price < price then
					break
				end
				local counter_hash = redis.call('HGETALL', 'order:' .. counter_id)
				local counter = {}
				for j=1,#counter_hash,2 do
					counter[counter_hash[j]] = counter_hash[j+1]
				end
				local counter_remaining = tonumber(counter.quantity) - tonumber(counter.filled_qty)
				if counter_remaining <= 0 then
					redis.call('ZREM', buys_key, counter_id)
					goto continue
				end
				local match_qty = math.min(remaining_qty, counter_remaining)
				remaining_qty = remaining_qty - match_qty
				redis.call('HINCRBYFLOAT', 'order:' .. order_id, 'filled_qty', match_qty)
				redis.call('HINCRBYFLOAT', 'order:' .. counter_id, 'filled_qty', match_qty)
				counter_remaining = counter_remaining - match_qty
				if counter_remaining == 0 then
					redis.call('ZREM', buys_key, counter_id)
					redis.call('HSET', 'order:' .. counter_id, 'status', 'FILLED')
				end
				table.insert(matches, {counter_id, counter_price, match_qty})
				if remaining_qty == 0 then
					redis.call('ZREM', sells_key, order_id)
					redis.call('HSET', 'order:' .. order_id, 'status', 'FILLED')
					break
				end
				::continue::
			end
		end

		if side == 'BUY' then
			match_buy()
		else
			match_sell()
		end

		return cjson.encode(matches)
	`)

	buysKey := fmt.Sprintf("orderbook:%d:buys", order.MarketID)
	sellsKey := fmt.Sprintf("orderbook:%d:sells", order.MarketID)

	keys := []string{buysKey, sellsKey}
	args := []interface{}{strconv.Itoa(order.ID), string(order.Side), fmt.Sprintf("%f", order.Price), fmt.Sprintf("%f", order.Quantity)}

	result, err := script.Run(config.Ctx, config.RDB, keys, args...).Result()
	if err != nil {
		return nil, err
	}
	scriptResult, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("expected string result from lua script, got %T", result)
	}

	var matches []models.MatchResult
	if err := json.Unmarshal([]byte(scriptResult), &matches); err != nil {
		return nil, err
	}

	// Update order filled_qty from Redis
	redisOrder, err := getRedisOrder(order.ID)
	if err != nil {
		return nil, err
	}
	order.FilledQty = redisOrder.FilledQty
	order.Status = redisOrder.Status

	return matches, nil
}

func addToRedisOrderbook(order *models.Order) error {
	orderKey := fmt.Sprintf("order:%d", order.ID)
	redisOrder := models.RedisOrder{
		ID:        order.ID,
		UserID:    order.UserID,
		MarketID:  order.MarketID,
		Side:      string(order.Side),
		Price:     order.Price,
		Quantity:  order.Quantity,
		FilledQty: order.FilledQty,
		Status:    order.Status,
		CreatedAt: order.CreatedAt,
	}

	data, err := json.Marshal(redisOrder)
	if err != nil {
		return err
	}
	var hashMap map[string]interface{}
	json.Unmarshal(data, &hashMap)

	if _, err := config.RDB.HMSet(config.Ctx, orderKey, hashMap).Result(); err != nil {
		return err
	}

	var zsetKey string
	var score float64
	if order.Side == models.SideBuy {
		zsetKey = fmt.Sprintf("orderbook:%d:buys", order.MarketID)
		score = -order.Price // For descending order
	} else {
		zsetKey = fmt.Sprintf("orderbook:%d:sells", order.MarketID)
		score = order.Price // Ascending
	}

	_, err = config.RDB.ZAdd(config.Ctx, zsetKey, redis.Z{Score: score, Member: order.ID}).Result()
	return err
}

func removeFromRedisOrderbook(orderID int, side string, marketID int) error {
	var zsetKey string
	if side == string(models.SideBuy) {
		zsetKey = fmt.Sprintf("orderbook:%d:buys", marketID)
	} else {
		zsetKey = fmt.Sprintf("orderbook:%d:sells", marketID)
	}
	config.RDB.ZRem(config.Ctx, zsetKey, orderID)

	orderKey := fmt.Sprintf("order:%d", orderID)
	return config.RDB.Del(config.Ctx, orderKey).Err()
}

func updateRedisOrder(order *models.Order) error {
	orderKey := fmt.Sprintf("order:%d", order.ID)
	return config.RDB.HMSet(config.Ctx, orderKey, map[string]interface{}{
		"filled_qty": order.FilledQty,
		"status":     order.Status,
	}).Err()
}

func getRedisOrder(orderID int) (*models.RedisOrder, error) {
	orderKey := fmt.Sprintf("order:%d", orderID)
	data, err := config.RDB.HGetAll(config.Ctx, orderKey).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("order not found in redis")
	}

	bytes, _ := json.Marshal(data)
	var redisOrder models.RedisOrder
	json.Unmarshal(bytes, &redisOrder)
	return &redisOrder, nil
}
