package main

import (
	"authentication/data"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "80"

var counts int64

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Println("Starting Authentication service")

	//Connect to the database
	con := connectToDb()
	if con == nil {
		log.Panic("Can not connect to the database!")
	}

	//setup config
	app := Config{
		DB:     con,
		Models: data.New(con),
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic("cannot start the server", err)
	}
}

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDb() *sql.DB {
	counts = 0
	dsn := os.Getenv("DSN")

	for {
		connection, err := OpenDB(dsn)
		if err != nil {
			log.Println("Postgress not yet ready ...")
			counts++
		} else {
			log.Println("connected to the database!")
			return connection
		}
		if counts > 10 {
			log.Println("Error in connecting to the database")
			return nil
		}
		log.Println("Backing off for 2 seconds....")
		time.Sleep(2 * time.Second)
		continue
	}
}
