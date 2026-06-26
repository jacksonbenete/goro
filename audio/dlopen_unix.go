//go:build nofakecgo && !windows

package audio

import (
	"fmt"

	"github.com/ebitengine/purego"
)

func openDynamicLibrary(names []string) (uintptr, error) {
	var lastErr error
	for _, name := range names {
		handle, err := purego.Dlopen(name, purego.RTLD_LAZY|purego.RTLD_GLOBAL)
		if err == nil {
			return handle, nil
		}
		lastErr = err
	}
	return 0, fmt.Errorf("none of %v loaded: %w", names, lastErr)
}
