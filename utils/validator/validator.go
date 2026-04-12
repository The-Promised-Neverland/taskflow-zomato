package validator

import (
	"context"
	"fmt"
	"sync"

	domain_error "taskflow/internal/domain/errors"
	"taskflow/utils/logger"

	"github.com/go-playground/validator/v10"
)

var (
	instance     *validator.Validate
	instanceOnce sync.Once
)

func getInstance() *validator.Validate {
	instanceOnce.Do(func() {
		instance = validator.New()
	})
	return instance
}

// Validate checks a struct against its validation tags.
// If validation fails, it logs the field issues and returns a domain AppError.
func Validate(ctx context.Context, input interface{}) error {
	err := getInstance().StructCtx(ctx, input)
	if err == nil {
		return nil
	}

	fields := make(map[string]string)
	for _, e := range err.(validator.ValidationErrors) {
		fields[e.Field()] = fmt.Sprintf("failed on '%s' validation", e.Tag())
	}

	logger.FromContext(ctx).WarnContext(ctx, "validation failed", "fields", fields)

	return domain_error.Raise(
		domain_error.CODE_VALIDATION_FAILED,
		"",
		nil,
		domain_error.WithExtraData(map[string]interface{}{"fields": fields}),
	)
}
