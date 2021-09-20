package main

import (
	"github.com/caarlos0/env"
	"github.com/gorilla/mux"
	"github.com/pimka/go-onenote/db"
	"github.com/pimka/go-onenote/server"
	"log"
	"time"
)

const DBURL = "postgres://puser:puser123@localhost:5432/postgres?sslmode=disable"

func main() {
	conf := server.Config{}
	if err := env.Parse(&conf); err != nil {
		log.Fatalf("could not parse env vars for config: %v", err)
	}

	database, err := db.Connect(DBURL)
	if err != nil {
		log.Fatal(err)
	}

	nh := db.NewNoteDB(database.Conn)
	vl := server.VLimiter{Visitors: make(map[string]*server.Visitor)}

	service := &server.Server{
		VPurger:  server.NewPurger(vl, time.Minute, 5),
		DBPurger: db.NewPurger(nh, time.Minute, 5),
		Router:   mux.NewRouter(),
		NH:       nh,
	}
	service.Start(conf)
	defer service.Stop()
}
