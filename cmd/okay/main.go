package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NathanFirmo/okay/internal/parser"
	"github.com/NathanFirmo/okay/internal/runner"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "okay",
		Short: "Declarative integration testing tool",
		Long:  "A tool to run declarative integration tests defined in YAML files.",
		Run: func(cmd *cobra.Command, args []string) {
			err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".test.yaml") {
					return nil
				}

				filepath := strings.ToLower(info.Name())

				config, err := parser.ParseSuite(filepath)
				if err != nil {
					fmt.Printf("Error parsing config: %v\n", err)
					os.Exit(1)
				}
				r, err := runner.NewRunner(config, filepath)
				if err != nil {
					fmt.Printf("Error creating runner: %v\n", err)
					os.Exit(1)
				}
				err = r.Run()
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				return nil
			})

			if err != nil {
				fmt.Printf("Error walking directory: %v\n", err)
			}
		},
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
