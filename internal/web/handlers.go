package web

import (
	"FilmCounter/internal/domain"
	"net/http"
	"sort"
	"strconv"
)

func (s *Server) handleFilms(w http.ResponseWriter, r *http.Request) {
	films, err := s.store.AllFilms()
	if err != nil {
		s.logger.Printf("Error getting films: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Films any
	}{
		Films: films,
	}

	if err := s.templates.ExecuteTemplate(w, "films.html", data); err != nil {
		s.logger.Printf("Error executing template: %v", err)
	}
}

type DirectorFilms struct {
	Director string
	Films    []domain.Film
}

func (s *Server) handleDirectors(w http.ResponseWriter, r *http.Request) {
	filmsByDirector, err := s.store.AllFilmsByDirectors()
	if err != nil {
		s.logger.Printf("failed to fetch films by directors: %v", err)
		http.Error(w, "failed to fetch films by directors", http.StatusInternalServerError)
		return
	}

	groups := make([]DirectorFilms, 0, len(filmsByDirector))
	total := 0
	for director, films := range filmsByDirector {
		total += len(films)
		sort.Slice(films, func(i, j int) bool {
			if films[i].Year != films[j].Year {
				return films[i].Year < films[j].Year
			}
			return films[i].Name < films[j].Name
		})

		groups = append(groups, DirectorFilms{
			Director: director,
			Films:    films,
		})
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Director < groups[j].Director
	})

	data := struct {
		Groups []DirectorFilms
		Total  int
	}{
		Groups: groups,
		Total:  total,
	}

	if err := s.templates.ExecuteTemplate(w, "directors.html", data); err != nil {
		s.logger.Printf("failed to render directors page: %v", err)
	}
}

type importSearchData struct {
	Query string
	Count string
	Films []domain.Film
}

func (s *Server) handleImportForm(w http.ResponseWriter, r *http.Request) {
	if err := s.templates.ExecuteTemplate(w, "import.html", nil); err != nil {
		s.logger.Printf("render import form: %v", err)
	}
}

func (s *Server) handleImportSearch(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	query := r.FormValue("name")
	if query == "" {
		http.Error(w, "film name is required", http.StatusBadRequest)
		return
	}
	count := r.FormValue("count")
	var lim int
	if count == "" {
		lim = 3
	} else {
		var err error
		lim, err = strconv.Atoi(count)
		if err != nil {
			http.Error(w, "invalid count", http.StatusBadRequest)
		}
	}

	films, err := s.pkClient.FilmsByName(query, lim)
	if err != nil {
		s.logger.Printf("search films by name %q: %v", query, err)
		http.Error(w, "failed to search films", http.StatusBadGateway)
		return
	}

	data := importSearchData{
		Query: query,
		Count: count,
		Films: films,
	}

	if err := s.templates.ExecuteTemplate(w, "import.html", data); err != nil {
		s.logger.Printf("render import results: %v", err)
	}
}

func (s *Server) handleImportAdd(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, "bad film id", http.StatusBadRequest)
		return
	}

	year, err := strconv.Atoi(r.FormValue("year"))
	if err != nil {
		http.Error(w, "bad film year", http.StatusBadRequest)
		return
	}

	film := domain.Film{
		ID:           id,
		Name:         r.FormValue("name"),
		OriginalName: r.FormValue("original_name"),
		Year:         year,
		Countries:    r.Form["countries"],
		Directors:    r.Form["directors"],
	}

	if err := s.store.AddFilm(film); err != nil {
		s.logger.Printf("add film %q: %v", film.Name, err)
		http.Error(w, "failed to add film", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/films", http.StatusSeeOther)
}

func (s *Server) handleRemove(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "bad film id", http.StatusBadRequest)
		return
	}

	err = s.store.DeleteFilmByID(id)
	if err != nil {
		s.logger.Printf("failed to delete film %q: %v", id, err)
		http.Error(w, "failed to delete film", http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/films", http.StatusSeeOther)
}
