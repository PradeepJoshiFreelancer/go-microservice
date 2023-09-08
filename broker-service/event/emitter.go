package event

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Emitter struct {
	connection *amqp.Connection
}

func (e *Emitter) setup() error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	return declareExchange(channel)
}
func (e *Emitter) Push(event, severity string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	log.Println("inside Emitter")
	log.Println(event)
	channel, err := e.connection.Channel()
	if err != nil {
		log.Println(err)
		return err
	}
	defer channel.Close()

	log.Println("Pushing message... ")
	err = channel.PublishWithContext(
		ctx,
		"log_topic",
		severity,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(event),
		},
	)
	if err != nil {
		log.Println(err)

		return err
	}
	log.Println("Message Pushed")

	return nil
}

func NewEmiter(conn *amqp.Connection) (Emitter, error) {
	emitter := Emitter{
		connection: conn,
	}
	err := emitter.setup()
	if err != nil {
		log.Println(err)
		return Emitter{}, err
	}
	return emitter, nil
}
