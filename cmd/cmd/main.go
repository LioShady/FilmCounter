package main

import (
	"FilmCounter/internal/config"
	"FilmCounter/internal/filereader"
	"FilmCounter/internal/pkapi"
	"FilmCounter/internal/storage"
	"fmt"
	"log"
	"os"
	"sync"

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

	reader, err := filereader.New("list.txt", logger)
	if err != nil {
		logger.Fatal(err)
	}
	defer reader.Close()

	films, err := reader.Read(30)
	if err != nil {
		logger.Fatal(err)
	}

	cfg, err := config.New()
	if err != nil {
		logger.Fatal(err)
	}

	DBURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.DBUSER, cfg.DBPASS, cfg.DBHOST, cfg.DBPORT, cfg.DBNAME)

	store, err := storage.New(DBURL, logger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.Close()

	client := pkapi.New(cfg.PoiskKinoApiKey, logger)

	const workers = 5

	jobs := make(chan string)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			for filmName := range jobs {
				fullFilm, err := client.FilmByName(filmName)
				if err != nil {
					logger.Printf("worker %d: fetch %q: %v", workerID, filmName, err)
					continue
				}

				if err := store.AddFilm(fullFilm); err != nil {
					logger.Printf("worker %d: save %q: %v", workerID, fullFilm.Name, err)
					continue
				}

				logger.Printf("worker %d: saved %q", workerID, fullFilm.Name)
			}
		}(i + 1)
	}

	for _, film := range films {
		jobs <- film
	}
	close(jobs)

	wg.Wait()

}
