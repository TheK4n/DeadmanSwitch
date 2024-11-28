package switcher


import (
	"strings"
	"os/exec"
)

type ShellCommandSwitcher struct {
	Command string
}

func (s ShellCommandSwitcher) Execute() ([]byte, error) {
	commandSlice := strings.Fields(s.Command)
	cmd := exec.Command(commandSlice[0], commandSlice[1:]...)
	return cmd.CombinedOutput()
}