package console

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/ananas53/kpo-hw/hw-2/internal/adapter/gradebookpath"
	"github.com/ananas53/kpo-hw/hw-2/internal/application"
	"github.com/ananas53/kpo-hw/hw-2/internal/domain"
)

type Options struct {
	Database string
	Args     []string
}

func ParseOptions(args []string) (Options, error) {
	flags := flag.NewFlagSet("pos", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	database := flags.String("db", "pos.json", "путь к ведомости")
	err := flags.Parse(args)
	if errors.Is(err, flag.ErrHelp) {
		return Options{Database: *database, Args: []string{"help"}}, nil
	}
	if err != nil {
		return Options{}, err
	}
	if strings.TrimSpace(*database) == "" {
		return Options{}, errors.New("укажите путь к ведомости в -db")
	}
	options := Options{Database: *database, Args: flags.Args()}
	var report string
	switch {
	case options.IsDemo() && len(options.Args) <= 2:
		report = demoPath(options.Args[1:])
	case len(options.Args) == 2 && options.Args[0] == "export":
		report = options.Args[1]
	default:
		return options, nil
	}
	if err := gradebookpath.CheckReportPath(options.Database, report); err != nil {
		return Options{}, err
	}
	return options, nil
}

func (o Options) IsDemo() bool {
	return len(o.Args) > 0 && o.Args[0] == "demo"
}

type CLI struct {
	pos    *application.POS
	output io.Writer
}

func New(pos *application.POS, output io.Writer) *CLI {
	return &CLI{pos: pos, output: output}
}

type command struct {
	usage       string
	description string
	minArgs     int
	maxArgs     int
	run         func([]string) error
}

func (c *CLI) commands() []command {
	return []command{
		{"add-student ID ИМЯ", "добавить студента", 2, 2, c.addStudent},
		{"students", "показать студентов", 0, 0, c.students},
		{"add-criterion ID НАЗВАНИЕ МАКСИМУМ ВЕС", "добавить критерий", 4, 4, c.addCriterion},
		{"criteria", "показать рубрикатор", 0, 0, c.criteria},
		{"grade СТУДЕНТ КРИТЕРИЙ БАЛЛ [КОММЕНТАРИЙ]", "выставить или исправить оценку", 3, 4, c.grade},
		{"grades СТУДЕНТ", "показать оценки", 1, 1, c.grades},
		{"remove-grade СТУДЕНТ КРИТЕРИЙ", "отменить оценку", 2, 2, c.removeGrade},
		{"total СТУДЕНТ", "посчитать итог по всему рубрикатору", 1, 1, c.total},
		{"progress СТУДЕНТ", "посчитать результат по выставленным оценкам", 1, 1, c.progress},
		{"totals", "показать итоговую ведомость", 0, 0, c.totals},
		{"export ФАЙЛ.csv", "выгрузить итоговую ведомость", 1, 1, c.export},
		{"demo [ФАЙЛ.csv]", "показать сквозной сценарий", 0, 1, c.demo},
		{"help", "показать список команд", 0, 0, func([]string) error { return c.help() }},
	}
}

func (c *CLI) Run(args []string) error {
	if len(args) == 0 {
		return c.help()
	}
	for _, cmd := range c.commands() {
		name, _, _ := strings.Cut(cmd.usage, " ")
		if args[0] != name {
			continue
		}
		params := args[1:]
		if len(params) < cmd.minArgs || len(params) > cmd.maxArgs {
			return fmt.Errorf("использование: pos %s", cmd.usage)
		}
		return cmd.run(params)
	}
	return fmt.Errorf("неизвестная команда %q. Список команд: pos help", args[0])
}

func (c *CLI) help() error {
	table := tabwriter.NewWriter(c.output, 0, 0, 2, ' ', 0)
	fmt.Fprintln(table, "ПОС — преподавательская оценка студентов")
	fmt.Fprintln(table, "Использование: pos [-db ФАЙЛ.json] КОМАНДА [АРГУМЕНТЫ]")
	for _, cmd := range c.commands() {
		fmt.Fprintf(table, "  %s\t%s\n", cmd.usage, cmd.description)
	}
	return table.Flush()
}

func (c *CLI) message(format string, args ...any) error {
	_, err := fmt.Fprintf(c.output, format+"\n", args...)
	return err
}

func number(value, field string) (float64, error) {
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %q не является числом", field, value)
	}
	return result, nil
}

