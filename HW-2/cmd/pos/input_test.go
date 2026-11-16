package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInvalidUTF8IDLeavesGradebookReadable(t *testing.T) {
	database := filepath.Join(t.TempDir(), "grades.json")
	command := func(args ...string) error {
		return run(append([]string{"-db", database}, args...), io.Discard)
	}
	if err := command("add-student", "id-�", "Анна"); err != nil {
		t.Fatal(err)
	}
	if err := command("add-criterion", "id-�", "Практика", "10", "1"); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(database)
	if err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add-student", "id-\xff", "Иван"},
		{"add-criterion", "id-\xff", "Защита", "10", "1"},
	} {
		if err := command(args...); err == nil || !strings.Contains(err.Error(), "UTF-8") {
			t.Errorf("%s: got %v, want invalid UTF-8 error", args[0], err)
		}
	}
	after, err := os.ReadFile(database)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Error("invalid ID changed the saved gradebook")
	}
	if err := command("students"); err != nil {
		t.Fatalf("reading the gradebook after rejected IDs: %v", err)
	}
}

func TestDemoRejectsGradebookDestination(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "data")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(root, "alias")
	if err := os.Symlink("data", alias); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name         string
		databaseName string
		dir          string
		file         string
	}{
		{"same path", "pos.json", directory, "pos.json"},
		{"directory component", "pos.json", directory + string(os.PathSeparator) + ".", "pos.json"},
		{"symlink to directory", "pos.json", alias, "pos.json"},
		{"different letter case", "pos.json", directory, "POS.JSON"},
		{"normalization and case", "caf\u00e9.json", directory, "CAFE\u0301.JSON"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			database := filepath.Join(directory, tc.databaseName)
			report := tc.dir + string(os.PathSeparator) + tc.file
			var output bytes.Buffer
			err := run([]string{"-db", database, "demo", report}, &output)
			if err == nil || !strings.Contains(err.Error(), "совпадает с путём ведомости") {
				t.Errorf("demo error = %v, want a gradebook-path conflict", err)
			}
			if output.Len() != 0 {
				t.Error("demo started with a conflicting destination")
			}
			if _, err := os.Lstat(report); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("demo created a file at the reserved path: %v", err)
				if err == nil {
					if removeErr := os.Remove(report); removeErr != nil {
						t.Fatal(removeErr)
					}
				}
			}
		})
	}
}

func TestDemoRejectsDanglingGradebookLink(t *testing.T) {
	directory := t.TempDir()
	database := filepath.Join(directory, "alias.json")
	target := filepath.Join(directory, "grades.json")
	if err := os.Symlink("grades.json", database); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-db", database, "demo", target}, io.Discard); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("demo error = %v, want missing gradebook target", err)
	}
	if _, err := os.Lstat(target); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("demo created the missing gradebook target: %v", err)
	}
}

func TestCSVCommandsRejectGradebookLock(t *testing.T) {
	for _, name := range []string{"export", "demo"} {
		t.Run(name, func(t *testing.T) {
			directory := t.TempDir()
			database := filepath.Join(directory, "grades.json")
			command := func(db string, args ...string) error {
				return run(append([]string{"-db", db}, args...), io.Discard)
			}
			if err := command(database, "add-student", "anna", "Анна"); err != nil {
				t.Fatal(err)
			}
			if err := command(database, "add-criterion", "practice", "Практика", "10", "1"); err != nil {
				t.Fatal(err)
			}
			alias := filepath.Join(directory, "alias.json")
			if err := os.Symlink("grades.json", alias); err != nil {
				t.Fatal(err)
			}
			lock := database + ".lock"
			if err := command(alias, name, lock); err == nil || !strings.Contains(err.Error(), "путём блокировки ведомости") {
				t.Errorf("%s error = %v, want a gradebook-lock conflict", name, err)
			}
			if _, err := os.Lstat(lock); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("export created the gradebook lock path: %v", err)
			}
			if err := command(database, "grade", "anna", "practice", "8"); err != nil {
				t.Fatalf("gradebook is no longer writable: %v", err)
			}
		})
	}
}

func TestInvalidUTF8TextLeavesGradebookUnchanged(t *testing.T) {
	database := filepath.Join(t.TempDir(), "grades.json")
	command := func(args ...string) error {
		return run(append([]string{"-db", database}, args...), io.Discard)
	}
	for _, args := range [][]string{
		{"add-student", "anna", "Анна"},
		{"add-criterion", "practice", "Практика", "10", "1"},
		{"grade", "anna", "practice", "7", "Готово"},
	} {
		if err := command(args...); err != nil {
			t.Fatal(err)
		}
	}
	before, err := os.ReadFile(database)
	if err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add-student", "new", "имя-\xff"},
		{"add-criterion", "new", "название-\xff", "10", "1"},
		{"grade", "anna", "practice", "8", "комментарий-\xff"},
	} {
		if err := command(args...); err == nil || !strings.Contains(err.Error(), "UTF-8") {
			t.Errorf("%s: got %v, want invalid UTF-8 error", args[0], err)
		}
	}
	after, err := os.ReadFile(database)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Error("invalid text changed the saved gradebook")
	}
}
