package domain_error

import (
	"errors"
	"net/http"
)

func ExtractErrorCode(err error) ErrorCode {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.GetErrorCode()
	}
	return ErrorCode("")
}

func IsSameError(err error, code ErrorCode) bool {
	return code == ExtractErrorCode(err)
}

type HTTPError struct {
	Status    int                    `json:"-"`
	Code      string                 `json:"error_code"`
	Message   string                 `json:"message"`
	ExtraData map[string]interface{} `json:"data,omitempty"`
}

// GetHTTPError turns each error into an HTTPError with the right status code.
// AppErrors use ErrCodeToStatusMap; unknown app errors fall back to 400.
// Non-AppErrors fall back to 500.
func GetHTTPError(err error) HTTPError {
	var ae *AppError
	if errors.As(err, &ae) {
		status, ok := ErrCodeToStatusMap[ae.ErrorCode]
		if !ok {
			status = http.StatusBadRequest
		}
		return HTTPError{
			Status:    status,
			Code:      ae.GetCode(),
			Message:   ae.GetMsg(),
			ExtraData: ae.ExtraData,
		}
	}

	return HTTPError{
		Status:  http.StatusInternalServerError,
		Code:    string(CODE_UNKNOWN_ERROR),
		Message: err.Error(),
	}
}
