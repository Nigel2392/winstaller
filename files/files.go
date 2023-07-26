package files

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Nigel2392/winstaller/admin"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

const (
	F_PRIVILEGED uint32 = 1 << iota
	F_FORCESAVE
	F_CREATEDIR
)

func Ensure[T ~string | ~[]byte](path string, data T, flags ...uint32) error {
	return WithEnsurance(path, ensureWriteBytes(data), flags...)
}

func EnsureCopy(path string, r io.Reader, flags ...uint32) error {
	return WithEnsurance(path, ensureIOCopied(r), flags...)
}

func ensureWriteBytes[T ~string | ~[]byte](data T) func(f *os.File) error {
	return func(f *os.File) error {
		_, err := f.Write([]byte(data))
		return err
	}
}

func ensureIOCopied(r io.Reader) func(f *os.File) error {
	return func(f *os.File) error {
		_, err := io.Copy(f, r)
		return err
	}
}

func WithEnsurance(path string, writeFunc func(*os.File) error, flags ...uint32) error {
	var flag uint32
	for _, f := range flags {
		flag |= f
	}
	var needsUAC = flag&F_PRIVILEGED != 0
	var forceSave = flag&F_FORCESAVE != 0
	var createDir = flag&F_CREATEDIR != 0
	var err error
	if needsUAC {
		if err = admin.Make(1); err != nil {
			return err
		}
	}
	if createDir {
		var base, _ = filepath.Split(path)
		err = os.MkdirAll(base, os.ModePerm)
		if err != nil && !os.IsExist(err) {
			return err
		}
	}
	_, err = os.Stat(path)
	if os.IsNotExist(err) || forceSave {
		var f *os.File
		f, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return errors.New("error while opening or creating file: " + err.Error())
		}
		defer f.Close()
		err = writeFunc(f)
		if err != nil {
			return errors.New("error while writing file: " + err.Error())
		}
	} else if err != nil {
		return errors.New("unknown error while checking file existence: " + err.Error())
	}
	return nil
}

func Shortcut(src, dst string) error {
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY)
	oleShellObject, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return err
	}
	defer oleShellObject.Release()
	wshell, err := oleShellObject.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer wshell.Release()
	cs, err := oleutil.CallMethod(wshell, "CreateShortcut", dst)
	if err != nil {
		return err
	}
	idispatch := cs.ToIDispatch()
	oleutil.PutProperty(idispatch, "TargetPath", src)
	oleutil.CallMethod(idispatch, "Save")
	return nil
}

func Copy(src, dst string) (n int64, err error) {
	var (
		srcFile *os.File
		dstFile *os.File
	)
	srcFile, err = os.Open(src)
	if err != nil {
		return 0, fmt.Errorf("failed to open %s: %w", src, err)
	}
	defer srcFile.Close()

	dstFile, err = os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return 0, fmt.Errorf("failed to create %s: %w", dst, err)
	}
	defer dstFile.Close()

	n, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return 0, fmt.Errorf("failed to copy %s to %s: %w", src, dst, err)
	}
	return n, nil
}

func Exists(path string) (bool, error) {
	var _, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
