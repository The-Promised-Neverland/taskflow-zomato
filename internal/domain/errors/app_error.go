package domain_error

import "fmt"

type ErrorCode string

type AppError struct {
	ErrorCode ErrorCode
	Message   string
	Cause     error
	ExtraData map[string]interface{}
}

func (ae *AppError) Error() string {
	return ae.GetMsg()
}

func (ae *AppError) Unwrap() error {
	return ae.Cause
}

func (ae *AppError) GetErrorCode() ErrorCode {
	return ae.ErrorCode
}

func (ae *AppError) GetCode() string {
	return string(ae.ErrorCode)
}

func (ae *AppError) AddExtraData(key string, val interface{}) {
	if ae.ExtraData == nil {
		ae.ExtraData = make(map[string]interface{})
	}
	ae.ExtraData[key] = val
}

func (ae *AppError) GetMsg() string {
	if ae.Message != "" {
		return ae.Message
	}

	var result string
	if msg, ok := msgMap[ae.ErrorCode]; ok {
		result = msg
		if ae.Cause != nil {
			result = fmt.Sprintf("%s: %s", result, ae.Cause.Error())
		}
	}

	if result == "" {
		if ae.Cause != nil {
			result = fmt.Sprintf("%s: %s", ae.GetCode(), ae.Cause.Error())
		} else {
			result = ae.GetCode()
		}
	}

	ae.Message = result
	return result
}
