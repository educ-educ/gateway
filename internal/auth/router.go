package auth

import (
	"log"
)

const (
	JSONContentType string = "application/json"
	RmqConnTLS      string = "amqps://hljhaczs:J5W5ouprR8fKEkmGKvpmd0ijcs3BbX8J@toad.rmq.cloudamqp.com/"
)

type Role struct {
	RoleName    string `json:"RoleName"`
	Description string `json:"Description"`
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
