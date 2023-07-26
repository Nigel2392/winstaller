package winstall

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/Nigel2392/winstaller/admin"
	"golang.org/x/sys/windows/registry"
)

type Installation interface {
	Install(installer Installer) error
	IsInstalled(installer Installer) (bool, error)
}

type Installer struct {
	Directory       string
	Installations   []Installation
	OnFinish        func()
	IsInstalledFunc func() (bool, error)

	Flags InstallerFlags
}

func (i Installer) MakePath(paths ...string) string {
	for i, path := range paths {
		paths[i] = filepath.Clean(path)
	}
	return filepath.Join(append([]string{i.Directory}, paths...)...)
}

func (i *Installer) NewInstallation(installations ...Installation) {
	if i.Installations == nil {
		i.Installations = make([]Installation, 0)
	}
	i.Installations = append(i.Installations, installations...)
}

func (i *Installer) Install() error {
	var installed, err = i.IsInstalled()
	if err != nil {
		return fmt.Errorf("failed to check if installed: %w", err)
	}

	if installed && !i.Flags.ForceInstall() {
		return nil
	}

	if i.Flags.NeedsAdministrator() {
		if !admin.Is() {
			return admin.Make(1)
		}
	}

	for _, file := range i.Installations {
		var isInstalled, err = file.IsInstalled(*i)
		if err != nil {
			return fmt.Errorf("failed to check if installed (%T): %w", file, err)
		}
		if isInstalled && i.Flags.ForceInstall() {
			continue
		}
		err = file.Install(*i)
		if err != nil {
			return fmt.Errorf("failed to install (%T): %w", file, err)
		}
	}

	if i.OnFinish != nil {
		i.OnFinish()
	}

	return nil
}

func (i *Installer) IsInstalled() (bool, error) {
	if i.IsInstalledFunc != nil {
		return i.IsInstalledFunc()
	}
	return false, nil
}

func (i *Installer) InstallFile(filename, path string, data []byte) {
	i.NewInstallation(File{
		Filename: filename,
		Path:     path,
		Data:     data,
	})
}

func (i *Installer) InstallFileReader(filename, path string, readerFunc func(path, name string) (reader io.Reader, err error)) {
	i.NewInstallation(IOFile{
		Filename:   filename,
		Path:       path,
		ReaderFunc: readerFunc,
	})
}

func (i *Installer) InstallDirectory(path string) {
	i.NewInstallation(Directory{
		Path: path,
	})
}

func (i *Installer) InstallCopiedFile(from, to string) {
	i.NewInstallation(Copy{
		From: from,
		To:   to,
	})
}

func (i *Installer) InstallShortcut(from, to string) {
	i.NewInstallation(Shortcut{
		From: from,
		To:   to,
	})
}

func (i *Installer) InstallRegistryKey(key registry.Key, path string, access uint32) *Regkey {
	var regkey = &Regkey{
		Key:    key,
		Path:   path,
		Access: access,
	}
	i.NewInstallation(regkey)
	return regkey
}

func (i *Installer) InstallCommand(command string, args ...string) {
	i.NewInstallation(Command{
		Command: command,
		Args:    args,
	})
}
