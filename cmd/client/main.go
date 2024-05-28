package main

import (
	"fmt"
	"pubsub/internal/gamelogic"

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
}
