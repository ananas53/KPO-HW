package domain

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	ErrStudentNotFound    = errors.New("студент не найден")
	ErrCriterionNotFound  = errors.New("критерий отсутствует в рубрикаторе")
	ErrEvaluationNotFound = errors.New("оценка не найдена")
)

type Student struct {
	id   string
	name string
}

func (s Student) ID() string   { return s.id }
func (s Student) Name() string { return s.name }

type Criterion struct {
	id       string
	name     string
	maxScore float64
	weight   float64
}

func (c Criterion) ID() string        { return c.id }
func (c Criterion) Name() string      { return c.name }
func (c Criterion) MaxScore() float64 { return c.maxScore }
func (c Criterion) Weight() float64   { return c.weight }

// Score is a value object. Its upper bound is defined by the evaluation's criterion.
type Score struct {
	value float64
}

func (s Score) Value() float64 { return s.value }

type Evaluation struct {
	studentID   string
	criterionID string
	score       Score
	comment     string
}

func (e Evaluation) StudentID() string   { return e.studentID }
func (e Evaluation) CriterionID() string { return e.criterionID }
func (e Evaluation) Score() Score        { return e.score }
func (e Evaluation) Comment() string     { return e.comment }

func identity(id, name string) (string, string, error) {
	if !utf8.ValidString(id) {
		return "", "", errors.New("идентификатор должен быть в кодировке UTF-8")
	}

	if !utf8.ValidString(name) {
		return "", "", errors.New("имя или название должно быть в кодировке UTF-8")
	}

	id, name = strings.TrimSpace(id), strings.TrimSpace(name)
	if id == "" || strings.IndexFunc(id, unicode.IsSpace) >= 0 {
		return "", "", errors.New("укажите непустой идентификатор без пробелов")
	}

	if name == "" {
		return "", "", errors.New("название или имя не может быть пустым")
	}

	return id, name, nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func (b *Gradebook) newEvaluation(studentID, criterionID string, points float64, comment string) (Evaluation, error) {
	_, err := b.student(studentID)

	if err != nil {
		return Evaluation{}, err
	}

	criterion, ok := b.rubric[criterionID]

	if !ok {
		return Evaluation{}, fmt.Errorf("%w: %q", ErrCriterionNotFound, criterionID)
	}

	if !finite(points) || points < 0 || points > criterion.maxScore {
		return Evaluation{}, fmt.Errorf("балл по критерию %q должен быть в диапазоне [0, %g]", criterionID, criterion.maxScore)
	}

	if !utf8.ValidString(comment) {
		return Evaluation{}, errors.New("комментарий должен быть в кодировке UTF-8")
	}

	return Evaluation{
		studentID:   studentID,
		criterionID: criterionID,
		score:       Score{value: points},
		comment:     strings.TrimSpace(comment),
	}, nil
}
