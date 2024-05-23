package main

import (
    amqp "github.com/rabbitmq/amqp091-go"
    "fmt"
)

func main() {
	fmt.Println("Starting Peril server...")
    connStr := "amqp://guest:guest@localhost:5672/"
    _, err :=  amqp.Dial(connStr)
    if err != nil {
        panic("jaaj")
    }
	fmt.Println("Connected")
}
