package winstall

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

var _ Installation = Command{}

type Command struct {
	Command         string
	Args            []string
	Stdout          *os.File
	Stderr          *os.File
	SysProcAttr     *syscall.SysProcAttr
	IsInstalledFunc func(installer Installer) (bool, error)
}

func (i Command) Install(installer Installer) (err error) {
	var cmd = exec.Command(i.Command, i.Args...)
	cmd.Stdout = i.Stdout
	cmd.Stderr = i.Stderr
	cmd.SysProcAttr = i.SysProcAttr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run %s: %w", i.Command, err)
	}
	return nil
}

func (i Command) IsInstalled(installer Installer) (bool, error) {
	if i.IsInstalledFunc != nil {
		return i.IsInstalledFunc(installer)
	}
	return false, nil
}
