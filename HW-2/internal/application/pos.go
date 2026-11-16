package application

import "github.com/ananas53/kpo-hw/hw-2/internal/domain"

// Repository stores the gradebook.
type Repository interface {
	// Load returns a snapshot that can be edited independently of the repository.
	Load() (*domain.Gradebook, error)

	// Update applies change to a private draft. A callback error leaves the
	// stored gradebook unchanged. After success, the draft is independent
	// of the stored state.
	Update(change func(*domain.Gradebook) error) error
}

type TotalsExporter interface {
	Export(path string, totals []domain.Total) error
}

// POS is the facade used by the console and other inbound adapters.
type POS struct {
	repository Repository
	calculator *domain.Calculator
	exporter   TotalsExporter
}

func NewPOS(repository Repository, calculator *domain.Calculator, exporter TotalsExporter) *POS {
	return &POS{repository: repository, calculator: calculator, exporter: exporter}
}

func (p *POS) AddStudent(id, name string) error {
	return p.repository.Update(func(book *domain.Gradebook) error {
		return book.AddStudent(id, name)
	})
}

func (p *POS) AddCriterion(id, name string, maxScore, weight float64) error {
	return p.repository.Update(func(book *domain.Gradebook) error {
		return book.AddCriterion(id, name, maxScore, weight)
	})
}

func (p *POS) SetEvaluation(studentID, criterionID string, points float64, comment string) error {
	return p.repository.Update(func(book *domain.Gradebook) error {
		return book.SetEvaluation(studentID, criterionID, points, comment)
	})
}

func (p *POS) RemoveEvaluation(studentID, criterionID string) error {
	return p.repository.Update(func(book *domain.Gradebook) error {
		return book.RemoveEvaluation(studentID, criterionID)
	})
}

func (p *POS) Students() ([]domain.Student, error) {
	book, err := p.repository.Load()

	if err != nil {
		return nil, err
	}

	return book.Students(), nil
}

func (p *POS) Criteria() ([]domain.Criterion, error) {
	book, err := p.repository.Load()

	if err != nil {
		return nil, err
	}

	return book.Criteria(), nil
}

func (p *POS) Evaluations(studentID string) ([]domain.Evaluation, error) {
	book, err := p.repository.Load()

	if err != nil {
		return nil, err
	}

	return book.Evaluations(studentID)
}

func (p *POS) Total(studentID string) (domain.Total, error) {
	return p.total(studentID, domain.AllCriteria{})
}

func (p *POS) Progress(studentID string) (domain.Total, error) {
	return p.total(studentID, domain.GradedOnly{})
}

func (p *POS) total(studentID string, strategy domain.ScoringStrategy) (domain.Total, error) {
	book, err := p.repository.Load()

	if err != nil {
		return domain.Total{}, err
	}

	return p.calculator.Calculate(book, studentID, strategy)
}

func (p *POS) Totals() ([]domain.Total, error) {
	book, err := p.repository.Load()

	if err != nil {
		return nil, err
	}

	if len(book.Criteria()) == 0 {
		return nil, domain.ErrEmptyRubric
	}

	students := book.Students()
	totals := make([]domain.Total, 0, len(students))

	for _, student := range students {
		total, err := p.calculator.Calculate(book, student.ID(), domain.AllCriteria{})

		if err != nil {
			return nil, err
		}

		totals = append(totals, total)
	}

	return totals, nil
}

func (p *POS) ExportTotals(path string) error {
	totals, err := p.Totals()

	if err != nil {
		return err
	}

	return p.exporter.Export(path, totals)
}
