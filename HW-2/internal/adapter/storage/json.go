package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ananas53/kpo-hw/hw-2/internal/adapter/gradebookpath"
	"github.com/ananas53/kpo-hw/hw-2/internal/domain"
)

type JSON struct {
	path string
}

func NewJSON(path string) *JSON {
	return &JSON{path: path}
}

func (r *JSON) Load() (*domain.Gradebook, error) {
	file, err := os.Open(r.path)

	if errors.Is(err, os.ErrNotExist) {
		return domain.NewGradebook(), nil
	}

	if err != nil {
		return nil, fmt.Errorf("открыть ведомость: %w", err)
	}

	defer file.Close()

	book, err := readGradebook(file)

	if err != nil {
		return nil, fmt.Errorf("прочитать ведомость %q: %w", r.path, err)
	}

	return book, nil
}

func (r *JSON) Update(change func(*domain.Gradebook) error) error {
	path, lock, err := gradebookpath.Resolve(r.path)

	if err != nil {
		return fmt.Errorf("определить путь ведомости %q: %w", r.path, err)
	}

	target := JSON{path: path}

	err = os.Mkdir(lock, 0o700)

	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("ведомость заблокирована для записи (%s)", lock)
		}

		return fmt.Errorf("заблокировать ведомость: %w", err)
	}

	defer os.Remove(lock)

	book, err := target.Load()

	if err != nil {
		return err
	}

	err = change(book)

	if err != nil {
		return err
	}

	err = target.save(book)

	if err != nil {
		return fmt.Errorf("сохранить ведомость %q: %w", path, err)
	}

	return nil
}

func (r *JSON) save(book *domain.Gradebook) error {
	state, err := makeSnapshot(book)

	if err != nil {
		return err
	}

	file, err := os.CreateTemp(filepath.Dir(r.path), ".pos-*")

	if err != nil {
		return err
	}

	defer os.Remove(file.Name())

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	writeErr := encoder.Encode(state)

	if writeErr == nil {
		writeErr = file.Sync()
	}

	err = errors.Join(writeErr, file.Close())

	if err != nil {
		return err
	}

	return os.Rename(file.Name(), r.path)
}

func readGradebook(reader io.Reader) (*domain.Gradebook, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var state *snapshot

	err := decoder.Decode(&state)

	if err != nil {
		return nil, err
	}

	if state == nil {
		return nil, errors.New("ожидается JSON-объект ведомости")
	}

	err = decoder.Decode(new(any))

	if err != io.EOF {
		return nil, errors.New("после ведомости обнаружены лишние данные")
	}

	return state.restore()
}
