package main

import (
	"fmt"
	"os"
	"os/signal"
	"pubsub/internal/pubsub"
	"pubsub/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	connStr := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connStr)
	defer conn.Close()
	if err != nil {
		panic("Error establishing connection")
	}
	fmt.Println("Rabbit has been caught...")
	myC, err := conn.Channel()
	if err != nil {
		panic("Rabbit channel failed to open")
	}
	err = pubsub.PublishJSON(myC, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
	if err != nil {
		panic("Publising failed")
	}
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
}
