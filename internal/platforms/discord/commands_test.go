package discord

import (
	"testing"
)

func TestCommands(t *testing.T) {
	if len(Commands) == 0 {
		t.Fatal("expected Commands to have registered commands")
	}

	commandMap := make(map[string]bool)
	for _, cmd := range Commands {
		if cmd.Name == "" {
			t.Error("command name cannot be empty")
		}
		if cmd.Description == "" {
			t.Errorf("command %s description cannot be empty", cmd.Name)
		}
		commandMap[cmd.Name] = true
	}

	expected := []string{
		"now",
		"user_id",
		"clear",
	}

	for _, name := range expected {
		if !commandMap[name] {
			t.Errorf("missing expected command: %s", name)
		}
	}
}
