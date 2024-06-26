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
	Spawn  = "spawn"
	Move   = "move"
	Status = "status"
	Help   = "help"
	Spam   = "spam"
	Quit   = "quit"
)

func runClientLoop(ng *gamelogic.GameState) {
	for {
		textInput := gamelogic.GetInput()
		fmt.Printf("TextInput is: %s\n", textInput)
		command := textInput[0]
		if command == Quit {
			gamelogic.PrintQuit()
			break
		}
		switch command {
		case Spawn:
			err := ng.CommandSpawn(textInput)
			if err != nil {
				fmt.Printf("Error spawning command: %s\n", err)
			}
		case Move:
			move, err := ng.CommandMove(textInput)
			if err != nil {
				fmt.Printf("Error with move: %s\n", err)
				continue
			}
			fmt.Printf("Move worked: %v\n", move)
		case Status:
			fmt.Println("Status should be presented")
			ng.CommandStatus()
		case Help:
			fmt.Println("Help should be printed")
			gamelogic.PrintClientHelp()
		case Spam:
			fmt.Println("Spamming not allowed yet!")
		default:
			fmt.Printf("Command not recognized: %s\n", textInput[0])
		}
	}
}

func main() {
	fmt.Println("Starting Peril client...")
	connStr := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connStr)
	if err != nil {
		panic("Error establishing connection")
	}
	defer conn.Close()

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		panic(err)
	}
	fmt.Printf("username is: %s\n", username)

	chn, _, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, pubsub.Transient)
	if err != nil {
		panic("Error declaring and binding channel")
	}
	chn, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, routing.GameLogSlug, "game_logs.*", pubsub.Durable)
	if err != nil {
		panic("Error declaring and binding channel")
	}
	defer chn.Close()

	newGame := gamelogic.NewGameState(username)
	runClientLoop(newGame)

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
