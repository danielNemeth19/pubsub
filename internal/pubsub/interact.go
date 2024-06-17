package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	Direct  = "direct"
	Fanout  = "fanout"
	Headers = "headers"
	Topic   = "topic"
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

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int, handler func(T)) error {
	chn, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
        return fmt.Errorf("Declare and bind failed within subscribe: %w", err)
	}
	msgChannel, err := chn.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	go func() {
		for msg := range msgChannel {
			fmt.Printf("Received message: %s\n", string(msg.Body))
            var out T
			json.Unmarshal(msg.Body, &out)
            fmt.Println(out)
            handler(out)
            msg.Ack(false)
		}
	}()
	return nil
}
