package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "80"

type Config struct {
	Rabbit *amqp.Connection
}

func main() {
	//connet to Rabbit MQ
	rabbitMQConn, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitMQConn.Close()
	app := Config{
		Rabbit: rabbitMQConn,
	}

	log.Printf("Starting the broker service at port:%s\n", webPort)

	//define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}
	//start the server
	err = srv.ListenAndServe()
	if err != nil {
		log.Panic("Unable to start server", err)
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
