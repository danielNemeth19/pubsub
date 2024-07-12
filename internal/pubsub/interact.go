package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"pubsub/internal/routing"

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

type AckType string

const (
	Ack         AckType = "ack"
	NackRequeue AckType = "nackrequeue"
	NackDiscard AckType = "nackdiscard"
)

func GetDeadLetterConfig() amqp.Table {
	return amqp.Table{"x-dead-letter-exchange": routing.ExchangePerilDlx}
}

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

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
    var packet bytes.Buffer
    encoder := gob.NewEncoder(&packet)
    err := encoder.Encode(val)

	if err != nil {
		panic("Gob encoding failed")
	}
	err = ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false, false,
		amqp.Publishing{ContentType: "application/gob", Body: packet.Bytes()},
	)
	return err
}

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int, table amqp.Table) (*amqp.Channel, amqp.Queue, error) {
	chn, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("Channel creation failed: %w", err)
	}

	fmt.Printf("Creating queue with name: %s, durable: %v, autoDelete: %v, exclusive: %v\n",
		queueName, simpleQueueType == Durable, simpleQueueType == Transient, simpleQueueType == Transient)

	queue, err := chn.QueueDeclare(queueName, simpleQueueType == Durable, simpleQueueType == Transient, simpleQueueType == Transient, false, table)

	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("Queue declaration failed: %w", err)
	}
	fmt.Printf("Binding queue %s to exchange %s with key %s\n", queue.Name, exchange, key)
	err = chn.QueueBind(queue.Name, key, exchange, false, nil)
	if err != nil {
		fmt.Println(err)
		return nil, amqp.Queue{}, fmt.Errorf("Queue binding failed: %w", err)
	}
	fmt.Println("Queue bound successfully")
	return chn, queue, nil
}

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int, handler func(out T, conn *amqp.Connection) AckType) error {
	deadLetterTable := GetDeadLetterConfig()
	chn, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType, deadLetterTable)
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
	if err != nil {
		return fmt.Errorf("Failed to start consuming messages: %w\n", err)
	}
	go func() {
		for msg := range msgChannel {
			log.Printf("Received message: %s\n", string(msg.Body))
			var out T
			err := json.Unmarshal(msg.Body, &out)
			if err != nil {
				log.Printf("Failed to unmarshall message: %v\n", err)
				continue
			}
			log.Printf("Out message to call handler with: %v\n", out)
			ackType := handler(out, conn)
			switch ackType {
			case Ack:
				msg.Ack(false)
				log.Println("Message acknowledged")
			case NackRequeue:
				msg.Nack(false, true)
				log.Println("Message negatively acknowledged and re-queued")
			case NackDiscard:
				msg.Nack(false, false)
				log.Println("Message negatively acknowledged and discarded")
			}
		}
	}()
	return nil
}

func SubscribeGob[T any](conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int, handler func(out T) AckType) error {
	deadLetterTable := GetDeadLetterConfig()
	chn, _, err := DeclareAndBind(conn, exchange, queueName, "game_logs.*", Durable, deadLetterTable)
	if err != nil {
		panic("Error declaring and binding channel")
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
	if err != nil {
		return fmt.Errorf("Failed to start consuming messages: %w", err)
	}
	go func() {
		for msg := range msgChannel {
			log.Printf("Received message: %s\n", string(msg.Body))
			// var gl routing.GameLog
			var out T
            network := bytes.NewReader(msg.Body)
            dec := gob.NewDecoder(network)
            err := dec.Decode(&out)
			if err != nil {
                log.Printf("Failed to decode gob: %v\n", err)
				continue
			}
			log.Printf("Out message to call handler with: %v\n", out)
			ackType := handler(out)
			switch ackType {
			case Ack:
				msg.Ack(false)
				log.Println("Message acknowledged")
			case NackRequeue:
				msg.Nack(false, true)
				log.Println("Message negatively acknowledged and re-queued")
			case NackDiscard:
				msg.Nack(false, false)
				log.Println("Message negatively acknowledged and discarded")
			}
		}
	}()
    return nil
}

