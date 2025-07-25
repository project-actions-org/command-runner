/*
Copyright © 2025 Ralph Schindler <ralph@ralphschindler.com>
*/
package cmd

import (
	"os"
	"path/filepath"
	"sort"

	internal "github.com/project-actions-org/command-runner/internal"
	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var commandSpecs []struct {
	spec *internal.Command
	name string
}
var debug bool = os.Getenv("DEBUG") == "true"
var rootCmd *cobra.Command
var verbose bool

func Execute() {
	internal.Logger.Debug("Executing command")

	err := rootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}

func init() {
	projectScriptName := os.Getenv("PROJECT_SCRIPT_NAME")

	if projectScriptName == "" {
		projectScriptName = filepath.Base(os.Args[0])
	}

	// Disable sorting of commands, we want to keep the order of the files specified by their own ordering preferences
	cobra.EnableCommandSorting = false

	rootCmd = &cobra.Command{
		Use:   projectScriptName,
		Short: "Project Actions Command Runner",
		Long:  `Project Actions Command Runner Description`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !debug && verbose {
				internal.Logger.SetLevel(logrus.InfoLevel)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "increase output verbosity")

	// cwd + /.project
	projectActionsDir := filepath.Join(os.Getenv("PWD"), ".project")

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

	internal.Logger.Debug("Is outside container?", isOutsideContainer)

	// dynamically add commands from yml files in the "commands" subdirectory
	files, err := os.ReadDir(projectActionsDir + "/commands")

	if err != nil {
		internal.Logger.Error("Could not read commands directory:", err)

		return
	}

	internal.Logger.Debug("Looking for files in ", projectActionsDir+"/commands")

	for _, f := range files {

		// skip directories
		if f.IsDir() {
			continue
		}

		// skip files that are not .yaml or .yml
		if f.Name()[len(f.Name())-5:] != ".yaml" && f.Name()[len(f.Name())-4:] != ".yml" {
			continue
		}

		commandSpec, err := internal.ParseCommandFile(projectActionsDir + "/commands/" + f.Name())

		if err != nil {
			internal.Logger.Error("Could not parse command file:", err)

			return
		}

		baseName := f.Name()

		if ext := filepath.Ext(baseName); ext == ".yml" || ext == ".yaml" {
			baseName = baseName[:len(baseName)-len(ext)]
		}

		commandSpecs = append(commandSpecs, struct {
			spec *internal.Command
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
				internal.RunCommand(cmdSpec.name, cmdSpec.spec, isOutsideContainer)
			},
		}

		rootCmd.AddCommand(cmd)
	}
}
