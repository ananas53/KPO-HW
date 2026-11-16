package console

import (
	"fmt"
	"strings"
)

func demoPath(args []string) string {
	if len(args) == 1 {
		return args[0]
	}
	return "demo-totals.csv"
}

func (c *CLI) demo(args []string) error {
	path := demoPath(args)
	scenario := [][]string{
		{"add-student", "anna", "Анна Орлова"},
		{"add-student", "ivan", "Иван Волков"},
		{"add-criterion", "practice", "Практика", "20", "3"},
		{"add-criterion", "review", "Защита", "10", "2"},
		{"students"},
		{"criteria"},
		{"grade", "anna", "practice", "12"},
		{"grade", "anna", "practice", "16", "Исправленный балл"},
		{"grade", "anna", "review", "9"},
		{"grade", "ivan", "practice", "18"},
		{"grade", "ivan", "review", "7"},
		{"remove-grade", "ivan", "review"},
		{"grades", "anna"},
		{"progress", "ivan"},
		{"totals"},
		{"export", path},
	}
	for _, step := range scenario {
		if err := c.message("\n> %s", strings.Join(step, " ")); err != nil {
			return err
		}
		if err := c.Run(step); err != nil {
			return fmt.Errorf("демо, команда %s: %w", step[0], err)
		}
	}
	return nil
}
