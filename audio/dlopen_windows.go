//go:build nofakecgo && windows

package audio

import (
	"fmt"
	"syscall"
)

func openDynamicLibrary(names []string) (uintptr, error) {
	var lastErr error
	for _, name := range names {
		handle, err := syscall.LoadLibrary(name)
		if err == nil {
			return uintptr(handle), nil
		}
		lastErr = err
	}
	return 0, fmt.Errorf("none of %v loaded: %w", names, lastErr)
}
