package winstall

import (
	"fmt"
	"io"
	"os"

	"github.com/Nigel2392/winstaller/files"
)

var (
	_ Installation = File{}
	_ Installation = IOFile{}
	_ Installation = Directory{}
)

type File struct {
	Filename string
	Path     string
	Data     []byte
}

func (i File) Install(installer Installer) (err error) {
	var flags uint32
	if installer.Flags.NeedsAdministrator() {
		flags = files.F_PRIVILEGED
	}
	if installer.Flags.ForceInstall() {
		flags |= files.F_FORCESAVE
	}
	flags |= files.F_CREATEDIR
	return files.Ensure(installer.MakePath(i.Path, i.Filename), i.Data, flags)
}

func (i File) IsInstalled(installer Installer) (bool, error) {
	return files.Exists(installer.MakePath(i.Path, i.Filename))
}

type IOFile struct {
	Filename   string
	Path       string
	ReaderFunc func(path, name string) (io.Reader, error)
}

func (i IOFile) Install(installer Installer) error {
	var flags uint32
	if installer.Flags.NeedsAdministrator() {
		flags |= files.F_PRIVILEGED
	}
	if installer.Flags.ForceInstall() {
		flags |= files.F_FORCESAVE
	}
	flags |= files.F_CREATEDIR

	var path = installer.MakePath(i.Path, i.Filename)
	var reader, err = i.ReaderFunc(path, i.Filename)
	if err != nil {
		return fmt.Errorf("failed to get reader from %s: %w", path, err)
	}

	return files.EnsureCopy(path, reader, flags)
}

func (i IOFile) IsInstalled(installer Installer) (bool, error) {
	return files.Exists(installer.MakePath(i.Path, i.Filename))
}

type Directory struct {
	Path string
}

func (i Directory) Install(installer Installer) error {
	return os.MkdirAll(installer.MakePath(i.Path), 0755)
}

func (i Directory) IsInstalled(installer Installer) (bool, error) {
	return files.Exists(installer.MakePath(i.Path))
}

type Copy struct {
	From string
	To   string
}

func (i Copy) Install(installer Installer) error {
	var _, err = files.Copy(i.From, i.To)
	return err
}

func (i Copy) IsInstalled(installer Installer) (bool, error) {
	return files.Exists(i.To)
}
