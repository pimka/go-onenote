package server

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pimka/go-onenote/db"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Config struct {
	Port             int      `env:"PORT" envDefault:"8080"`
	AllowedOrigins   []string `env:"ALLOWED_ORIGINS" envSeparator:"," envDefault:"http://localhost:8000"`
	AllowedMethods   []string `env:"ALLOWED_METHODS" envSeparator:"," envDefault:"GET,POST,PATCH,DELETE"`
	AllowedHeaders   []string `env:"ALLOWED_HEADERS" envSeparator:"," envDefault:"Origin,X-Requested-With,Content-Type,Accept,Access-Control-Allow-Origin,Authorization"`
	AllowCredentials bool     `env:"ALLOWED_CREDENTIALS" envDefault:"true"`
}

type Server struct {
	Router *mux.Router
	NH     db.NoteHandler
	Purger *db.Purger
}

func (s *Server) routes() {
	noteRouter := s.Router.PathPrefix("/note/").Subrouter()
	noteRouter.HandleFunc("/", s.listNotes()).Methods("GET")
	noteRouter.HandleFunc("/", s.addNote()).Methods("POST")
	noteRouter.HandleFunc("/{uid}", s.getNote()).Methods("GET")
	noteRouter.HandleFunc("/{uid}", s.updateNote()).Methods("PATCH")
	noteRouter.HandleFunc("/{uid}", s.deleteNote()).Methods("DELETE")
	noteRouter.HandleFunc("/api/", s.popNote()).Methods("DELETE")
	noteRouter.HandleFunc("/api/", s.peekNote()).Methods("GET")
}

func setContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(writer, request)
	})
}

func (s *Server) Start(c Config) {
	cors := cors.New(cors.Options{
		AllowedOrigins:   c.AllowedOrigins,
		AllowedMethods:   c.AllowedMethods,
		AllowedHeaders:   c.AllowedHeaders,
		AllowCredentials: c.AllowCredentials,
	})

	s.Router.Use(setContentType)
	s.routes()
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", c.Port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      cors.Handler(s.Router),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	s.Purger.Purge(context.Background())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("HTTP server is shutting down")
	os.Exit(0)
}

func (s *Server) Stop() {
	s.Purger.Stop() <- struct{}{}
	<-s.Purger.Done()
}
