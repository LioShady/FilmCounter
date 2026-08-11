package storage

import (
	"FilmCounter/internal/domain"
	"database/sql"
	"io"
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/lib/pq"
)

func newTestLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}

func setupMockDB(t *testing.T) (*Store, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	store := &Store{
		db:     db,
		logger: newTestLogger(),
	}

	t.Cleanup(func() {
		store.Close()
	})

	return store, mock
}

func TestNew_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("create table if not exists films").
		WillReturnResult(sqlmock.NewResult(0, 0))

	store := &Store{
		db:     db,
		logger: newTestLogger(),
	}

	err = store.prepareDB()
	if err != nil {
		t.Fatalf("prepareDB returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestNew_InvalidDBURL(t *testing.T) {
	store, err := New("invalid-url", newTestLogger())
	if err == nil {
		t.Fatal("New with invalid URL returned nil error, want error")
	}
	if store != nil {
		t.Fatal("New with invalid URL returned store, want nil")
	}
}

func TestStore_AddFilm_Success(t *testing.T) {
	store, mock := setupMockDB(t)

	film := domain.Film{
		ID:           123,
		Name:         "The Matrix",
		OriginalName: "The Matrix",
		Year:         1999,
		Countries:    []string{"USA", "Australia"},
		Directors:    []string{"Lana Wachowski", "Lilly Wachowski"},
	}

	mock.ExpectExec("insert into films").
		WithArgs(film.Name, film.OriginalName, film.Year, film.ID, pq.Array(film.Countries), pq.Array(film.Directors)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := store.AddFilm(film)
	if err != nil {
		t.Fatalf("AddFilm returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestStore_AddFilm_Duplicate(t *testing.T) {
	store, mock := setupMockDB(t)

	film := domain.Film{
		ID:           123,
		Name:         "The Matrix",
		OriginalName: "The Matrix",
		Year:         1999,
		Countries:    []string{"USA"},
		Directors:    []string{"Lana Wachowski"},
	}

	mock.ExpectExec("insert into films").
		WithArgs(film.Name, film.OriginalName, film.Year, film.ID, pq.Array(film.Countries), pq.Array(film.Directors)).
		WillReturnError(sql.ErrConnDone)

	err := store.AddFilm(film)
	if err == nil {
		t.Fatal("AddFilm with duplicate returned nil error, want error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestStore_AddFilm_DBError(t *testing.T) {
	store, mock := setupMockDB(t)

	film := domain.Film{
		ID:   123,
		Name: "The Matrix",
		Year: 1999,
	}

	mock.ExpectExec("insert into films").
		WithArgs(film.Name, film.OriginalName, film.Year, film.ID, pq.Array(film.Countries), pq.Array(film.Directors)).
		WillReturnError(sql.ErrConnDone)

	err := store.AddFilm(film)
	if err == nil {
		t.Fatal("AddFilm with DB error returned nil error, want error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestStore_DeleteFilmByID_Success(t *testing.T) {
	store, mock := setupMockDB(t)

	filmID := 123

	mock.ExpectExec("delete from films where api_id = \\$1").
		WithArgs(filmID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := store.DeleteFilmByID(filmID)
	if err != nil {
		t.Fatalf("DeleteFilmByID returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestStore_DeleteFilmByID_NotFound(t *testing.T) {
	store, mock := setupMockDB(t)

	filmID := 99999

	mock.ExpectExec("delete from films where api_id = \\$1").
		WithArgs(filmID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := store.DeleteFilmByID(filmID)
	if err != nil {
		t.Fatalf("DeleteFilmByID returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestStore_DeleteFilmByID_DBError(t *testing.T) {
	store, mock := setupMockDB(t)

	filmID := 123

	mock.ExpectExec("delete from films where api_id = \\$1").
		WithArgs(filmID).
		WillReturnError(sql.ErrConnDone)

	err := store.DeleteFilmByID(filmID)
	if err == nil {
		t.Fatal("DeleteFilmByID with DB error returned nil error, want error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestStore_AllFilms_Empty(t *testing.T) {
	store, mock := setupMockDB(t)

	rows := sqlmock.NewRows([]string{"api_id", "name", "original_name", "year", "countries", "directors"})

	mock.ExpectQuery("select api_id, name, original_name, year, countries, directors from films").
		WillReturnRows(rows)

	films, err := store.AllFilms()
	if err != nil {
		t.Fatalf("AllFilms returned error: %v", err)
	}

	if len(films) != 0 {
		t.Fatalf("expected 0 films, got %d", len(films))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestStore_AllFilms_Success(t *testing.T) {
	store, mock := setupMockDB(t)

	expectedFilms := []domain.Film{
		{
			ID:           1,
			Name:         "Film 1",
			OriginalName: "Original 1",
			Year:         2020,
			Countries:    []string{"USA"},
			Directors:    []string{"Director 1"},
		},
		{
			ID:           2,
			Name:         "Film 2",
			OriginalName: "Original 2",
			Year:         2021,
			Countries:    []string{"UK", "France"},
			Directors:    []string{"Director 2", "Director 3"},
		},
	}

	rows := sqlmock.NewRows([]string{"api_id", "name", "original_name", "year", "countries", "directors"}).
		AddRow(1, "Film 1", "Original 1", 2020, pq.Array([]string{"USA"}), pq.Array([]string{"Director 1"})).
		AddRow(2, "Film 2", "Original 2", 2021, pq.Array([]string{"UK", "France"}), pq.Array([]string{"Director 2", "Director 3"}))

	mock.ExpectQuery("select api_id, name, original_name, year, countries, directors from films").
		WillReturnRows(rows)

	films, err := store.AllFilms()
	if err != nil {
		t.Fatalf("AllFilms returned error: %v", err)
	}

	if len(films) != len(expectedFilms) {
		t.Fatalf("expected %d films, got %d", len(expectedFilms), len(films))
	}

	for i, expected := range expectedFilms {
		if diff := cmp.Diff(expected, films[i]); diff != "" {
			t.Fatalf("film %d mismatch (-want +got):\n%s", i, diff)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestStore_AllFilms_QueryError(t *testing.T) {
	store, mock := setupMockDB(t)

	mock.ExpectQuery("select api_id, name, original_name, year, countries, directors from films").
		WillReturnError(sql.ErrConnDone)

	films, err := store.AllFilms()
	if err == nil {
		t.Fatal("AllFilms with query error returned nil error, want error")
	}
	if films != nil {
		t.Fatalf("expected nil films on error, got %+v", films)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestStore_AllFilms_ScanError(t *testing.T) {
	store, mock := setupMockDB(t)

	rows := sqlmock.NewRows([]string{"api_id", "name", "original_name", "year", "countries", "directors"}).
		AddRow("invalid_type", "Film", "Original", 2020, pq.Array([]string{"USA"}), pq.Array([]string{"Director"}))

	mock.ExpectQuery("select api_id, name, original_name, year, countries, directors from films").
		WillReturnRows(rows)

	films, err := store.AllFilms()
	if err == nil {
		t.Fatal("AllFilms with scan error returned nil error, want error")
	}
	if films != nil {
		t.Fatalf("expected nil films on error, got %+v", films)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestStore_AllFilmsByDirectors_Success(t *testing.T) {
	store, mock := setupMockDB(t)

	expectedFilms := []domain.Film{
		{
			ID:           1,
			Name:         "Film 1",
			OriginalName: "Original 1",
			Year:         2020,
			Countries:    []string{"USA"},
			Directors:    []string{"Director A", "Director B"},
		},
		{
			ID:           2,
			Name:         "Film 2",
			OriginalName: "Original 2",
			Year:         2021,
			Countries:    []string{"UK"},
			Directors:    []string{"Director A"},
		},
		{
			ID:           3,
			Name:         "Film 3",
			OriginalName: "Original 3",
			Year:         2022,
			Countries:    []string{"France"},
			Directors:    []string{"Director C"},
		},
	}

	rows := sqlmock.NewRows([]string{"api_id", "name", "original_name", "year", "countries", "directors"}).
		AddRow(1, "Film 1", "Original 1", 2020, pq.Array([]string{"USA"}), pq.Array([]string{"Director A", "Director B"})).
		AddRow(2, "Film 2", "Original 2", 2021, pq.Array([]string{"UK"}), pq.Array([]string{"Director A"})).
		AddRow(3, "Film 3", "Original 3", 2022, pq.Array([]string{"France"}), pq.Array([]string{"Director C"}))

	mock.ExpectQuery("select api_id, name, original_name, year, countries, directors from films order by directors").
		WillReturnRows(rows)

	grouped, err := store.AllFilmsByDirectors()
	if err != nil {
		t.Fatalf("AllFilmsByDirectors returned error: %v", err)
	}

	if len(grouped) != 3 {
		t.Fatalf("expected 3 director groups, got %d", len(grouped))
	}

	expectedGroups := map[string][]domain.Film{
		"Director A, Director B": {expectedFilms[0]},
		"Director A":             {expectedFilms[1]},
		"Director C":             {expectedFilms[2]},
	}

	for key, expectedFilmsGroup := range expectedGroups {
		films, ok := grouped[key]
		if !ok {
			t.Fatalf("director group %q not found", key)
		}
		if len(films) != len(expectedFilmsGroup) {
			t.Fatalf("director group %q: expected %d films, got %d", key, len(expectedFilmsGroup), len(films))
		}
		for i, expected := range expectedFilmsGroup {
			if diff := cmp.Diff(expected, films[i]); diff != "" {
				t.Fatalf("film in group %q mismatch (-want +got):\n%s", key, diff)
			}
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestStore_AllFilmsByDirectors_Empty(t *testing.T) {
	store, mock := setupMockDB(t)

	rows := sqlmock.NewRows([]string{"api_id", "name", "original_name", "year", "countries", "directors"})

	mock.ExpectQuery("select api_id, name, original_name, year, countries, directors from films order by directors").
		WillReturnRows(rows)

	grouped, err := store.AllFilmsByDirectors()
	if err != nil {
		t.Fatalf("AllFilmsByDirectors returned error: %v", err)
	}

	if len(grouped) != 0 {
		t.Fatalf("expected 0 groups, got %d", len(grouped))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestStore_AllFilmsByDirectors_WithEmptyDirectors(t *testing.T) {
	store, mock := setupMockDB(t)

	film := domain.Film{
		ID:           1,
		Name:         "Film",
		OriginalName: "Original",
		Year:         2020,
		Countries:    []string{"USA"},
		Directors:    []string{},
	}

	rows := sqlmock.NewRows([]string{"api_id", "name", "original_name", "year", "countries", "directors"}).
		AddRow(1, "Film", "Original", 2020, pq.Array([]string{"USA"}), pq.Array([]string{}))

	mock.ExpectQuery("select api_id, name, original_name, year, countries, directors from films order by directors").
		WillReturnRows(rows)

	grouped, err := store.AllFilmsByDirectors()
	if err != nil {
		t.Fatalf("AllFilmsByDirectors returned error: %v", err)
	}

	if len(grouped) != 1 {
		t.Fatalf("expected 1 group, got %d", len(grouped))
	}

	films, ok := grouped[""]
	if !ok {
		t.Fatal("group with empty directors not found")
	}
	if len(films) != 1 {
		t.Fatalf("expected 1 film in group, got %d", len(films))
	}

	if diff := cmp.Diff(film, films[0]); diff != "" {
		t.Fatalf("film mismatch (-want +got):\n%s", diff)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestStore_AllFilmsByDirectors_MultipleFilmsSameDirector(t *testing.T) {
	store, mock := setupMockDB(t)

	expectedFilms := []domain.Film{
		{
			ID:           1,
			Name:         "Film 1",
			OriginalName: "Original 1",
			Year:         2020,
			Countries:    []string{"USA"},
			Directors:    []string{"Director A"},
		},
		{
			ID:           2,
			Name:         "Film 2",
			OriginalName: "Original 2",
			Year:         2021,
			Countries:    []string{"UK"},
			Directors:    []string{"Director A"},
		},
		{
			ID:           3,
			Name:         "Film 3",
			OriginalName: "Original 3",
			Year:         2022,
			Countries:    []string{"France"},
			Directors:    []string{"Director A"},
		},
	}

	rows := sqlmock.NewRows([]string{"api_id", "name", "original_name", "year", "countries", "directors"}).
		AddRow(1, "Film 1", "Original 1", 2020, pq.Array([]string{"USA"}), pq.Array([]string{"Director A"})).
		AddRow(2, "Film 2", "Original 2", 2021, pq.Array([]string{"UK"}), pq.Array([]string{"Director A"})).
		AddRow(3, "Film 3", "Original 3", 2022, pq.Array([]string{"France"}), pq.Array([]string{"Director A"}))

	mock.ExpectQuery("select api_id, name, original_name, year, countries, directors from films order by directors").
		WillReturnRows(rows)

	grouped, err := store.AllFilmsByDirectors()
	if err != nil {
		t.Fatalf("AllFilmsByDirectors returned error: %v", err)
	}

	if len(grouped) != 1 {
		t.Fatalf("expected 1 director group, got %d", len(grouped))
	}

	films, ok := grouped["Director A"]
	if !ok {
		t.Fatal("director group 'Director A' not found")
	}
	if len(films) != 3 {
		t.Fatalf("expected 3 films in group, got %d", len(films))
	}

	for i, expected := range expectedFilms {
		if diff := cmp.Diff(expected, films[i]); diff != "" {
			t.Fatalf("film %d mismatch (-want +got):\n%s", i, diff)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestStore_AllFilmsByDirectors_QueryError(t *testing.T) {
	store, mock := setupMockDB(t)

	mock.ExpectQuery("select api_id, name, original_name, year, countries, directors from films order by directors").
		WillReturnError(sql.ErrConnDone)

	grouped, err := store.AllFilmsByDirectors()
	if err == nil {
		t.Fatal("AllFilmsByDirectors with query error returned nil error, want error")
	}
	if grouped != nil {
		t.Fatalf("expected nil grouped on error, got %+v", grouped)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestStore_AllFilmsByDirectors_ScanError(t *testing.T) {
	store, mock := setupMockDB(t)

	rows := sqlmock.NewRows([]string{"api_id", "name", "original_name", "year", "countries", "directors"}).
		AddRow("invalid_type", "Film", "Original", 2020, pq.Array([]string{"USA"}), pq.Array([]string{"Director"}))

	mock.ExpectQuery("select api_id, name, original_name, year, countries, directors from films order by directors").
		WillReturnRows(rows)

	grouped, err := store.AllFilmsByDirectors()
	if err == nil {
		t.Fatal("AllFilmsByDirectors with scan error returned nil error, want error")
	}
	if grouped != nil {
		t.Fatalf("expected nil grouped on error, got %+v", grouped)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestStore_Close(t *testing.T) {
	store, mock := setupMockDB(t)

	mock.ExpectClose()

	store.Close()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestStore_PrepareDB_Success(t *testing.T) {
	store, mock := setupMockDB(t)

	mock.ExpectExec("create table if not exists films").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := store.prepareDB()
	if err != nil {
		t.Fatalf("prepareDB returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestStore_PrepareDB_Error(t *testing.T) {
	store, mock := setupMockDB(t)

	mock.ExpectExec("create table if not exists films").
		WillReturnError(sql.ErrConnDone)

	err := store.prepareDB()
	if err == nil {
		t.Fatal("prepareDB with error returned nil error, want error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}
