/*
Copyright © 2025 Ralph Schindler <ralph@ralphschindler.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/project-actions-org/command-runner/internal/command"
	"github.com/spf13/cobra"
)

var commandSpecs []struct {
	spec *command.Command
	name string
}
var rootCmd *cobra.Command
var verbose bool

func Execute() {
	err := rootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.EnableCommandSorting = false

	// the name of the called program, stripping off any path information
	programName := filepath.Base(os.Args[0])

	rootCmd = &cobra.Command{
		Use:   programName,
		Short: "Project Actions Command Runner",
		Long:  `Project Actions Command Runner Description`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true, // hides cmd
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "increase output verbosity")

	projectActionsDir := "./project"

	if os.Getenv("PROJECT_ACTIONS_DIR") != "" {
		projectActionsDir = os.Getenv("PROJECT_ACTIONS_DIR")
	}

	isOutsideContainer := true

	// if /.dockerenv, or /run/.containerenv exists, we are inside a container
	if _, err := os.Stat("/.dockerenv"); err == nil {
		isOutsideContainer = false
	} else if _, err := os.Stat("/run/.containerenv"); err == nil {
		isOutsideContainer = false
	}

	// dynamically add commands from yml files in the "commands" subdirectory
	files, err := os.ReadDir(projectActionsDir + "/commands")

	if err != nil {
		// fmt.Println("Could not read commands directory:", err)

		return
	}

	for _, f := range files {
		if f.IsDir() || f.Name()[len(f.Name())-4:] != ".yml" {
			continue
		}

		commandSpec, err := command.ParseCommandFile(projectActionsDir + "/commands/" + f.Name())

		if err != nil {
			fmt.Println("Could not parse command file:", err)
			return
		}

		baseName := f.Name()

		if ext := filepath.Ext(baseName); ext == ".yml" {
			baseName = baseName[:len(baseName)-len(ext)]
		}

		commandSpecs = append(commandSpecs, struct {
			spec *command.Command
			name string
		}{
			spec: commandSpec,
			name: baseName,
		})
	}

	// Sort commands by order value
	sort.Slice(commandSpecs, func(i, j int) bool {
		return commandSpecs[i].spec.Help.Order < commandSpecs[j].spec.Help.Order
	})

	// Now add them in sorted order
	for _, cmdSpec := range commandSpecs {
		cmd := &cobra.Command{
			Use:   cmdSpec.name,
			Short: cmdSpec.spec.Help.Short,
			Long:  cmdSpec.spec.Help.Long,
			Run: func(cmd *cobra.Command, args []string) {
				command.RunCommand(cmdSpec.name, cmdSpec.spec, isOutsideContainer)
			},
		}

		rootCmd.AddCommand(cmd)
	}
}
