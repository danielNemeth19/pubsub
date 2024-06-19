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


func handlerPause(gs *gamelogic.GameState) func(ps routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Printf("> ")
		gs.HandlePause(ps)
        return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState) func(am gamelogic.ArmyMove) pubsub.AckType {
	return func(am gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Printf("> ")
        outcome := gs.HandleMove(am)
        if outcome == gamelogic.MoveOutComeSafe || outcome == gamelogic.MoveOutcomeMakeWar {
            return pubsub.Ack
        } 
		return pubsub.NackDiscard
	}
}

func runClientLoop(chn *amqp.Channel, ng *gamelogic.GameState) {
	for {
		textInput := gamelogic.GetInput()
		fmt.Printf("TextInput is: %s\n", textInput)
		if len(textInput) == 0 {
			fmt.Println("Empty command")
			continue
		}
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
			err = pubsub.PublishJSON(chn, routing.ExchangePerilTopic, "army_moves"+"."+ng.GetUsername(), move)
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

	chn, _, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, routing.GameLogSlug, "game_logs.*", pubsub.Durable)
	if err != nil {
		panic("Error declaring and binding channel")
	}
	defer chn.Close()

	newGame := gamelogic.NewGameState(username)
    err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, "army_move"+"."+username, "army_moves.*", pubsub.Transient, handlerMove(newGame))
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, pubsub.Transient, handlerPause(newGame))
	runClientLoop(chn, newGame)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	fmt.Println("Client running... Press Ctr-C to exit.")
	<-signalChan
	fmt.Println("Received signal, exiting...")
}
