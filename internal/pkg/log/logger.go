package log

import (
	"context"

	"go.uber.org/zap"
)

var logger *zap.Logger

func InitialiseLogger() {
	logger, _ = zap.NewProduction()
}

func getContextFields(ctx context.Context) []zap.Field {
	var fields []zap.Field

	if userID := ctx.Value("user-id"); userID != nil {
		if id, ok := userID.(string); ok && id != "" {
			fields = append(fields, zap.String("user-id", id))
		}
	}

	return fields

}

func ErrorWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	if logger == nil {
		InitialiseLogger()
	}
	contextFields := getContextFields(ctx)
	allFields := append(contextFields, fields...)
	logger.Error(msg, allFields...)
}
