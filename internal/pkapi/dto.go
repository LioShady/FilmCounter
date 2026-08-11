package pkapi

type searchResponse struct {
	Docs []filmDoc `json:"docs"`
}

type filmDoc struct {
	ID           int           `json:"id"`
	Name         string        `json:"name"`
	OriginalName string        `json:"alternativeName"`
	Year         int           `json:"year"`
	Countries    []filmCountry `json:"countries"`
}

type filmCountry struct {
	CountryName string `json:"name"`
}

type filmResponseByID struct {
	Persons []filmPerson `json:"persons"`
}

type filmPerson struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Profession string `json:"enProfession"`
}

func (fp *filmResponseByID) directorNames() ([]string, error) {
	//var names []string
	names := make([]string, 0)
	for _, p := range fp.Persons {
		if p.Profession == "director" {
			names = append(names, p.Name)
		}
	}
	return names, nil
}
