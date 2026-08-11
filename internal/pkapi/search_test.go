package pkapi

import (
	"FilmCounter/internal/domain"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func newTestLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}

type mockResponse struct {
	statusCode int
	body       interface{}
}

func setupTestClient(t *testing.T, mockResponses map[string]mockResponse) *Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-KEY") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		for path, mock := range mockResponses {
			if strings.Contains(r.URL.Path, path) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(mock.statusCode)

				if mock.body != nil {
					jsonBytes, err := json.Marshal(mock.body)
					if err != nil {
						t.Fatalf("failed to marshal mock response: %v", err)
					}
					w.Write(jsonBytes)
				}
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
	}))

	t.Cleanup(server.Close)

	return &Client{
		baseURL:    server.URL,
		key:        "test-api-key",
		httpClient: server.Client(),
		logger:     newTestLogger(),
	}
}

func TestFilmByName_Success(t *testing.T) {
	mockResponses := map[string]mockResponse{
		"/v1.4/movie/search": {
			statusCode: http.StatusOK,
			body: searchResponse{
				Docs: []filmDoc{
					{
						ID:           123,
						Name:         "The Matrix",
						OriginalName: "The Matrix",
						Year:         1999,
						Countries: []filmCountry{
							{CountryName: "USA"},
							{CountryName: "Australia"},
						},
					},
				},
			},
		},
		"/v1.4/movie/123": {
			statusCode: http.StatusOK,
			body: filmResponseByID{
				Persons: []filmPerson{
					{
						ID:         1,
						Name:       "Lana Wachowski",
						Profession: "director",
					},
					{
						ID:         2,
						Name:       "Lilly Wachowski",
						Profession: "director",
					},
					{
						ID:         3,
						Name:       "Keanu Reeves",
						Profession: "actor",
					},
				},
			},
		},
	}

	client := setupTestClient(t, mockResponses)

	film, err := client.FilmByName("The Matrix")
	if err != nil {
		t.Fatalf("FilmByName returned error: %v", err)
	}

	expected := domain.Film{
		ID:           123,
		Name:         "The Matrix",
		OriginalName: "The Matrix",
		Year:         1999,
		Countries:    []string{"USA", "Australia"},
		Directors:    []string{"Lana Wachowski", "Lilly Wachowski"},
	}

	if diff := cmp.Diff(expected, film); diff != "" {
		t.Fatalf("Film mismatch (-want +got):\n%s", diff)
	}
}

