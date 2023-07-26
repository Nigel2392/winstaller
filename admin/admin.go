package admin

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

var hasAdmin = checkForAdmin()

func Is() bool {
	return hasAdmin
}

func Make(funcsUp int) error {
	pc, _, _, ok := runtime.Caller(1 + funcsUp)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		fmt.Println("Admin check, called from", details.Name())
	}
	if hasAdmin {
		return nil
	}
	var err = ensureAdmin()
	if err != nil {
		return err
	}
	hasAdmin = true
	return nil
}

func ensureAdmin() error {
	// Check if the program is running as admin
	var admin = checkForAdmin()
	if admin {
		return nil
	}
	var exe, _ = os.Executable()
	var cwd, _ = os.Getwd()
	return Force(exe, cwd, os.Args[1:]...)
}

// Force will attempt to run the program as admin
//
// If succeeded, the program will exit with code 0
// The next time the program is restarted, the global hasAdmin variable will be true,
// and the program will not attempt to elevate again
func Force(path string, currentWorkingDir string, arguments ...string) error {
	verbPtr, _ := syscall.UTF16PtrFromString("runas")
	exePtr, _ := syscall.UTF16PtrFromString(path)
	cwdPtr, _ := syscall.UTF16PtrFromString(currentWorkingDir)
	argPtr, _ := syscall.UTF16PtrFromString(strings.Join(arguments, " "))

	var showCmd int32 = 1 //SW_NORMAL

	var err = windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	if err != nil {
		return err
	}

	os.Exit(0)
	return nil
}

func checkForAdmin() bool {
	f, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return false
	}
	f.Close()
	return true
}
