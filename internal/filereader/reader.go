package filereader

import (
	"bufio"
	"log"
	"os"
)

// Reader reads film names from file
type Reader struct {
	file   *os.File
	logger *log.Logger
}

// New creates new Reader
func New(path string, logger *log.Logger) (*Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &Reader{file, logger}, nil
}

func (r *Reader) Close() {
	err := r.file.Close()
	if err != nil {
		r.logger.Fatal(err)
	}
}

func (r *Reader) Read(limit int) ([]string, error) {
	scanner := bufio.NewScanner(r.file)
	var films = make([]string, 0)
	i := 0
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		films = append(films, line)
		i = i + 1
		if i >= limit {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return films, nil
}
