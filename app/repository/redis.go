package repository

import (
	"case-trading/app/helper/config"
	"case-trading/app/models"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func (s *Service) MatchOrderInRedis(tx *gorm.DB, order *models.Order, market models.Market) ([]models.MatchResult, error) {
	scriptContent, err := os.ReadFile("helper/scripts/matching.lua")
	if err != nil {
		return nil, fmt.Errorf("failed to read lua script file: %w", err)
	}

	script := redis.NewScript(string(scriptContent))

	buysKey := fmt.Sprintf("orderbook:%d:buys", order.MarketID)
	sellsKey := fmt.Sprintf("orderbook:%d:sells", order.MarketID)

	keys := []string{buysKey, sellsKey}
	args := []interface{}{
		strconv.Itoa(order.ID),
		string(order.Side),
		fmt.Sprintf("%.8f", order.Price), // lebih presisi
		fmt.Sprintf("%.8f", order.Quantity),
	}

	// 1. Jalankan script
	cmd := script.Run(config.Ctx, config.RDB, keys, args...)

	// 2. Cek error eksekusi terlebih dahulu
	if cmd.Err() != nil {
		// Ini menangkap error kompilasi, koneksi, dll
		return nil, fmt.Errorf("failed to execute lua script: %w", cmd.Err())
	}

	// 3. Ambil hasil
	jsonStr := cmd.String()

	// 4. Debugging sederhana (hapus setelah OK)
	// fmt.Printf("Raw Lua result: %q\n", jsonStr)

	// 5. Cek apakah hasilnya terlihat seperti pesan error Redis
	if strings.Contains(jsonStr, "ERR ") || strings.Contains(jsonStr, "Error compiling script") || strings.HasPrefix(jsonStr, "eval ") {
		return nil, fmt.Errorf("lua script failed to compile or execute: %s", jsonStr)
	}

	// 6. Penanganan hasil normal
	var matches []models.MatchResult

	// Kosong = tidak ada match
	if jsonStr == "{}" || jsonStr == "" || jsonStr == "[]" {
		if err := s.syncOrderFromRedis(order); err != nil {
			return nil, err
		}
		return []models.MatchResult{}, nil
	}

	// Coba parse sebagai array
	if err := json.Unmarshal([]byte(jsonStr), &matches); err != nil {
		return nil, fmt.Errorf("invalid json from lua: %s (parse error: %w)", jsonStr, err)
	}

	// Sukses → sync order
	if err := s.syncOrderFromRedis(order); err != nil {
		return nil, err
	}

	return matches, nil
}

// Helper kecil agar kode lebih rapi
func (s *Service) syncOrderFromRedis(order *models.Order) error {
	redisOrder, err := getRedisOrder(order.ID)
	if err != nil {
		return fmt.Errorf("failed to sync order from redis: %w", err)
	}

	order.FilledQty = redisOrder.FilledQty
	order.Status = redisOrder.Status
	return nil
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