func (c *CLI) addStudent(args []string) error {
	if err := c.pos.AddStudent(args[0], args[1]); err != nil {
		return err
	}
	return c.message("Студент %q добавлен.", args[0])
}

func (c *CLI) students(_ []string) error {
	students, err := c.pos.Students()
	if err != nil {
		return err
	}
	if len(students) == 0 {
		return c.message("Студентов пока нет.")
	}
	for _, student := range students {
		if err := c.message("%q — %q", student.ID(), student.Name()); err != nil {
			return err
		}
	}
	return nil
}

func (c *CLI) addCriterion(args []string) error {
	maxScore, err := number(args[2], "максимум")
	if err != nil {
		return err
	}
	weight, err := number(args[3], "вес")
	if err != nil {
		return err
	}
	if err := c.pos.AddCriterion(args[0], args[1], maxScore, weight); err != nil {
		return err
	}
	return c.message("Критерий %q добавлен.", args[0])
}

func (c *CLI) criteria(_ []string) error {
	criteria, err := c.pos.Criteria()
	if err != nil {
		return err
	}
	if len(criteria) == 0 {
		return c.message("Рубрикатор пока пуст.")
	}
	for _, criterion := range criteria {
		err := c.message("%q — %q (максимум %g, вес %g)",
			criterion.ID(), criterion.Name(), criterion.MaxScore(), criterion.Weight())
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CLI) grade(args []string) error {
	points, err := number(args[2], "балл")
	if err != nil {
		return err
	}
	comment := ""
	if len(args) == 4 {
		comment = args[3]
	}
	if err := c.pos.SetEvaluation(args[0], args[1], points, comment); err != nil {
		return err
	}
	return c.message("Оценка сохранена.")
}

func (c *CLI) grades(args []string) error {
	evaluations, err := c.pos.Evaluations(args[0])
	if err != nil {
		return err
	}
	if len(evaluations) == 0 {
		return c.message("Оценок пока нет.")
	}
	for _, evaluation := range evaluations {
		if evaluation.Comment() == "" {
			err = c.message("%q — %g", evaluation.CriterionID(), evaluation.Score().Value())
		} else {
			err = c.message("%q — %g (%q)", evaluation.CriterionID(), evaluation.Score().Value(), evaluation.Comment())
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CLI) removeGrade(args []string) error {
	if err := c.pos.RemoveEvaluation(args[0], args[1]); err != nil {
		return err
	}
	return c.message("Оценка отменена.")
}

func (c *CLI) total(args []string) error {
	total, err := c.pos.Total(args[0])
	if err != nil {
		return err
	}
	return c.printTotal("Итог", total)
}

func (c *CLI) progress(args []string) error {
	total, err := c.pos.Progress(args[0])
	if err != nil {
		return err
	}
	return c.printTotal("Прогресс", total)
}

func (c *CLI) totals(_ []string) error {
	totals, err := c.pos.Totals()
	if err != nil {
		return err
	}
	if len(totals) == 0 {
		return c.message("Студентов пока нет.")
	}
	for _, total := range totals {
		if err := c.printTotal("Итог", total); err != nil {
			return err
		}
	}
	return nil
}

func (c *CLI) printTotal(label string, total domain.Total) error {
	return c.message("%s %q [%s]: %.2f/%d, оценено %d/%d критериев",
		label, total.Student.Name(), total.Student.ID(), total.Points, domain.TotalMax, total.Graded, total.Criteria)
}

func (c *CLI) export(args []string) error {
	if err := c.pos.ExportTotals(args[0]); err != nil {
		return err
	}
	return c.message("Итоги выгружены в %q.", args[0])
}
