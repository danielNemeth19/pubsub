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
	defer conn.Close()
	if err != nil {
		panic("Error establishing connection")
	}
	username, _ := gamelogic.ClientWelcome()
	fmt.Printf("username is: %s\n", username)
    chn, _, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, pubsub.Transient)
    defer func() {
        fmt.Println("Closing channel and connection")
        err := chn.Close()
        if err != nil {
            fmt.Println("Closing channel failed: ", err)

        }
        err = conn.Close()
        if err != nil {
            fmt.Println("Closing connection failed: ", err)

        }
    }()
    signalChan := make(chan os.Signal, 1)
    signal.Notify(signalChan, os.Interrupt )
    <- signalChan
}
