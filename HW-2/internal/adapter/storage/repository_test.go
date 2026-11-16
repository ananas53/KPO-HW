package storage_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/ananas53/kpo-hw/hw-2/internal/adapter/storage"
	"github.com/ananas53/kpo-hw/hw-2/internal/application"
	"github.com/ananas53/kpo-hw/hw-2/internal/domain"
)

func TestRepositorySnapshots(t *testing.T) {
	repositories := []struct {
		name string
		new  func(*testing.T) application.Repository
	}{
		{"memory", func(*testing.T) application.Repository { return storage.NewMemory() }},
		{"json", func(t *testing.T) application.Repository {
			return storage.NewJSON(filepath.Join(t.TempDir(), "pos.json"))
		}},
	}
	for _, implementation := range repositories {
		t.Run(implementation.name, func(t *testing.T) {
			seeded := func(t *testing.T) application.Repository {
				t.Helper()
				repo := implementation.new(t)
				mustSucceed(t, repo.Update(func(book *domain.Gradebook) error {
					if err := book.AddStudent("anna", "Анна"); err != nil {
						return err
					}
					if err := book.AddCriterion("practice", "Практика", 10, 1); err != nil {
						return err
					}
					return book.SetEvaluation("anna", "practice", 7, "")
				}))
				return repo
			}
			t.Run("load returns a detached snapshot", func(t *testing.T) {
				repo := seeded(t)
				snapshot, err := repo.Load()
				mustSucceed(t, err)
				mustSucceed(t, snapshot.AddStudent("ivan", "Иван"))
				mustSucceed(t, snapshot.AddCriterion("exam", "Экзамен", 10, 1))
				mustSucceed(t, snapshot.SetEvaluation("anna", "practice", 8, ""))
				stored, err := repo.Load()
				mustSucceed(t, err)
				if len(stored.Students()) != 1 || len(stored.Criteria()) != 1 {
					t.Fatalf("editing a snapshot changed the repository: %d students, %d criteria",
						len(stored.Students()), len(stored.Criteria()))
				}
				assertScore(t, stored, 7)
			})

			t.Run("failed update discards changes", func(t *testing.T) {
				repo := seeded(t)
				rejected := errors.New("reject update")
				err := repo.Update(func(book *domain.Gradebook) error {
					if err := book.SetEvaluation("anna", "practice", 8, ""); err != nil {
						return err
					}
					return rejected
				})
				if !errors.Is(err, rejected) {
					t.Fatalf("Update error = %v, want %v", err, rejected)
				}
				stored, err := repo.Load()
				mustSucceed(t, err)
				assertScore(t, stored, 7)
			})

			t.Run("successful update saves a detached snapshot", func(t *testing.T) {
				repo := seeded(t)
				var draft *domain.Gradebook
				mustSucceed(t, repo.Update(func(book *domain.Gradebook) error {
					draft = book
					return book.SetEvaluation("anna", "practice", 8, "")
				}))
				mustSucceed(t, draft.SetEvaluation("anna", "practice", 9, ""))
				stored, err := repo.Load()
				mustSucceed(t, err)
				assertScore(t, stored, 8)
			})
		})
	}
}

func mustSucceed(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertScore(t *testing.T, book *domain.Gradebook, want float64) {
	t.Helper()
	grades, err := book.Evaluations("anna")
	mustSucceed(t, err)
	if len(grades) != 1 || grades[0].Score().Value() != want {
		t.Fatalf("evaluations = %v, want one evaluation with score %g", grades, want)
	}
}
