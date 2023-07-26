package dialog

import (
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

type DialogType string

// MessageDialogOptions contains the options for the Message dialogs, EG Info, Warning, etc runtime methods
type Options struct {
	Type          DialogType
	Title         string
	Message       string
	Buttons       []string
	DefaultButton string
	CancelButton  string
	Icon          []byte
}

const (
	Info     DialogType = "info"
	Warning  DialogType = "warning"
	Error    DialogType = "error"
	Question DialogType = "question"
)

func calculateMessageDialogFlags(options Options) uint32 {
	var flags uint32

	switch options.Type {
	case Info:
		flags = windows.MB_OK | windows.MB_ICONINFORMATION
	case Error:
		flags = windows.MB_ICONERROR | windows.MB_OK
	case Question:
		flags = windows.MB_YESNO
		if strings.TrimSpace(strings.ToLower(options.DefaultButton)) == "no" {
			flags |= windows.MB_DEFBUTTON2
		}
	case Warning:
		flags = windows.MB_OK | windows.MB_ICONWARNING
	}

	return flags
}

// MessageDialog show a message dialog to the user
func Dialog(options Options) (string, error) {
	var title, err = syscall.UTF16PtrFromString(options.Title)
	if err != nil {
		return "", err
	}
	message, err := syscall.UTF16PtrFromString(options.Message)
	if err != nil {
		return "", err
	}
	var flags = calculateMessageDialogFlags(options)
	var button, _ = windows.MessageBox(0, message, title, flags|windows.MB_SYSTEMMODAL)
	var responses = []string{"", "Ok", "Cancel", "Abort", "Retry", "Ignore", "Yes", "No", "", "", "Try Again", "Continue"}
	var result = "Error"
	if int(button) < len(responses) {
		result = responses[button]
	}
	return result, nil
}
