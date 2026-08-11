package pkapi

import (
	"FilmCounter/internal/domain"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// FilmByName fetches first film information found by name.
// The name does not need to be exact.
func (c *Client) FilmByName(name string) (domain.Film, error) {
	c.logger.Printf("fetching film by name: %q", name)
	films, err := c.fetchFilms(name, 1)
	if err != nil {
		return domain.Film{}, err
	}
	film := films[0]
	err = c.setDirectors(&film)
	if err != nil {
		return domain.Film{}, err
	}
	c.logger.Printf("successfully fetched single film (requested name: %q found %q (%d))", name, film.Name, film.Year)

	return film, nil
}

// FilmsByName fetches multiple film information found by name.
// The name does not need to be exact.
func (c *Client) FilmsByName(name string, limit int) ([]domain.Film, error) {
	c.logger.Printf("fetching films by name: %q, limit=%d", name, limit)
	films, err := c.fetchFilms(name, limit)
	if err != nil {
		return nil, err
	}
	for i := range films {
		err = c.setDirectors(&films[i])
		if err != nil {
			return nil, fmt.Errorf("error setting director for %q: %w", films[i].Name, err)
		}
	}
	c.logger.Println("successfully fetched multiple films")
	return films, nil
}

// fetchFilms performs the actual API request to search for films by name.
// It returns a slice of films with basic information (without directors).
// The limit parameter controls how many films to return.
func (c *Client) fetchFilms(name string, limit int) ([]domain.Film, error) {
	fullURL := c.baseURL + "/v1.4/movie/search"
	parsedUrl, err := url.Parse(fullURL)
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}
	query := parsedUrl.Query()
	query.Set("query", name)
	query.Set("page", "1")
	query.Set("limit", strconv.Itoa(limit))
	parsedUrl.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", parsedUrl.String(), nil)

	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Add("X-API-KEY", c.key)
	req.Header.Add("accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("api response status %d: %s", res.StatusCode, body)
	}

	var filmResponse searchResponse
	err = json.NewDecoder(res.Body).Decode(&filmResponse)
	if err != nil {
		return nil, fmt.Errorf("decode movie response: %w", err)
	}

	if len(filmResponse.Docs) == 0 {
		return nil, fmt.Errorf("can't find films with name '%s'", name)
	}
	var films = make([]domain.Film, len(filmResponse.Docs))
	for i, doc := range filmResponse.Docs {
		var countries []string
		if len(doc.Countries) > 0 {
			for _, country := range doc.Countries {
				countries = append(countries, country.CountryName)
			}
		}
		film := domain.Film{
			ID:           doc.ID,
			Name:         doc.Name,
			OriginalName: doc.OriginalName,
			Year:         doc.Year,
			Countries:    countries,
		}

		films[i] = film
	}
	return films, nil
}

// setDirectors fetches additional information about a film by its ID
// and adds director names to the provided Film struct.
func (c *Client) setDirectors(film *domain.Film) error {
	c.logger.Printf("setting directors for %q (%d)", film.Name, film.Year)
	fullURL := c.baseURL + "/v1.4/movie/" + strconv.Itoa(film.ID)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Add("X-API-KEY", c.key)
	req.Header.Add("accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("API response status %d: %s", res.StatusCode, body)
	}
	var filmResponse filmResponseByID

	err = json.NewDecoder(res.Body).Decode(&filmResponse)

	if err != nil {
		return fmt.Errorf("decode movie response: %w", err)
	}

	film.Directors, err = filmResponse.directorNames()
	if err != nil {
		return err
	}
	return nil
}
