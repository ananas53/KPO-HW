package storage_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ananas53/kpo-hw/hw-2/internal/adapter/storage"
	"github.com/ananas53/kpo-hw/hw-2/internal/domain"
)

func TestJSONUpdateThroughSymlinkUpdatesTarget(t *testing.T) {
	directory := t.TempDir()
	target := filepath.Join(directory, "grades.json")
	repo := storage.NewJSON(target)
	mustSucceed(t, repo.Update(func(book *domain.Gradebook) error {
		return book.AddStudent("anna", "Анна")
	}))
	link := filepath.Join(directory, "alias.json")
	mustSucceed(t, os.Symlink("grades.json", link))
	mustSucceed(t, storage.NewJSON(link).Update(func(book *domain.Gradebook) error {
		return book.AddStudent("ivan", "Иван")
	}))
	actual, err := os.Readlink(link)
	mustSucceed(t, err)
	if actual != "grades.json" {
		t.Fatalf("symlink target = %q, want grades.json", actual)
	}
	book, err := repo.Load()
	mustSucceed(t, err)
	students := book.Students()
	if len(students) != 2 || students[0].ID() != "anna" || students[1].ID() != "ivan" {
		t.Fatalf("students in target = %v, want anna and ivan", students)
	}
}

func TestJSONUpdateThroughSymlinkHonorsTargetLock(t *testing.T) {
	directory := t.TempDir()
	target := filepath.Join(directory, "grades.json")
	repo := storage.NewJSON(target)
	mustSucceed(t, repo.Update(func(book *domain.Gradebook) error {
		return book.AddStudent("anna", "Анна")
	}))
	link := filepath.Join(directory, "alias.json")
	mustSucceed(t, os.Symlink("grades.json", link))
	mustSucceed(t, os.Mkdir(target+".lock", 0o700))
	err := storage.NewJSON(link).Update(func(*domain.Gradebook) error {
		t.Fatal("update through a link bypassed the target lock")
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "заблокирована") {
		t.Fatalf("Update error = %v, want a lock error", err)
	}
}

func TestJSONUpdateRejectsDanglingSymlink(t *testing.T) {
	directory := t.TempDir()
	link := filepath.Join(directory, "alias.json")
	mustSucceed(t, os.Symlink("missing.json", link))
	err := storage.NewJSON(link).Update(func(*domain.Gradebook) error {
		t.Fatal("update through a dangling link reached the callback")
		return nil
	})
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Update error = %v, want os.ErrNotExist", err)
	}
	actual, err := os.Readlink(link)
	mustSucceed(t, err)
	if actual != "missing.json" {
		t.Fatalf("symlink target = %q, want missing.json", actual)
	}
}
