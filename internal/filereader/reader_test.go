package filereader

import (
	"io"
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func newTestLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}

func TestNew_FileNotFound(t *testing.T) {
	r, err := New("testdata/fake.txt", newTestLogger())
	if err == nil {
		t.Fatal("New() with no file err == nil, want error")
	}
	if r != nil {
		t.Fatalf("New() with no file returned a reader, want nil")
	}
}

func TestRead_AllLines(t *testing.T) {
	r, err := New("testdata/films.txt", newTestLogger())
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}
	defer r.Close()

	got, err := r.Read(10)
	if err != nil {
		t.Fatalf("Read() returned an error: %v", err)
	}

	want := []string{"film1", "film2", "film3", "film4", "film5", "film6", "film7"}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("mismatch (-want +got):\n%s", diff)
	}
}

func TestRead_WithLimit(t *testing.T) {
	r, err := New("testdata/films.txt", newTestLogger())
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}
	defer r.Close()

	got, err := r.Read(3)
	if err != nil {
		t.Fatalf("Read() returned an error: %v", err)
	}

	want := []string{"film1", "film2", "film3"}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("mismatch (-want +got):\n%s", diff)
	}
}

func TestRead_EmptyFile(t *testing.T) {
	r, err := New("testdata/empty.txt", newTestLogger())
	if err != nil {
		t.Fatalf("New() returned an error: %s", err.Error())
	}
	if r == nil {
		t.Fatalf("New() returned a nil reader")
	}
	defer r.Close()

	got, err := r.Read(10)
	if err != nil {
		t.Fatalf("Read() returned an error: %s", err.Error())
	}
	if len(got) != 0 {
		t.Fatalf("Read() returned %d lines, want 0", len(got))
	}
	want := make([]string, 0)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("mismatch (-want +got):\n%s", diff)
	}
}
