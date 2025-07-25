package internal

import (
	"fmt"
	"os"
	"os/exec"
)

type ActionFunc func() error

var actionMap = map[string]ActionFunc{
	// compose up variations
	"compose-up":        ComposeWithArgument("up", "-d"),
	"docker-compose-up": ComposeWithArgument("up", "-d"),
	"podman-compose-up": ComposeWithArgument("up", "-d"),
	// compose down variations
	"compose-down":        ComposeWithArgument("down"),
	"docker-compose-down": ComposeWithArgument("down"),
	"podman-compose-down": ComposeWithArgument("down"),
	// compose stop variations
	"compose-stop":        ComposeWithArgument("stop"),
	"docker-compose-stop": ComposeWithArgument("stop"),
	"podman-compose-stop": ComposeWithArgument("stop"),
}

func ComposeWithArgument(args ...string) ActionFunc {
	return func() error {
		path, err := getComposePath()

		if err != nil {
			Logger.Error("Failed to get compose path:", err)
			return err
		}

		Logger.Debug("Compose path:", append(path[0:1], append(path[1:], args...)...))

		cmd := exec.Command(path[0], append(path[1:], args...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}

		return nil
	}
}

func getComposePath() ([]string, error) {
	// Split commands into executable and arguments
	commands := []struct {
		args []string
	}{
		{[]string{"docker", "compose"}},
		{[]string{"docker-compose"}},
		{[]string{"podman", "compose"}},
		{[]string{"podman-compose"}},
	}

	for _, cmd := range commands {
		c := exec.Command(cmd.args[0], append(cmd.args[1:], "--version")...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		err := c.Run()
		if err == nil {
			return cmd.args, nil
		}
	}

	return nil, fmt.Errorf("compose not found")
}
