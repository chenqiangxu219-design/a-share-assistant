//go:build windows

package cache

import "fmt"

func getCacheOwnership(path string) (int, int, error) {
	_ = path
	return 0, 0, fmt.Errorf("ownership not supported on Windows")
}
