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
	fmt.Println("Starting Peril server...")
	connStr := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connStr)
	defer conn.Close()
	if err != nil {
		fmt.Println("Error establishing connection")
	}
	myC, err := conn.Channel()
	if err != nil {
		fmt.Println("Rabbit channel failed to open")
	}
	err = pubsub.CreateExchange(myC, routing.ExchangePerilDirect, "direct", pubsub.Durable)
	if err != nil {
		fmt.Println("Exchange was not created: ", err)
	}
	gamelogic.PrintServerHelp()

	for {
		textInput := gamelogic.GetInput()
		fmt.Printf("TextInput is: %s\n", textInput)
		if len(textInput) == 0 {
			fmt.Println("Empty slice")
			continue
		} else if textInput[0] == routing.PauseKey {
            fmt.Println("Pause should be posted")
			err = pubsub.PublishJSON(myC, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
			if err != nil {
				fmt.Println("Publish has not been successful", err)
			}
		} else if textInput[0] == "resume" {
            fmt.Println("Resume should be posted")
			err = pubsub.PublishJSON(myC, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
			if err != nil {
				fmt.Println("Publish has not been successful", err)
			}
        } else if textInput[0] == "quit" {
            fmt.Println("Quit message received")
            break
        } else {
            fmt.Printf("Command not recognized: %s\n", textInput[0])
        }
	}
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
}
