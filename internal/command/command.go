package command

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

type Command struct {
	Help    Help   `yaml:"help"`
	Context string `yaml:"context"`
	Steps   []Step `yaml:"steps"`
}

type Help struct {
	Short string `yaml:"short"`
	Long  string `yaml:"long"`
	Order int    `yaml:"order"`
}

type Step struct {
	Action  *Action `yaml:"action,omitempty"`
	Echo    *string `yaml:"echo,omitempty"`
	Command *string `yaml:"command,omitempty"`
	Run     *string `yaml:"run,omitempty"`
}

type Action struct {
	Name string `yaml:"action,omitempty"`
}

func ParseCommandFile(path string) (*Command, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	var command Command
	decoder := yaml.NewDecoder(file)

	if err := decoder.Decode(&command); err != nil {
		return nil, err
	}

	return &command, nil
}

func RunCommand(commandName string, command *Command, isOutsideContainer bool) error {
	var runContext string = "shell"

	if command.Context == "outside-container" && !isOutsideContainer {
		fmt.Println(commandName, "must be run outside a container")

		return nil
	}

	// if command.Context starts with "inside-container", we need to run the command inside a container
	if strings.HasPrefix(command.Context, "inside-container") {
		if isOutsideContainer {
			// @todo set the runContext to the first service in the docker-compose.yml file

			// if command.Context does not have a service specified, return an error
			if !strings.Contains(command.Context, ":") {
				return fmt.Errorf("command.Context does not have a service specified")
			}

			runContext = strings.Split(command.Context, ":")[1]
		}
	}

	for _, step := range command.Steps {
		if step.Action != nil {
			// this should execute the provided action
			fmt.Println("Action not implemented yet:", step.Action.Name)
			continue
		}

		if step.Run != nil {
			runInContext(*step.Run, runContext)
			continue
		}

		if step.Echo != nil {
			fmt.Println(*step.Echo)
			continue
		}

		if step.Command != nil {
			fmt.Println("Calling other commands is not implemented yet:", *step.Command)
			continue
		}
	}

	return nil
}

func runInContext(shellCode string, runContext string) error {
	var args []string

	if runContext == "shell" {
		args = []string{"sh", "-c", shellCode}
	} else {
		args = []string{"docker-compose", "exec", runContext, "sh", "-c", shellCode}
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
