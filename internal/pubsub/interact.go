package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	Durable   = iota
	Transient = iota
)

func CreateExchange(ch *amqp.Channel, name, exchangeType string, exchangeParam int) error {
	err := ch.ExchangeDeclare(name, exchangeType, exchangeParam == Durable, exchangeParam == Transient, false, false, nil)
	if err != nil {
		return err
	}
	return nil
}

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

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int) (*amqp.Channel, amqp.Queue, error) {
	chn, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("Channel creation failed: %w", err)
	}

	fmt.Printf("Creating queue with name: %s, durable: %v, autoDelete: %v, exclusive: %v\n",
		queueName, simpleQueueType == Durable, simpleQueueType == Transient, simpleQueueType == Transient)

	queue, err := chn.QueueDeclare(queueName, simpleQueueType == Durable, simpleQueueType == Transient, simpleQueueType == Transient, false, nil)

	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("Queue declaration failed: %w", err)
	}
	fmt.Printf("Binding queue %s to exchange %s with key %s\n", queueName, exchange, key)
	err = chn.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		fmt.Println(err)
		return nil, amqp.Queue{}, fmt.Errorf("Queue binding failed: %w", err)
	}
	fmt.Println("Queue bound successfully")
	return chn, queue, nil
}
