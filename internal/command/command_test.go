package command

import (
	"bytes"
	"os"
	"testing"
)

func TestParseCommandFile(t *testing.T) {
	tests := []struct {
		name    string
		want    *Command
		wantErr bool
	}{
		{
			name: "parse full command",
			want: &Command{
				Help: Help{
					Short: "This shows up on the command list screen",
					Long:  "This shows up on the help screen\n",
					Order: 10,
				},
				Context: "outside-container",
				Steps: []Step{
					{
						Run: func() *string {
							cmd := "docker-compose up -d"
							return &cmd
						}(),
					},
					{
						Echo: func() *string {
							msg := "Hello World"
							return &msg
						}(),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommandFile("testdata/full.yml")
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommandFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("ParseCommandFile() returned nil")
				return
			}

			// Test Help struct
			if got.Help.Short != tt.want.Help.Short {
				t.Errorf("ParseCommandFile() got.Help.Short = %v, want %v", got.Help.Short, tt.want.Help.Short)
			}
			if got.Help.Long != tt.want.Help.Long {
				t.Errorf("ParseCommandFile() got.Help.Long = %v, want %v", got.Help.Long, tt.want.Help.Long)
			}
			if got.Help.Order != tt.want.Help.Order {
				t.Errorf("ParseCommandFile() got.Help.Order = %v, want %v", got.Help.Order, tt.want.Help.Order)
			}

			// Test Context
			if got.Context != tt.want.Context {
				t.Errorf("ParseCommandFile() got.Context = %v, want %v", got.Context, tt.want.Context)
			}

			// Test Steps
			if len(got.Steps) != len(tt.want.Steps) {
				t.Errorf("ParseCommandFile() got.Steps length = %v, want %v", len(got.Steps), len(tt.want.Steps))
			}

			// Test first step (Run)
			if got.Steps[0].Run == nil {
				t.Errorf("ParseCommandFile() got.Steps[0].Run is nil")
			} else if *got.Steps[0].Run != *tt.want.Steps[0].Run {
				t.Errorf("ParseCommandFile() got.Steps[0].Run = %v, want %v", *got.Steps[0].Run, *tt.want.Steps[0].Run)
			}

			// Test second step (Echo)
			if got.Steps[1].Echo == nil {
				t.Errorf("ParseCommandFile() got.Steps[1].Echo is nil")
			} else if *got.Steps[1].Echo != *tt.want.Steps[1].Echo {
				t.Errorf("ParseCommandFile() got.Steps[1].Echo = %v, want %v", *got.Steps[1].Echo, *tt.want.Steps[1].Echo)
			}
		})
	}
}

func TestRunCommand(t *testing.T) {
	cmd := &Command{
		Help: Help{
			Short: "Test command",
			Long:  "This is a test command",
			Order: 1,
		},
		Context: "outside-container",
		Steps: []Step{
			{
				Run: func() *string {
					cmd := "docker-compose up -d"
					return &cmd
				}(),
			},
		},
	}

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	buf := &bytes.Buffer{}
	os.Stdout = os.NewFile(1, "stdout")
	os.Stdout.WriteTo(buf)

	err := RunCommand("test", cmd, true)
	if err != nil {
		t.Errorf("RunCommand() error = %v", err)
	}
	// Restore stdout
	os.Stdout = oldStdout
}
