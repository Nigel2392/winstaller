package winstall

type InstallerFlags uint32

const (
	F_PRIVILEGED InstallerFlags = 1 << iota
	F_FORCESAVE
)

func (i InstallerFlags) NeedsAdministrator() bool {
	return i&F_PRIVILEGED != 0
}

func (i InstallerFlags) ForceInstall() bool {
	return i&F_FORCESAVE != 0
}
