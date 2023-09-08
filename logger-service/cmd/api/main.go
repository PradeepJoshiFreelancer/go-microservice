package main

import (
	"context"
	"fmt"
	"log"
	"log-service/data"
	"net"
	"net/http"
	"net/rpc"
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

	//Register and start RPC server
	go app.rpcListen()
	err = rpc.Register(new(RPCServer))

	//start web server
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

func (app *Config) rpcListen() error {
	log.Println("starting the RPC Server at port:", rpcPort)
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))
	if err != nil {
		return err
	}
	defer listen.Close()

	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(rpcConn)
	}
	return nil
}

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
