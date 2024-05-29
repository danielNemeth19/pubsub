package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
    Durable = iota
    Transient = iota
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	packet, err := json.Marshal(val)
	if err != nil {
		panic("Marshalling failed")
	}
	err = ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false, false,
		amqp.Publishing{ContentType: "application/json", Body: packet},
	)
	return err
}

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int)(*amqp.Channel, amqp.Queue, error) {
    chn, err := conn.Channel() 
    if err != nil {
        panic("Channel cannot be created")
    }
    queue, err := chn.QueueDeclare("", simpleQueueType == Durable, simpleQueueType == Transient, simpleQueueType == Transient, false, nil )
    if err != nil {
        panic("Queue declaration failed")
    }
    err = chn.QueueBind("", key, exchange, false, nil)
    if err != nil {
        panic("Bind did not work")
    }
    return chn, queue, err
}
