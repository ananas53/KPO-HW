package domain_test

import (
	"errors"
	"math"
	"testing"

	"github.com/ananas53/kpo-hw/hw-2/internal/domain"
)

func newGradebook(t *testing.T) *domain.Gradebook {
	t.Helper()
	book := domain.NewGradebook()
	requireSuccess(t, book.AddStudent("anna", "Анна"))
	requireSuccess(t, book.AddCriterion("practice", "Практика", 20, 3))
	requireSuccess(t, book.AddCriterion("review", "Защита", 10, 2))
	return book
}

func requireSuccess(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestGradebookRejectsInvalidEvaluations(t *testing.T) {
	book := newGradebook(t)
	requireSuccess(t, book.SetEvaluation("anna", "practice", 20, "максимум допустим"))
	tests := []struct {
		name      string
		student   string
		criterion string
		score     float64
		wantError error
	}{
		{"above maximum", "anna", "practice", 20.01, nil},
		{"negative", "anna", "practice", -1, nil},
		{"not a number", "anna", "practice", math.NaN(), nil},
		{"infinity", "anna", "practice", math.Inf(1), nil},
		{"unknown student", "other", "practice", 5, domain.ErrStudentNotFound},
		{"outside rubric", "anna", "unknown", 5, domain.ErrCriterionNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := book.SetEvaluation(tt.student, tt.criterion, tt.score, "")
			if err == nil || (tt.wantError != nil && !errors.Is(err, tt.wantError)) {
				t.Fatalf("SetEvaluation error = %v, want validation error (%v)", err, tt.wantError)
			}
			grades, err := book.Evaluations("anna")
			requireSuccess(t, err)
			if len(grades) != 1 || grades[0].Score().Value() != 20 {
				t.Fatalf("invalid grade changed the current evaluation: %+v", grades)
			}
		})
	}
}

func TestGradebookValidatesRubricAndIdentity(t *testing.T) {
	book := newGradebook(t)
	if err := book.AddStudent("anna", "Другая Анна"); err == nil {
		t.Fatal("duplicate student accepted")
	}
	if err := book.AddStudent("a b", "Пётр"); err == nil {
		t.Fatal("student ID with a space accepted")
	}
	tests := []struct {
		name   string
		max    float64
		weight float64
	}{
		{"zero maximum", 0, 1},
		{"zero weight", 10, 0},
		{"negative weight", 10, -1},
		{"nonfinite maximum", math.Inf(1), 1},
		{"nonfinite weight", 10, math.NaN()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := book.AddCriterion("invalid", tt.name, tt.max, tt.weight); err == nil {
				t.Fatal("invalid criterion accepted")
			}
		})
	}
	if err := book.AddCriterion("practice", "Другой максимум", 100, 1); err == nil {
		t.Fatal("duplicate criterion accepted")
	}
	if got := len(book.Criteria()); got != 2 {
		t.Fatalf("rubric has %d criteria, want 2", got)
	}
}

func TestCalculatorUsesCurrentGradesAndSelectedStrategy(t *testing.T) {
	book := newGradebook(t)
	calculator := domain.NewCalculator()
	official := domain.AllCriteria{}
	progress := domain.GradedOnly{}
	assertTotal := func(strategy domain.ScoringStrategy, want float64, graded int) {
		t.Helper()
		result, err := calculator.Calculate(book, "anna", strategy)
		requireSuccess(t, err)
		if math.Abs(result.Points-want) > 1e-10 || result.Graded != graded || result.Criteria != 2 {
			t.Fatalf("total = %+v, want %.2f with %d/2 grades", result, want, graded)
		}
	}
	assertTotal(official, 0, 0)
	if _, err := calculator.Calculate(book, "anna", progress); !errors.Is(err, domain.ErrNoGrades) {
		t.Fatalf("ungraded progress = %v, want ErrNoGrades", err)
	}
	requireSuccess(t, book.SetEvaluation("anna", "practice", 16, ""))
	assertTotal(official, 4.8, 1)
	assertTotal(progress, 8, 1)
	requireSuccess(t, book.SetEvaluation("anna", "review", 9, ""))
	assertTotal(official, 8.4, 2)
	requireSuccess(t, book.SetEvaluation("anna", "practice", 0, "нулевая оценка засчитывается"))
	assertTotal(official, 3.6, 2)
	assertTotal(progress, 3.6, 2)
	requireSuccess(t, book.RemoveEvaluation("anna", "review"))
	assertTotal(official, 0, 1)
}

func TestCalculatorNormalizesAcrossLargeWeights(t *testing.T) {
	book := domain.NewGradebook()
	requireSuccess(t, book.AddStudent("anna", "Анна"))
	requireSuccess(t, book.AddCriterion("x", "X", math.MaxFloat64, math.MaxFloat64))
	requireSuccess(t, book.AddCriterion("y", "Y", 20, math.MaxFloat64))
	requireSuccess(t, book.SetEvaluation("anna", "x", math.MaxFloat64, ""))
	requireSuccess(t, book.SetEvaluation("anna", "y", 10, ""))
	result, err := domain.NewCalculator().Calculate(book, "anna", domain.AllCriteria{})
	requireSuccess(t, err)
	if result.Points != 7.5 {
		t.Fatalf("total = %g, want 7.5", result.Points)
	}
}
