package storage

import (
	"fmt"

	"github.com/ananas53/kpo-hw/hw-2/internal/domain"
)

type snapshot struct {
	Students    []studentRow    `json:"students"`
	Criteria    []criterionRow  `json:"criteria"`
	Evaluations []evaluationRow `json:"evaluations"`
}

type studentRow struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type criterionRow struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	MaxScore float64 `json:"max_score"`
	Weight   float64 `json:"weight"`
}

type evaluationRow struct {
	StudentID   string   `json:"student_id"`
	CriterionID string   `json:"criterion_id"`
	Score       *float64 `json:"score"`
	Comment     string   `json:"comment,omitempty"`
}

func makeSnapshot(book *domain.Gradebook) (snapshot, error) {
	state := snapshot{
		Students:    make([]studentRow, 0),
		Criteria:    make([]criterionRow, 0),
		Evaluations: make([]evaluationRow, 0),
	}

	for _, student := range book.Students() {
		state.Students = append(state.Students, studentRow{ID: student.ID(), Name: student.Name()})
		evaluations, err := book.Evaluations(student.ID())

		if err != nil {
			return snapshot{}, err
		}

		for _, evaluation := range evaluations {
			score := evaluation.Score().Value()
			state.Evaluations = append(state.Evaluations, evaluationRow{
				StudentID:   evaluation.StudentID(),
				CriterionID: evaluation.CriterionID(),
				Score:       &score,
				Comment:     evaluation.Comment(),
			})
		}
	}

	for _, criterion := range book.Criteria() {
		state.Criteria = append(state.Criteria, criterionRow{
			ID:       criterion.ID(),
			Name:     criterion.Name(),
			MaxScore: criterion.MaxScore(),
			Weight:   criterion.Weight(),
		})
	}

	return state, nil
}

func (s snapshot) restore() (*domain.Gradebook, error) {
	book := domain.NewGradebook()

	for _, row := range s.Students {
		err := book.AddStudent(row.ID, row.Name)

		if err != nil {
			return nil, err
		}
	}

	for _, row := range s.Criteria {
		err := book.AddCriterion(row.ID, row.Name, row.MaxScore, row.Weight)

		if err != nil {
			return nil, err
		}
	}

	seen := make(map[[2]string]struct{}, len(s.Evaluations))

	for _, row := range s.Evaluations {
		pair := [2]string{row.StudentID, row.CriterionID}

		_, duplicate := seen[pair]

		if duplicate {
			return nil, fmt.Errorf("повторная оценка: студент %q, критерий %q", row.StudentID, row.CriterionID)
		}

		if row.Score == nil {
			return nil, fmt.Errorf("отсутствует балл: студент %q, критерий %q", row.StudentID, row.CriterionID)
		}

		err := book.SetEvaluation(row.StudentID, row.CriterionID, *row.Score, row.Comment)

		if err != nil {
			return nil, err
		}

		seen[pair] = struct{}{}
	}

	return book, nil
}
