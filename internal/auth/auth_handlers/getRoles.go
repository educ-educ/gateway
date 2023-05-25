package auth_handlers

import (
	"context"
	"fmt"
	"github.com/educ-educ/gateway/internal/auth"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"io"
	"time"
)

type GetRolesHandler struct {
	ctx              context.Context
	ch               *amqp.Channel
	sendExchangeName string
	recvExchangeName string
	recvQueueName    string
	deliveryChan     <-chan amqp.Delivery
}

func NewGetRolesHandler(ctx context.Context, conn *amqp.Connection) *GetRolesHandler {
	handler := &GetRolesHandler{
		ctx: ctx,
	}

	var err error
	handler.ch, err = conn.Channel()
	auth.FailOnError(err, "Failed to open a channel")

	handler.sendExchangeName = "GetRoles"
	err = handler.ch.ExchangeDeclare(
		handler.sendExchangeName, "fanout", true, false, false, false, nil,
	)
	auth.FailOnError(err, "Cannot declare send exchange")

	handler.recvExchangeName = "GetRolesReply"
	err = handler.ch.ExchangeDeclare(
		handler.recvExchangeName,
		"fanout", true, false, false, false, nil,
	)
	auth.FailOnError(err, "Cannot declare receive exchange")

	handler.recvQueueName = "GetRolesReply"
	_, err = handler.ch.QueueDeclare(
		handler.recvQueueName, true, false, false, false, nil,
	)
	auth.FailOnError(err, "Cannot declare queue")

	err = handler.ch.QueueBind(handler.recvQueueName, "", handler.recvExchangeName, false, nil)
	auth.FailOnError(err, "Failed to bind a queue")

	handler.deliveryChan, err = handler.ch.Consume(
		handler.recvQueueName, "", true, false, false, false, nil,
	)
	auth.FailOnError(err, "Failed to start consuming")

	return handler
}

func (handler *GetRolesHandler) Handle(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	auth.FailOnError(err, "Cannot read request body")

	ctx, cancel := context.WithTimeout(handler.ctx, 2*time.Second)
	defer cancel()

	headers := amqp.Table{}
	headers["reply_to"] = fmt.Sprint(auth.RmqConnTLS, handler.recvQueueName, "?exchange:", handler.recvExchangeName)

	err = handler.ch.PublishWithContext(ctx,
		handler.sendExchangeName,
		"",
		false,
		false,
		amqp.Publishing{
			Headers:     headers,
			ContentType: auth.JSONContentType,
			Body:        bodyBytes,
		})
	auth.FailOnError(err, "Cannot send a message")

	cancel()

	ctx, cancel = context.WithTimeout(handler.ctx, 2*time.Second)
	defer cancel()

	select {
	case message := <-handler.deliveryChan:
		_, err = c.Writer.Write(message.Body)
	case <-ctx.Done():
		_, err = c.Writer.Write([]byte(ctx.Err().Error()))
	}
}
