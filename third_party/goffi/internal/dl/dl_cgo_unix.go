//go:build cgo && (linux || darwin || freebsd)

package dl

/*
#cgo linux LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const (
	RTLD_LAZY   = int(C.RTLD_LAZY)
	RTLD_NOW    = int(C.RTLD_NOW)
	RTLD_GLOBAL = int(C.RTLD_GLOBAL)
	RTLD_LOCAL  = int(C.RTLD_LOCAL)
)

const RTLD_DEFAULT = uintptr(0)

func Dlopen(path string, mode int) (uintptr, error) {
	cpath := C.CString(path)
	if cpath == nil {
		return 0, fmt.Errorf("dlopen failed: out of memory")
	}
	defer C.free(unsafe.Pointer(cpath))

	handle := C.dlopen(cpath, C.int(mode))
	if handle == nil {
		return 0, fmt.Errorf("dlopen failed: %s", dlerrorString())
	}
	return uintptr(handle), nil
}

func Dlsym(handle uintptr, name string) (uintptr, error) {
	cname := C.CString(name)
	if cname == nil {
		return 0, fmt.Errorf("dlsym failed: out of memory")
	}
	defer C.free(unsafe.Pointer(cname))

	symbol := C.dlsym(unsafe.Pointer(handle), cname)
	if symbol == nil {
		return 0, fmt.Errorf("dlsym failed: %s", dlerrorString())
	}
	return uintptr(symbol), nil
}

func Dlclose(handle uintptr) error {
	if handle == 0 {
		return nil
	}
	if C.dlclose(unsafe.Pointer(handle)) != 0 {
		return fmt.Errorf("dlclose failed: %s", dlerrorString())
	}
	return nil
}

func dlerrorString() string {
	err := C.dlerror()
	if err == nil {
		return "unknown error"
	}
	return C.GoString(err)
}
