package export

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/ananas53/kpo-hw/hw-2/internal/domain"
)

type CSV struct{}

func NewCSV() *CSV { return &CSV{} }

func (*CSV) Export(path string, totals []domain.Total) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)

	if err != nil {
		return fmt.Errorf("создать CSV: %w", err)
	}

	writer := csv.NewWriter(file)
	writeErr := writeTotals(writer, totals)
	writer.Flush()
	err = errors.Join(writeErr, writer.Error(), file.Close())

	if err != nil {
		os.Remove(path)
		return fmt.Errorf("записать CSV: %w", err)
	}

	return nil
}

func writeTotals(writer *csv.Writer, totals []domain.Total) error {
	err := writer.Write([]string{"student_id", "student_name", "total", "graded_criteria", "criteria"})

	if err != nil {
		return err
	}

	for _, total := range totals {
		err := writer.Write([]string{
			total.Student.ID(),
			total.Student.Name(),
			strconv.FormatFloat(total.Points, 'f', 2, 64),
			strconv.Itoa(total.Graded),
			strconv.Itoa(total.Criteria),
		})

		if err != nil {
			return err
		}
	}

	return nil
}
