package domain

import "errors"

const TotalMax = 10

var (
	ErrEmptyRubric = errors.New("рубрикатор пуст")
	ErrNoGrades    = errors.New("у студента пока нет оценок")
)

// ScoringStrategy decides whether a rubric criterion contributes to the total weight.
type ScoringStrategy interface {
	Include(graded bool) bool
}

// AllCriteria counts missing grades as zero in the official total.
type AllCriteria struct{}

func (AllCriteria) Include(bool) bool { return true }

// GradedOnly shows progress against the criteria already graded.
type GradedOnly struct{}

func (GradedOnly) Include(graded bool) bool { return graded }

type Total struct {
	Student  Student
	Points   float64
	Graded   int
	Criteria int
}

// Calculator calculates a weighted total from the current state of a gradebook.
type Calculator struct{}

func NewCalculator() *Calculator { return &Calculator{} }

func (*Calculator) Calculate(book *Gradebook, studentID string, strategy ScoringStrategy) (Total, error) {
	student, err := book.student(studentID)

	if err != nil {
		return Total{}, err
	}

	criteria := book.Criteria()

	if len(criteria) == 0 {
		return Total{}, ErrEmptyRubric
	}

	evaluations := book.evaluations[studentID]
	var largestWeight float64

	for _, criterion := range criteria {
		_, graded := evaluations[criterion.id]

		if strategy.Include(graded) {
			largestWeight = max(largestWeight, criterion.weight)
		}
	}

	if largestWeight == 0 {
		return Total{}, ErrNoGrades
	}

	// Scaling by the largest weight keeps the weighted sum finite.
	var earned, possible float64

	for _, criterion := range criteria {
		evaluation, graded := evaluations[criterion.id]

		if !strategy.Include(graded) {
			continue
		}

		weight := criterion.weight / largestWeight
		possible += weight
		if graded {
			earned += weight * (evaluation.score.value / criterion.maxScore)
		}
	}

	return Total{
		Student:  student,
		Points:   TotalMax * (earned / possible),
		Graded:   len(evaluations),
		Criteria: len(criteria),
	}, nil
}
