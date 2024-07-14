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

const (
	Pause  = "pause"
	Resume = "resume"
	Quit   = "quit"
)

func runLoop(ch *amqp.Channel) {
	for {
		textInput := gamelogic.GetInput()
		fmt.Printf("TextInput is: %s\n", textInput)
		if len(textInput) == 0 {
			fmt.Println("Empty command")
			continue
		}
		command := textInput[0]
		if command == Quit {
			break
		}
		switch command {
		case Pause:
			fmt.Println("Pause should be posted")
			err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
			if err != nil {
				fmt.Println("Publish has not been successful", err)
			}
		case Resume:
			fmt.Println("Resume should be posted")
			err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
			if err != nil {
				fmt.Println("Publish has not been successful", err)
			}
		default:
			fmt.Printf("Command not recognized: %s\n", textInput[0])
		}
	}
}

func handlerLog() func(gl routing.GameLog) pubsub.AckType {
	return func(gl routing.GameLog) pubsub.AckType {
		defer fmt.Printf("> ")
		err := gamelogic.WriteLog(gl)
		if err != nil {
			fmt.Printf("Saving the log failed: %v\n", err)
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}

func setUpExchanges(ch *amqp.Channel) {
	err := pubsub.CreateExchange(ch, routing.ExchangePerilDirect, pubsub.Direct, pubsub.Durable)
	if err != nil {
		fmt.Println("Exchange was not created: ", err)
	}
	err = pubsub.CreateExchange(ch, routing.ExchangePerilTopic, pubsub.Topic, pubsub.Durable)
	if err != nil {
		fmt.Println("Exchange was not created: ", err)
	}
	err = pubsub.CreateExchange(ch, routing.ExchangePerilDlx, pubsub.Fanout, pubsub.Durable)
	if err != nil {
		fmt.Println("Exchange was not created: ", err)
	}
}

func setUpDeadLetter(conn *amqp.Connection) {
	deadLetterTable := pubsub.GetDeadLetterConfig()
	chn, _, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilDlx, routing.PerilDlq, "", pubsub.Durable, deadLetterTable)
	if err != nil {
		panic("Error declaring and binding channel")
	}
	defer chn.Close()
}

func setUpGameLogs(conn *amqp.Connection) {
	err := pubsub.SubscribeGob[routing.GameLog](conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		"game_logs.*",
		pubsub.Durable,
		pubsub.HandlerWithoutConn[routing.GameLog](handlerLog()),
	)
	if err != nil {
		panic("Error declaring and binding channel")
	}
}

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
	setUpExchanges(myC)
	setUpDeadLetter(conn)
	setUpGameLogs(conn)
	gamelogic.PrintServerHelp()
	runLoop(myC)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
}
