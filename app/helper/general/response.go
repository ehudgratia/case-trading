package general

import (
	"errors"

	"github.com/getsentry/sentry-go"
)

type Response struct {
	Status    bool   `json:"status"`
	Message   string `json:"message,omitempty"`
	Data      any    `json:"data,omitempty"`
	Total     int    `json:"total,omitempty"`
	NextState string `json:"next_state,omitempty"`
}

func RespOk(data any) Response {
	return Response{
		Status: true,
		Data:   data,
	}
}

func RespMsgOk(msg string) Response {
	return Response{
		Status:  true,
		Message: msg,
	}
}

func RespErr(err string) Response {
	sentry.CaptureException(errors.New(err))

	return Response{
		Status:  false,
		Message: err,
	}
}

func RespErrWithData(err string, data any) Response {
	sentry.CaptureException(errors.New(err))
	return Response{
		Status:  false,
		Message: err,
		Data:    data,
	}
}

func RespPageOk(data any, total int) Response {
	return Response{
		Status: true,
		Data:   data,
		Total:  total,
	}
}

func RespTotalOk(total int) Response {
	return Response{
		Status: true,
		Total:  total,
	}
}

func RespPageStateOk(data any, nextState string) Response {
	if nextState == "" {
		nextState = "last_page"
	}
	return Response{
		Status:    true,
		Data:      data,
		NextState: nextState,
	}
}
