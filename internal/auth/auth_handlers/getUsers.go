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

type GetUsersHandler struct {
	ctx              context.Context
	ch               *amqp.Channel
	sendExchangeName string
	recvExchangeName string
	recvQueueName    string
	deliveryChan     <-chan amqp.Delivery
}

func NewGetUsersHandler(ctx context.Context, conn *amqp.Connection) *GetUsersHandler {
	handler := &GetUsersHandler{
		ctx: ctx,
	}

	var err error
	handler.ch, err = conn.Channel()
	auth.FailOnError(err, "Failed to open a channel")

	handler.sendExchangeName = "GetUsers"
	err = handler.ch.ExchangeDeclare(
		handler.sendExchangeName, "fanout", true, false, false, false, nil,
	)
	auth.FailOnError(err, "Cannot declare send exchange")

	handler.recvExchangeName = "GetUsersReply"
	err = handler.ch.ExchangeDeclare(
		handler.recvExchangeName,
		"fanout", true, false, false, false, nil,
	)
	auth.FailOnError(err, "Cannot declare receive exchange")

	handler.recvQueueName = "GetUsersReply"
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

func (handler *GetUsersHandler) Handle(c *gin.Context) {
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
