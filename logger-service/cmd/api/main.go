package main

import (
	"context"
	"fmt"
	"log"
	"log-service/data"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort  = "80"
	rpcPort  = "5001"
	mongoURL = "mongodb://mongo:27017"
	gRpcPort = "50001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	mongoClient, err := connectToMongo()

	if err != nil {
		log.Panic(err)
	}
	client = mongoClient
	//create a context to disconnect
	cntx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	//disconnect
	defer func() {
		if err = client.Disconnect(cntx); err != nil {
			panic(err)
		}

	}()

	app := Config{
		Models: data.New(client),
	}

	//start web server
	// go app.serve()
	log.Println("Starting server at port", webPort)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}
	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}

}

// func (app *Config) serve() {
// 	srv := &http.Server{
// 		Addr:    fmt.Sprintf(":%s", webPort),
// 		Handler: app.routes(),
// 	}
// 	err := srv.ListenAndServe()
// 	if err != nil {
// 		log.Panic(err)
// 	}
// }

func connectToMongo() (*mongo.Client, error) {
	//create cpnnection options
	clientoptions := options.Client().ApplyURI(mongoURL)
	clientoptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})
	// connet to Mongo
	c, err := mongo.Connect(context.TODO(), clientoptions)
	if err != nil {
		log.Println("Error connecting: ", err)
		return nil, err
	}
	log.Println("Connected to Mongo!")
	return c, nil
}
