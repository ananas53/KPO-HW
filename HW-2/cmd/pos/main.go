package main

import (
	"fmt"
	"io"
	"os"

	"github.com/ananas53/kpo-hw/hw-2/internal/adapter/console"
	"github.com/ananas53/kpo-hw/hw-2/internal/adapter/export"
	"github.com/ananas53/kpo-hw/hw-2/internal/adapter/storage"
	"github.com/ananas53/kpo-hw/hw-2/internal/application"
	"github.com/ananas53/kpo-hw/hw-2/internal/domain"
	"go.uber.org/dig"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", err)
		os.Exit(1)
	}
}

func run(args []string, output io.Writer) error {
	options, err := console.ParseOptions(args)
	if err != nil {
		return err
	}
	container := dig.New()
	constructors := []any{
		func() application.Repository {
			if options.IsDemo() {
				return storage.NewMemory()
			}
			return storage.NewJSON(options.Database)
		},
		func() application.TotalsExporter { return export.NewCSV() },
		func() io.Writer { return output },
		domain.NewCalculator,
		application.NewPOS,
		console.New,
	}
	for _, constructor := range constructors {
		if err := container.Provide(constructor); err != nil {
			return fmt.Errorf("собрать зависимости ПОС: %w", err)
		}
	}
	return container.Invoke(func(cli *console.CLI) error {
		return cli.Run(options.Args)
	})
}
