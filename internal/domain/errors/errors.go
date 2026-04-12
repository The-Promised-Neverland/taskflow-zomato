package domain_error

import "fmt"

type ErrorOption func(*AppError)

func WithExtraData(data map[string]interface{}) ErrorOption {
	return func(ae *AppError) {
		ae.ExtraData = data
	}
}

// Raise builds an AppError. msg replaces the default message from msgMap.
// Use an empty string to keep the default message.
func Raise(code ErrorCode, msg string, cause error, opts ...ErrorOption) error {
	ae := &AppError{
		ErrorCode: code,
		Cause:     cause,
	}

	if msg != "" {
		ae.Message = msg + ": " + ae.GetMsg()
	} else {
		ae.Message = ae.GetMsg()
	}

	for _, opt := range opts {
		opt(ae)
	}

	if cause == nil {
		ae.Cause = fmt.Errorf("%s: %s", ae.GetCode(), ae.GetMsg()) //nolint:err113
	}

	return ae
}
