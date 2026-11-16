package storage

import "github.com/ananas53/kpo-hw/hw-2/internal/domain"

// Memory keeps the demo in a separate gradebook for the lifetime of its process.
type Memory struct {
	book *domain.Gradebook
}

func NewMemory() *Memory {
	return &Memory{book: domain.NewGradebook()}
}

func (r *Memory) Load() (*domain.Gradebook, error) {
	return r.book.Clone(), nil
}

func (r *Memory) Update(change func(*domain.Gradebook) error) error {
	draft := r.book.Clone()

	err := change(draft)

	if err != nil {
		return err
	}
	// The callback may keep the draft, so store a separate snapshot.
	r.book = draft.Clone()
	return nil
}
