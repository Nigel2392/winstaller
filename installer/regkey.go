package winstall

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

var _ Installation = (*Regkey)(nil)

type SetterFunc int8

const (
	SetNoValue SetterFunc = iota
	SetDWordValue
	SetQWordValue
	SetStringValue
	SetExpandStringValue
	SetStringsValue
	SetBinaryValue
)

type RegistryValue struct {
	Name       string
	Value      any
	SetterFunc SetterFunc
}

type Regkey struct {
	Key    registry.Key
	Path   string
	Access uint32
	Values []RegistryValue
}

func NewRegKey(key registry.Key, path string, access uint32) *Regkey {
	return &Regkey{
		Key:    key,
		Path:   path,
		Access: access,
	}
}

func (i *Regkey) AddValue(setterFunc SetterFunc, name string, value any) {
	if i.Values == nil {
		i.Values = make([]RegistryValue, 0)
	}
	i.Values = append(i.Values, RegistryValue{
		Name:       name,
		Value:      value,
		SetterFunc: setterFunc,
	})
}

func (i *Regkey) Install(installer Installer) error {
	var key, openedExisting, err = registry.CreateKey(i.Key, i.Path, i.Access)
	if err != nil {
		return fmt.Errorf("failed to create key: %w", err)
	}
	defer key.Close()
	if !openedExisting || installer.Flags.ForceInstall() {
		for _, value := range i.Values {
			switch value.SetterFunc {
			case SetDWordValue:
				err = key.SetDWordValue(value.Name, value.Value.(uint32))
			case SetQWordValue:
				err = key.SetQWordValue(value.Name, value.Value.(uint64))
			case SetStringValue:
				err = key.SetStringValue(value.Name, value.Value.(string))
			case SetExpandStringValue:
				err = key.SetExpandStringValue(value.Name, value.Value.(string))
			case SetStringsValue:
				err = key.SetStringsValue(value.Name, value.Value.([]string))
			case SetBinaryValue:
				err = key.SetBinaryValue(value.Name, value.Value.([]byte))
			case SetNoValue:
				err = nil
			default:
				panic("invalid setter function")
			}
			if err != nil {
				return fmt.Errorf("failed to set value %s: %w", value.Name, err)
			}
		}
	}
	return nil
}

func (i *Regkey) IsInstalled(installer Installer) (bool, error) {
	var key, err = registry.OpenKey(i.Key, i.Path, i.Access)
	if err != nil {
		return false, fmt.Errorf("failed to open key: %w", err)
	}
	defer key.Close()

	// Check if all values are present.
	var isInstalled bool = true
	for _, value := range i.Values {
		switch value.SetterFunc {
		case SetDWordValue:
			_, _, err = key.GetIntegerValue(value.Name)
		case SetQWordValue:
			_, _, err = key.GetIntegerValue(value.Name)
		case SetStringValue:
			_, _, err = key.GetStringValue(value.Name)
		case SetExpandStringValue:
			_, _, err = key.GetStringValue(value.Name)
		case SetStringsValue:
			_, _, err = key.GetStringsValue(value.Name)
		case SetBinaryValue:
			_, _, err = key.GetBinaryValue(value.Name)
		case SetNoValue:
			err = nil
		}
		if err != nil && err != registry.ErrNotExist {
			return false, err
		} else if err == registry.ErrNotExist {
			isInstalled = false
		}
		isInstalled = isInstalled && err == nil

		if !isInstalled {
			break
		}
	}
	return isInstalled, nil
}
