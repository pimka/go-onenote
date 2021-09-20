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
	Router   *mux.Router
	NH       db.NoteHandler
	DBPurger *db.NotePurger
	VPurger  *VisitorsPurger
}

func (s *Server) routes(vl *VLimiter) {
	noteRouter := s.Router.PathPrefix("/note/").Subrouter()
	noteRouter.Handle("/", Limiter(s.ListNotes(), vl)).Methods("GET")
	noteRouter.Handle("/", Limiter(s.AddNote(), vl)).Methods("POST")
	noteRouter.Handle("/{uid}", Limiter(SimpleAuth(s.GetNote()), vl)).Methods("GET")
	noteRouter.Handle("/{uid}", Limiter(s.UpdateNote(), vl)).Methods("PATCH")
	noteRouter.Handle("/{uid}", Limiter(SimpleAuth(s.DeleteNote()), vl)).Methods("DELETE")
	noteRouter.Handle("/api/", Limiter(s.PopNote(), vl)).Methods("DELETE")
	noteRouter.Handle("/api/", Limiter(s.PeekNote(), vl)).Methods("GET")
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
	s.routes(&s.VPurger.limiter)
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

	s.DBPurger.Purge(context.Background())
	s.VPurger.Purge()

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
	s.DBPurger.Stop() <- struct{}{}
	<-s.DBPurger.Done()
	s.VPurger.Stop() <- struct{}{}
	<-s.VPurger.Done()
}
