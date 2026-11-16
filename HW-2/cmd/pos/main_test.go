package main

import (
	"bytes"
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCommandWorkflowWithPersistentGradebook(t *testing.T) {
	directory := t.TempDir()
	database := filepath.Join(directory, "pos.json")
	report := filepath.Join(directory, "totals.csv")
	name := "Анна, Орлова"
	command := func(args ...string) (string, error) {
		var output bytes.Buffer
		err := run(append([]string{"-db", database}, args...), &output)
		return output.String(), err
	}
	succeed := func(args ...string) string {
		t.Helper()
		output, err := command(args...)
		if err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		return output
	}
	succeed("add-student", "anna", name)
	succeed("add-criterion", "practice", "Практика", "20", "3")
	succeed("add-criterion", "review", "Защита", "10", "2")
	succeed("grade", "anna", "practice", "16", "готово")
	if _, err := command("grade", "anna", "practice", "21"); err == nil {
		t.Fatal("CLI accepted a grade above the criterion maximum")
	}
	if actual := succeed("grades", "anna"); !strings.Contains(actual, "16") || !strings.Contains(actual, "готово") {
		t.Fatalf("valid grade was not preserved: %s", actual)
	}
	if actual := succeed("progress", "anna"); !strings.Contains(actual, "8.00/10") {
		t.Fatalf("wrong progress: %s", actual)
	}
	succeed("grade", "anna", "review", "9")
	succeed("grade", "anna", "review", "8")
	if actual := succeed("total", "anna"); !strings.Contains(actual, "8.00/10") {
		t.Fatalf("wrong total after correction: %s", actual)
	}
	succeed("remove-grade", "anna", "review")
	succeed("export", report)
	file, err := os.Open(report)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"student_id", "student_name", "total", "graded_criteria", "criteria"},
		{"anna", name, "4.80", "1", "2"},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("CSV = %v, want %v", rows, want)
	}
	if _, err := command("export", database); err == nil {
		t.Fatal("CSV export overwrote the gradebook")
	}
	if actual := succeed("students"); !strings.Contains(actual, name) {
		t.Fatalf("gradebook was not preserved: %s", actual)
	}
}

func TestDemoUsesSeparateGradebook(t *testing.T) {
	directory := t.TempDir()
	database := filepath.Join(directory, "actual.json")
	report := filepath.Join(directory, "demo.csv")
	original := []byte("saved gradebook")
	if err := os.WriteFile(database, original, 0o600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := run([]string{"-db", database, "demo", report}, &output); err != nil {
		t.Fatal(err)
	}
	reportData, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []string{"anna,Анна Орлова,8.40,2,2", "ivan,Иван Волков,5.40,1,2"} {
		if !bytes.Contains(reportData, []byte(row)) {
			t.Fatalf("demo CSV missing %q: %s", row, reportData)
		}
	}
	databaseData, err := os.ReadFile(database)
	if err != nil || !bytes.Equal(databaseData, original) {
		t.Fatalf("demo changed the gradebook: got %q, error %v", databaseData, err)
	}
}
