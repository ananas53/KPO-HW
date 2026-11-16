package gradebookpath

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// Resolve returns the gradebook's target and the lock path beside it.
func Resolve(path string) (string, string, error) {
	info, err := os.Lstat(path)

	if errors.Is(err, os.ErrNotExist) {
		return path, path + ".lock", nil
	}

	if err != nil {
		return "", "", err
	}

	target, err := filepath.EvalSymlinks(path)

	if info.Mode()&os.ModeSymlink != 0 && errors.Is(err, os.ErrNotExist) {
		return "", "", fmt.Errorf("цель символической ссылки отсутствует: %w", err)
	}

	if err != nil {
		return "", "", err
	}

	return target, target + ".lock", nil
}

// CheckReportPath prevents an export from occupying the gradebook or its lock.
func CheckReportPath(database, report string) error {
	target, lock, err := Resolve(database)

	if err != nil {
		return fmt.Errorf("проверить путь ведомости %q: %w", database, err)
	}

	databaseDir, databaseName := filepath.Split(target)
	_, lockName := filepath.Split(lock)
	reportDir, reportName := filepath.Split(report)
	// Treat case and Unicode normalization variants as aliases on every volume.
	reportName = norm.NFC.String(reportName)
	var conflict error

	switch {
	case strings.EqualFold(norm.NFC.String(databaseName), reportName):
		conflict = fmt.Errorf("путь для CSV %q совпадает с путём ведомости %q", report, database)
	case strings.EqualFold(norm.NFC.String(lockName), reportName):
		conflict = fmt.Errorf("путь для CSV %q совпадает с путём блокировки ведомости %q", report, lock)
	default:
		return nil
	}

	databaseParent, err := directoryInfo(databaseDir)

	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("проверить каталог ведомости %q: %w", database, err)
	}

	reportParent, err := directoryInfo(reportDir)

	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("проверить каталог CSV %q: %w", report, err)
	}

	if os.SameFile(databaseParent, reportParent) {
		return conflict
	}

	return nil
}

func directoryInfo(path string) (os.FileInfo, error) {
	if path == "" {
		path = "."
	}

	return os.Stat(path)
}
