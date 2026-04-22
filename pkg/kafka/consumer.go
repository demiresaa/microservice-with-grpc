package kafka

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader  *kafka.Reader
	logger  *slog.Logger
	handler func(ctx context.Context, msg []byte) error
}

func NewConsumer(brokers []string, topic, groupID string, logger *slog.Logger) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3,
			MaxBytes: 10e6,
		}),
		logger: logger,
	}
}

func (c *Consumer) SetHandler(handler func(ctx context.Context, msg []byte) error) {
	c.handler = handler
}

func (c *Consumer) Consume(ctx context.Context) error {
	if c.handler == nil {
		return fmt.Errorf("handler not set for consumer")
	}

	c.logger.Info("starting kafka consumer", "topic", c.reader.Config().Topic)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("kafka consumer shutting down", "reason", ctx.Err())
			return c.reader.Close()
		default:
		}

		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			c.logger.Error("failed to read message", "error", err)
			continue
		}

		if err := c.handler(ctx, msg.Value); err != nil {
			c.logger.Error("failed to handle message",
				"error", err,
				"topic", msg.Topic,
				"partition", msg.Partition,
				"offset", msg.Offset,
			)
			continue
		}

		c.logger.Info("message processed successfully",
			"topic", msg.Topic,
			"partition", msg.Partition,
			"offset", msg.Offset,
		)
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
