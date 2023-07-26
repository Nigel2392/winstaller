package winstall

import "github.com/Nigel2392/winstaller/files"

var _ Installation = Shortcut{}

type Shortcut struct {
	From string
	To   string
}

func (i Shortcut) Install(installer Installer) (err error) {
	return files.Shortcut(i.From, i.To)
}

func (i Shortcut) IsInstalled(installer Installer) (bool, error) {
	return files.Exists(i.To)
}
