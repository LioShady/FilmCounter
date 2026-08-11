package main

import (
	"FilmCounter/internal/config"
	"FilmCounter/internal/pkapi"
	"FilmCounter/internal/storage"
	webapp "FilmCounter/internal/web"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("no .env file found")
	}

	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	logger := log.New(file, "APP: ", log.Ldate|log.Ltime|log.Lshortfile)

	cfg, err := config.New()
	if err != nil {
		log.Print(err)
		logger.Fatal(err)
	}
	DBURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.DBUSER, cfg.DBPASS, cfg.DBHOST, cfg.DBPORT, cfg.DBNAME)

	store, err := storage.New(DBURL, logger)
	if err != nil {
		log.Print(err)
		logger.Fatal(err)
	}
	defer store.Close()

	pkClient := pkapi.New(cfg.PoiskKinoApiKey, logger)

	server := webapp.NewServer(store, logger, pkClient)

	logger.Println("starting web server on :8080")

	if err := http.ListenAndServe(":8080", server.Routes()); err != nil {
		log.Print(err)
		logger.Fatal(err)
	}
}
