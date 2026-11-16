package storage

import (
	"errors"
	"strings"
	"testing"

	"github.com/ananas53/kpo-hw/hw-2/internal/domain"
)

func TestReadGradebookRevalidatesStoredEvaluations(t *testing.T) {
	state := `{
		"students":[{"id":"anna","name":"Анна"}],
		"criteria":[{"id":"practice","name":"Практика","max_score":10,"weight":1}],
		"evaluations":[{"student_id":"anna","criterion_id":"removed","score":9}]
	}`
	if _, err := readGradebook(strings.NewReader(state)); !errors.Is(err, domain.ErrCriterionNotFound) {
		t.Fatalf("readGradebook error = %v, want ErrCriterionNotFound", err)
	}
}
