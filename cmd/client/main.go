package main

import (
	"fmt"
	"os"
	"os/signal"
	"pubsub/internal/gamelogic"
	"pubsub/internal/pubsub"
	"pubsub/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	connStr := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connStr)
	if err != nil {
		panic("Error establishing connection")
	}
	defer conn.Close()

	username, _ := gamelogic.ClientWelcome()
	fmt.Printf("username is: %s\n", username)

	chn, _, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, pubsub.Transient)
	if err != nil {
		panic("Error declaring and binding channel")
	}
	defer chn.Close()

	msgChannel, err := chn.Consume(
		routing.PauseKey+"."+username, // queue
		"",                            // consumer
		true,                          // auto-ack
		false,                         // exclusive
		false,                         // no-local
		false,                         // no-wait
		nil,                           // args
	)
	if err != nil {
		panic("Error setting up consumer")
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	go func() {
		for msg := range msgChannel {
			fmt.Printf("Received message: %s\n", string(msg.Body))
		}
	}()
	fmt.Println("Client running... Press Ctr-C to exit.")
	<-signalChan
	fmt.Println("Received signal, exiting...")
}
