package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"pubsub/internal/gamelogic"
	"pubsub/internal/pubsub"
	"pubsub/internal/routing"
	"strconv"
	"time"

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

func handlerMove(gs *gamelogic.GameState) func(am gamelogic.ArmyMove, conn *amqp.Connection) pubsub.AckType {
	return func(am gamelogic.ArmyMove, conn *amqp.Connection) pubsub.AckType {
		defer fmt.Printf("> ")
		outcome := gs.HandleMove(am)
		if outcome == gamelogic.MoveOutcomeMakeWar {
			rw := gamelogic.RecognitionOfWar{Attacker: am.Player, Defender: gs.GetPlayerSnap()}
			log.Printf("Attacker: %s -- defender: %s\n", rw.Attacker.Username, rw.Defender.Username)
			chn, _ := conn.Channel()
			key := routing.WarRecognitionsPrefix + "." + gs.GetUsername()
			err := pubsub.PublishJSON(chn, routing.ExchangePerilTopic, key, rw)
			if err != nil {
				return pubsub.NackRequeue
			}
		}
		return pubsub.Ack
	}
}

func handlerWar(gs *gamelogic.GameState) func(rw gamelogic.RecognitionOfWar, conn *amqp.Connection) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar, conn *amqp.Connection) pubsub.AckType {
		defer fmt.Printf("> ")
		outcome, winner, loser := gs.HandleWar(rw)
		log.Printf("War handler of %s -- rw: %s\n", gs.Player.Username, rw.Attacker.Username)
		if outcome == gamelogic.WarOutcomeNotInvolved {
			log.Printf("Outcome: not involved (%s) (%s) -> message nack requeued\n", winner, loser)
			return pubsub.NackRequeue
		}
		if outcome == gamelogic.WarOutcomeNoUnits {
			log.Printf("Outcome: no units (%s) (%s) -> message nack discarded\n", winner, loser)
			return pubsub.NackDiscard
		}

		chn, _ := conn.Channel()
		key := routing.GameLogSlug + "." + rw.Attacker.Username

		if outcome == gamelogic.WarOutcomeOpponentWon || outcome == gamelogic.WarOutcomeYouWon {
			message := fmt.Sprintf("%s won a war against %s\n", winner, loser)
			gameLogMessage := routing.GameLog{CurrentTime: time.Now(), Message: message, Username: gs.Player.Username}
			err := pubsub.PublishGob(chn, routing.ExchangePerilTopic, key, gameLogMessage)
			if err != nil {
				return pubsub.NackRequeue
			}
			log.Printf("Outcome %d: (%s) (%s)", outcome, winner, loser)
			return pubsub.Ack
		}
		if outcome == gamelogic.WarOutcomeDraw {
			message := fmt.Sprintf("A war between %s and %s resulted in a draw\n", winner, loser)
			gameLogMessage := routing.GameLog{CurrentTime: time.Now(), Message: message, Username: gs.Player.Username}
			err := pubsub.PublishGob(chn, routing.ExchangePerilTopic, key, gameLogMessage)
			if err != nil {
				return pubsub.NackRequeue
			}
			log.Printf("Outcome: Draw (%s) (%s) -> message acked\n", winner, loser)
			return pubsub.Ack
		}
		log.Printf("Error happened: (%s) (%s) -> message nack discarded\n", winner, loser)
		return pubsub.NackRequeue
	}
}

func runClientLoop(conn *amqp.Connection, ng *gamelogic.GameState) {
	chn, err := conn.Channel()
	if err != nil {
		fmt.Errorf("Channel creation failed: %w", err)
	}
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
			if len(textInput) != 2 {
				fmt.Printf("Spamming failed: %v\n", textInput)
				continue
			}
			num, err := strconv.Atoi(textInput[1])
			if err != nil {
				fmt.Printf("Spam requires int, got: %s\n", textInput[1])
				continue
			}
			for range num {
				msg := gamelogic.GetMaliciousLog()
				key := routing.GameLogSlug + "." + ng.GetUsername()
			    gameLogMessage := routing.GameLog{CurrentTime: time.Now(), Message: msg, Username: ng.GetUsername()}
				err := pubsub.PublishGob(chn, routing.ExchangePerilTopic, key, gameLogMessage)
				if err != nil {
					fmt.Printf("Error with spamming: %s\n", err)
				}
			}
			fmt.Printf("Spamming not allowed yet! %s\n", textInput[1])
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

	newGame := gamelogic.NewGameState(username)
	err = pubsub.SubscribeJSON[gamelogic.ArmyMove](
		conn, routing.ExchangePerilTopic, "army_move"+"."+username,
		"army_moves.*", pubsub.Transient,
		pubsub.HandlerWithConn[gamelogic.ArmyMove](handlerMove(newGame)),
	)
	err = pubsub.SubscribeJSON[routing.PlayingState](
		conn, routing.ExchangePerilDirect, routing.PauseKey+"."+username,
		routing.PauseKey, pubsub.Transient,
		pubsub.HandlerWithoutConn[routing.PlayingState](handlerPause(newGame)),
	)
	err = pubsub.SubscribeJSON[gamelogic.RecognitionOfWar](
		conn, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.Durable, pubsub.HandlerWithConn[gamelogic.RecognitionOfWar](handlerWar(newGame)),
	)
	runClientLoop(conn, newGame)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	fmt.Println("Client running... Press Ctr-C to exit.")
	<-signalChan
	fmt.Println("Received signal, exiting...")
}
