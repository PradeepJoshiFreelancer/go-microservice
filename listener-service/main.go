package main

import (
	"fmt"
	"listener/event"
	"log"
	"math"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	//connet to Rabbit MQ
	rabbitMQConn, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitMQConn.Close()

	//Listen to any messages
	log.Println("Lisening and consuming messages from MQ")

	//create consumer
	consumer, err := event.NewConsumer(rabbitMQConn)
	if err != nil {
		log.Panic(err)
	}
	//watch for any new message
	err = consumer.Listen([]string{"log.INFO", "log.ERROR", "log.WARNING"})
	if err != nil {
		log.Println(err)
	}
}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	//dont continue unless rabbitmq is ready

	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			fmt.Println("rabbit mq not ready")
			counts++
		} else {
			log.Println("Connected tp rabbit MQ")
			connection = c
			break
		}
		if counts > 5 {
			log.Println("Unable to connect to RabbitMQ")
			log.Println(err)
			return nil, err
		}
		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off...")
		time.Sleep(backOff)
		continue
	}

	return connection, nil

}
