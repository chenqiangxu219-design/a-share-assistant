//go:build !windows

package cache

import (
	"fmt"
	"os"
	"syscall"
)

func getCacheOwnership(path string) (int, int, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, 0, err
	}

	sysStat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, fmt.Errorf("not supported on this platform")
	}
	return int(sysStat.Uid), int(sysStat.Gid), nil
}
