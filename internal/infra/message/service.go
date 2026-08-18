package message

import (
	"context"

	"github.com/clodoaldomarques/core-sdk/pkg/logger"
	"github.com/clodoaldomarques/core-sdk/pkg/sqs"
)

func Handler(ctx context.Context, msg *sqs.Message) error {
	logger.Info(ctx, "queue message", logger.Fields{
		"message": msg,
	})

	return nil
}
