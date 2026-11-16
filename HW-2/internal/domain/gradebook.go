package domain

import (
	"fmt"
	"maps"
	"slices"
)

// Gradebook is the aggregate for a single course.
type Gradebook struct {
	students    map[string]Student
	rubric      map[string]Criterion
	evaluations map[string]map[string]Evaluation
}

func NewGradebook() *Gradebook {
	return &Gradebook{
		students:    make(map[string]Student),
		rubric:      make(map[string]Criterion),
		evaluations: make(map[string]map[string]Evaluation),
	}
}

// Clone returns an independent copy of the gradebook.
func (b *Gradebook) Clone() *Gradebook {
	evaluations := make(map[string]map[string]Evaluation, len(b.evaluations))

	for studentID, grades := range b.evaluations {
		evaluations[studentID] = maps.Clone(grades)
	}

	return &Gradebook{
		students:    maps.Clone(b.students),
		rubric:      maps.Clone(b.rubric),
		evaluations: evaluations,
	}
}

func (b *Gradebook) AddStudent(id, name string) error {
	id, name, err := identity(id, name)

	if err != nil {
		return err
	}

	_, exists := b.students[id]

	if exists {
		return fmt.Errorf("студент %q уже существует", id)
	}

	b.students[id] = Student{id: id, name: name}
	return nil
}

func (b *Gradebook) AddCriterion(id, name string, maxScore, weight float64) error {
	id, name, err := identity(id, name)

	if err != nil {
		return err
	}

	_, exists := b.rubric[id]

	if exists {
		return fmt.Errorf("критерий %q уже существует", id)
	}

	if !finite(maxScore) || maxScore <= 0 {
		return fmt.Errorf("максимум критерия %q должен быть конечным положительным числом", id)
	}

	if !finite(weight) || weight <= 0 {
		return fmt.Errorf("вес критерия %q должен быть конечным положительным числом", id)
	}

	b.rubric[id] = Criterion{id: id, name: name, maxScore: maxScore, weight: weight}
	return nil
}

// SetEvaluation records a grade or replaces the current grade for this pair.
func (b *Gradebook) SetEvaluation(studentID, criterionID string, points float64, comment string) error {
	evaluation, err := b.newEvaluation(studentID, criterionID, points, comment)

	if err != nil {
		return err
	}

	if b.evaluations[studentID] == nil {
		b.evaluations[studentID] = make(map[string]Evaluation)
	}

	b.evaluations[studentID][criterionID] = evaluation
	return nil
}

func (b *Gradebook) RemoveEvaluation(studentID, criterionID string) error {
	_, err := b.student(studentID)

	if err != nil {
		return err
	}

	evaluations := b.evaluations[studentID]

	_, exists := evaluations[criterionID]

	if !exists {
		return fmt.Errorf("%w: студент %q, критерий %q", ErrEvaluationNotFound, studentID, criterionID)
	}

	delete(evaluations, criterionID)
	return nil
}

func (b *Gradebook) Students() []Student {
	return valuesByID(b.students)
}

func (b *Gradebook) Criteria() []Criterion {
	return valuesByID(b.rubric)
}

func (b *Gradebook) Evaluations(studentID string) ([]Evaluation, error) {
	_, err := b.student(studentID)

	if err != nil {
		return nil, err
	}

	return valuesByID(b.evaluations[studentID]), nil
}

func (b *Gradebook) student(id string) (Student, error) {
	student, ok := b.students[id]

	if !ok {
		return Student{}, fmt.Errorf("%w: %q", ErrStudentNotFound, id)
	}

	return student, nil
}

func valuesByID[T any](entries map[string]T) []T {
	values := make([]T, 0, len(entries))

	for _, id := range slices.Sorted(maps.Keys(entries)) {
		values = append(values, entries[id])
	}

	return values
}
