package auth_handlers

import (
	"context"
	"github.com/educ-educ/gateway/internal/auth"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"io"
	"time"
)

type AddRoleHandler struct {
	ctx              context.Context
	ch               *amqp.Channel
	sendExchangeName string
}

func NewAddRoleHandler(ctx context.Context, conn *amqp.Connection) *AddRoleHandler {
	handler := &AddRoleHandler{
		ctx: ctx,
	}

	var err error
	handler.ch, err = conn.Channel()
	auth.FailOnError(err, "Failed to open a channel")

	handler.sendExchangeName = "AddRole"
	err = handler.ch.ExchangeDeclare(
		handler.sendExchangeName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	auth.FailOnError(err, "Cannot declare exchange")

	return handler
}

func (handler *AddRoleHandler) Handle(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	auth.FailOnError(err, "Cannot read request body")

	ctx, cancel := context.WithTimeout(handler.ctx, 2*time.Second)
	defer cancel()

	err = handler.ch.PublishWithContext(ctx,
		handler.sendExchangeName,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: auth.JSONContentType,
			Body:        bodyBytes,
		})
	auth.FailOnError(err, "Cannot send a message")
}