func TestFilmByName_NoResults(t *testing.T) {
	mockResponses := map[string]mockResponse{
		"/v1.4/movie/search": {
			statusCode: http.StatusOK,
			body: searchResponse{
				Docs: []filmDoc{},
			},
		},
	}

	client := setupTestClient(t, mockResponses)

	film, err := client.FilmByName("NonExistentFilm12345")
	if err == nil {
		t.Fatal("FilmByName returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "can't find films") {
		t.Fatalf("Expected error about not finding film, got: %v", err)
	}

	emptyFilm := domain.Film{}
	if diff := cmp.Diff(emptyFilm, film); diff != "" {
		t.Fatalf("Expected empty film, got diff: %s", diff)
	}
}

func TestFilmByName_APIError(t *testing.T) {
	mockResponses := map[string]mockResponse{
		"/v1.4/movie/search": {
			statusCode: http.StatusInternalServerError,
			body:       map[string]string{"error": "internal server error"},
		},
	}

	client := setupTestClient(t, mockResponses)

	film, err := client.FilmByName("The Matrix")
	if err == nil {
		t.Fatal("FilmByName returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "api response status 500") {
		t.Fatalf("Expected API error, got: %v", err)
	}

	emptyFilm := domain.Film{}
	if diff := cmp.Diff(emptyFilm, film); diff != "" {
		t.Fatalf("Expected empty film, got diff: %s", diff)
	}
}

func TestFilmByName_DirectorError(t *testing.T) {
	mockResponses := map[string]mockResponse{
		"/v1.4/movie/search": {
			statusCode: http.StatusOK,
			body: searchResponse{
				Docs: []filmDoc{
					{
						ID:   123,
						Name: "The Matrix",
						Year: 1999,
					},
				},
			},
		},
		"/v1.4/movie/123": {
			statusCode: http.StatusInternalServerError,
			body:       map[string]string{"error": "director fetch failed"},
		},
	}

	client := setupTestClient(t, mockResponses)

	film, err := client.FilmByName("The Matrix")
	if err == nil {
		t.Fatal("FilmByName returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "API response status 500") {
		t.Fatalf("Expected API error, got: %v", err)
	}

	emptyFilm := domain.Film{}
	if diff := cmp.Diff(emptyFilm, film); diff != "" {
		t.Fatalf("Expected empty film, got diff: %s", diff)
	}
}

func TestFilmsByName_Success(t *testing.T) {
	mockResponses := map[string]mockResponse{
		"/v1.4/movie/search": {
			statusCode: http.StatusOK,
			body: searchResponse{
				Docs: []filmDoc{
					{
						ID:           123,
						Name:         "The Matrix",
						OriginalName: "The Matrix",
						Year:         1999,
						Countries: []filmCountry{
							{CountryName: "USA"},
						},
					},
					{
						ID:           456,
						Name:         "The Matrix Reloaded",
						OriginalName: "The Matrix Reloaded",
						Year:         2003,
						Countries: []filmCountry{
							{CountryName: "USA"},
							{CountryName: "Australia"},
						},
					},
				},
			},
		},
		"/v1.4/movie/123": {
			statusCode: http.StatusOK,
			body: filmResponseByID{
				Persons: []filmPerson{
					{
						ID:         1,
						Name:       "Lana Wachowski",
						Profession: "director",
					},
				},
			},
		},
		"/v1.4/movie/456": {
			statusCode: http.StatusOK,
			body: filmResponseByID{
				Persons: []filmPerson{
					{
						ID:         1,
						Name:       "Lana Wachowski",
						Profession: "director",
					},
				},
			},
		},
	}

	client := setupTestClient(t, mockResponses)

	films, err := client.FilmsByName("The Matrix", 2)
	if err != nil {
		t.Fatalf("FilmsByName returned error: %v", err)
	}

	if len(films) != 2 {
		t.Fatalf("Expected 2 films, got %d", len(films))
	}

	if films[0].Name != "The Matrix" {
		t.Errorf("Expected first film name 'The Matrix', got %q", films[0].Name)
	}
	if films[1].Name != "The Matrix Reloaded" {
		t.Errorf("Expected second film name 'The Matrix Reloaded', got %q", films[1].Name)
	}
}

func TestFilmsByName_LimitZero(t *testing.T) {
	mockResponses := map[string]mockResponse{
		"/v1.4/movie/search": {
			statusCode: http.StatusOK,
			body: searchResponse{
				Docs: []filmDoc{
					{
						ID:   123,
						Name: "Film 1",
						Year: 2020,
					},
					{
						ID:   456,
						Name: "Film 2",
						Year: 2020,
					},
				},
			},
		},
		"/v1.4/movie/123": {
			statusCode: http.StatusOK,
			body: filmResponseByID{
				Persons: []filmPerson{
					{
						ID:         1,
						Name:       "Director 1",
						Profession: "director",
					},
				},
			},
		},
		"/v1.4/movie/456": {
			statusCode: http.StatusOK,
			body: filmResponseByID{
				Persons: []filmPerson{
					{
						ID:         1,
						Name:       "Director 2",
						Profession: "director",
					},
				},
			},
		},
	}

	client := setupTestClient(t, mockResponses)

	films, err := client.FilmsByName("Test", 0)
	if err != nil {
		t.Fatalf("FilmsByName with limit 0 returned error: %v", err)
	}

	if len(films) != 2 {
		t.Fatalf("Expected 2 films with limit 0, got %d", len(films))
	}
}

func TestFilmsByName_DirectorError(t *testing.T) {
	mockResponses := map[string]mockResponse{
		"/v1.4/movie/search": {
			statusCode: http.StatusOK,
			body: searchResponse{
				Docs: []filmDoc{
					{
						ID:   123,
						Name: "The Matrix",
						Year: 1999,
					},
					{
						ID:   456,
						Name: "The Matrix Reloaded",
						Year: 2003,
					},
				},
			},
		},
		"/v1.4/movie/123": {
			statusCode: http.StatusOK,
			body: filmResponseByID{
				Persons: []filmPerson{
					{
						ID:         1,
						Name:       "Lana Wachowski",
						Profession: "director",
					},
				},
			},
		},
		"/v1.4/movie/456": {
			statusCode: http.StatusInternalServerError,
			body:       map[string]string{"error": "director fetch failed"},
		},
	}

	client := setupTestClient(t, mockResponses)

	films, err := client.FilmsByName("The Matrix", 2)
	if err == nil {
		t.Fatal("FilmsByName returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "error setting director for") {
		t.Fatalf("Expected director error, got: %v", err)
	}
	if films != nil {
		t.Fatalf("Expected nil films on error, got: %+v", films)
	}
}

func TestFetchFilms_Success(t *testing.T) {
	mockResponses := map[string]mockResponse{
		"/v1.4/movie/search": {
			statusCode: http.StatusOK,
			body: searchResponse{
				Docs: []filmDoc{
					{
						ID:           123,
						Name:         "Inception",
						OriginalName: "Inception",
						Year:         2010,
						Countries: []filmCountry{
							{CountryName: "USA"},
							{CountryName: "UK"},
						},
					},
				},
			},
		},
	}

	client := setupTestClient(t, mockResponses)

	films, err := client.fetchFilms("Inception", 1)
	if err != nil {
		t.Fatalf("fetchFilms returned error: %v", err)
	}

	if len(films) != 1 {
		t.Fatalf("Expected 1 film, got %d", len(films))
	}

	expected := domain.Film{
		ID:           123,
		Name:         "Inception",
		OriginalName: "Inception",
		Year:         2010,
		Countries:    []string{"USA", "UK"},
	}

	if diff := cmp.Diff(expected, films[0]); diff != "" {
		t.Fatalf("Film mismatch (-want +got):\n%s", diff)
	}
}

func TestFetchFilms_NoResults(t *testing.T) {
	mockResponses := map[string]mockResponse{
		"/v1.4/movie/search": {
			statusCode: http.StatusOK,
			body: searchResponse{
				Docs: []filmDoc{},
			},
		},
	}

	client := setupTestClient(t, mockResponses)

	films, err := client.fetchFilms("NonExistentFilm", 1)
	if err == nil {
		t.Fatal("fetchFilms returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "can't find films") {
		t.Fatalf("Expected 'can't find films' error, got: %v", err)
	}
	if films != nil {
		t.Fatalf("Expected nil films on error, got: %+v", films)
	}
}

func TestFetchFilms_MalformedURL(t *testing.T) {
	client := &Client{
		baseURL:    "http://invalid\ntest\nurl",
		key:        "test-key",
		httpClient: http.DefaultClient,
		logger:     newTestLogger(),
	}

	films, err := client.fetchFilms("Test", 1)
	if err == nil {
		t.Fatal("fetchFilms returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "parse URL") {
		t.Fatalf("Expected parse URL error, got: %v", err)
	}
	if films != nil {
		t.Fatalf("Expected nil films on error, got: %+v", films)
	}
}

func TestSetDirectors_Success(t *testing.T) {
	mockResponses := map[string]mockResponse{
		"/v1.4/movie/123": {
			statusCode: http.StatusOK,
			body: filmResponseByID{
				Persons: []filmPerson{
					{
						ID:         1,
						Name:       "Lana Wachowski",
						Profession: "director",
					},
					{
						ID:         2,
						Name:       "Lilly Wachowski",
						Profession: "director",
					},
					{
						ID:         3,
						Name:       "Keanu Reeves",
						Profession: "actor",
					},
					{
						ID:         4,
						Name:       "Hugo Weaving",
						Profession: "actor",
					},
				},
			},
		},
	}

	client := setupTestClient(t, mockResponses)

	film := &domain.Film{
		ID:   123,
		Name: "The Matrix",
	}

	err := client.setDirectors(film)
	if err != nil {
		t.Fatalf("setDirectors returned error: %v", err)
	}

	expected := []string{"Lana Wachowski", "Lilly Wachowski"}
	if diff := cmp.Diff(expected, film.Directors); diff != "" {
		t.Fatalf("Directors mismatch (-want +got):\n%s", diff)
	}
}

func TestSetDirectors_NoDirectors(t *testing.T) {
	mockResponses := map[string]mockResponse{
		"/v1.4/movie/123": {
			statusCode: http.StatusOK,
			body: filmResponseByID{
				Persons: []filmPerson{
					{
						ID:         1,
						Name:       "Some Actor",
						Profession: "actor",
					},
					{
						ID:         2,
						Name:       "Some Producer",
						Profession: "producer",
					},
				},
			},
		},
	}

	client := setupTestClient(t, mockResponses)

	film := &domain.Film{
		ID:   123,
		Name: "Some Film",
	}

	err := client.setDirectors(film)
	if err != nil {
		t.Fatalf("setDirectors returned error: %v", err)
	}

	if len(film.Directors) != 0 {
		t.Fatalf("Expected 0 directors, got %d: %v", len(film.Directors), film.Directors)
	}
}

func TestSetDirectors_APIError(t *testing.T) {
	mockResponses := map[string]mockResponse{
		"/v1.4/movie/123": {
			statusCode: http.StatusNotFound,
			body:       map[string]string{"error": "film not found"},
		},
	}

	client := setupTestClient(t, mockResponses)

	film := &domain.Film{
		ID:   123,
		Name: "NonExistent",
	}

	err := client.setDirectors(film)
	if err == nil {
		t.Fatal("setDirectors returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "API response status 404") {
		t.Fatalf("Expected API error, got: %v", err)
	}
}

func TestFetchFilms_EmptyResponse(t *testing.T) {
	mockResponses := map[string]mockResponse{
		"/v1.4/movie/search": {
			statusCode: http.StatusOK,
			body:       nil,
		},
	}

	client := setupTestClient(t, mockResponses)

	films, err := client.fetchFilms("Test", 1)
	if err == nil {
		t.Fatal("fetchFilms returned nil error, want error")
	}
	if films != nil {
		t.Fatalf("Expected nil films on error, got: %+v", films)
	}
}

func TestDirectorNames_Success(t *testing.T) {
	fr := &filmResponseByID{
		Persons: []filmPerson{
			{
				ID:         1,
				Name:       "Lana Wachowski",
				Profession: "director",
			},
			{
				ID:         2,
				Name:       "Lilly Wachowski",
				Profession: "director",
			},
			{
				ID:         3,
				Name:       "Keanu Reeves",
				Profession: "actor",
			},
		},
	}

	directors, err := fr.directorNames()
	if err != nil {
		t.Fatalf("directorNames returned error: %v", err)
	}

	expected := []string{"Lana Wachowski", "Lilly Wachowski"}
	if diff := cmp.Diff(expected, directors); diff != "" {
		t.Fatalf("Directors mismatch (-want +got):\n%s", diff)
	}
}

func TestDirectorNames_EmptyList(t *testing.T) {
	fr := &filmResponseByID{
		Persons: []filmPerson{},
	}

	directors, err := fr.directorNames()
	if err != nil {
		t.Fatalf("directorNames returned error: %v", err)
	}

	if len(directors) != 0 {
		t.Fatalf("Expected 0 directors, got %d", len(directors))
	}
}

func TestDirectorNames_OnlyActors(t *testing.T) {
	fr := &filmResponseByID{
		Persons: []filmPerson{
			{
				ID:         1,
				Name:       "Keanu Reeves",
				Profession: "actor",
			},
			{
				ID:         2,
				Name:       "Hugo Weaving",
				Profession: "actor",
			},
		},
	}

	directors, err := fr.directorNames()
	if err != nil {
		t.Fatalf("directorNames returned error: %v", err)
	}

	if len(directors) != 0 {
		t.Fatalf("Expected 0 directors, got %d: %v", len(directors), directors)
	}
}

func TestDirectorNames_MixedProfessions(t *testing.T) {
	fr := &filmResponseByID{
		Persons: []filmPerson{
			{
				ID:         1,
				Name:       "Lana Wachowski",
				Profession: "director",
			},
			{
				ID:         2,
				Name:       "John Doe",
				Profession: "producer",
			},
			{
				ID:         3,
				Name:       "Jane Smith",
				Profession: "director",
			},
			{
				ID:         4,
				Name:       "Bob Johnson",
				Profession: "writer",
			},
		},
	}

	directors, err := fr.directorNames()
	if err != nil {
		t.Fatalf("directorNames returned error: %v", err)
	}

	expected := []string{"Lana Wachowski", "Jane Smith"}
	if diff := cmp.Diff(expected, directors); diff != "" {
		t.Fatalf("Directors mismatch (-want +got):\n%s", diff)
	}
}

func TestSetDirectors_MissingPersons(t *testing.T) {
	mockResponses := map[string]mockResponse{
		"/v1.4/movie/123": {
			statusCode: http.StatusOK,
			body:       filmResponseByID{},
		},
	}

	client := setupTestClient(t, mockResponses)

	film := &domain.Film{
		ID:   123,
		Name: "Some Film",
	}

	err := client.setDirectors(film)
	if err != nil {
		t.Fatalf("setDirectors returned error: %v", err)
	}

	if len(film.Directors) != 0 {
		t.Fatalf("Expected 0 directors, got %d", len(film.Directors))
	}
}
