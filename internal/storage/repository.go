package storage

import (
	"FilmCounter/internal/domain"
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"
)

type Store struct {
	db     *sql.DB
	logger *log.Logger
}

func New(dbURL string, logger *log.Logger) (*Store, error) {
	db, err := openDB(dbURL)
	logger.Println("connected to database")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, logger: logger}
	err = s.prepareDB()
	if err != nil {
		_ = s.db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() {
	err := s.db.Close()
	if err != nil {
		s.logger.Printf("failed to close database connection: %v", err)
	} else {
		s.logger.Println("closed database connection")
	}
}

func openDB(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, fmt.Errorf("error opening db connection: %w", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging db: %w", err)
	}
	return db, nil
}

func (s *Store) prepareDB() error {
	_, err := s.db.Exec("create table if not exists films (id serial primary key, name text, original_name text, year integer, api_id integer unique, countries text[], directors text[])")
	if err != nil {
		return fmt.Errorf("error preparing films table: %w", err)
	}
	s.logger.Println("table films ready")
	return nil
}

func (s *Store) AddFilm(film domain.Film) error {
	s.logger.Printf("inserting film: %s (%d)", film.Name, film.Year)
	_, err := s.db.Exec("insert into films (name, original_name, year, api_id, countries, directors) values ($1, $2, $3, $4, $5, $6)", film.Name, film.OriginalName, film.Year, film.ID, pq.Array(film.Countries), pq.Array(film.Directors))
	if err != nil {
		return fmt.Errorf("error inserting films: %w", err)
	}
	s.logger.Println("film successfully inserted")
	return nil
}

func (s *Store) DeleteFilmByID(id int) error {
	s.logger.Printf("deleting film with id: %d", id)
	_, err := s.db.Exec("delete from films where api_id = $1", id)
	if err != nil {
		return fmt.Errorf("error deleting film: %w", err)
	}
	s.logger.Println("film successfully deleted")
	return nil
}

func (s *Store) AllFilms() ([]domain.Film, error) {
	rows, err := s.db.Query("select api_id, name, original_name, year, countries, directors from films")
	if err != nil {
		return nil, fmt.Errorf("error selecting films: %w", err)
	}
	defer rows.Close()
	var films []domain.Film
	for rows.Next() {
		var film domain.Film
		err = rows.Scan(&film.ID, &film.Name, &film.OriginalName, &film.Year, pq.Array(&film.Countries), pq.Array(&film.Directors))
		if err != nil {
			return nil, fmt.Errorf("error scanning rows: %w", err)
		}
		films = append(films, film)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error scanning rows: %w", err)
	}
	s.logger.Printf("fetched %d films", len(films))
	return films, nil
}

func (s *Store) AllFilmsByDirectors() (map[string][]domain.Film, error) {
	rows, err := s.db.Query("select api_id, name, original_name, year, countries, directors from films order by directors")
	if err != nil {
		return nil, fmt.Errorf("error selecting films: %w", err)
	}
	defer rows.Close()
	var m = make(map[string][]domain.Film)
	for rows.Next() {
		var film domain.Film
		err = rows.Scan(&film.ID, &film.Name, &film.OriginalName, &film.Year, pq.Array(&film.Countries), pq.Array(&film.Directors))
		if err != nil {
			return nil, fmt.Errorf("error scanning rows: %w", err)
		}
		directors := strings.Join(film.Directors, ", ")
		films, ok := m[directors]
		if !ok {
			m[directors] = make([]domain.Film, 0)
			films = m[directors]
		}
		m[directors] = append(films, film)
	}
	return m, nil
}
