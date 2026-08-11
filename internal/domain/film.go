package domain

import "fmt"

// Film contains information about a film
type Film struct {
	ID           int
	Name         string
	OriginalName string
	Year         int
	Countries    []string
	Directors    []string
}

func (f Film) String() string {
	return fmt.Sprintf("Название: %s \nОригинальное название: %s \nГод выхода: %d \nСтрана: %s \nРежиссер: %s \n", f.Name, f.OriginalName, f.Year, f.Countries, f.Directors)
}
