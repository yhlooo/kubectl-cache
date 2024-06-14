//go:build windows

package proxymgr

import "os"

// lockFile 锁文件
func lockFile(f *os.File) error {
	// Windows 不支持文件锁，什么也不用做
	return nil
}

// unlockFile 解锁文件
func unlockFile(f *os.File) error {
	// Windows 不支持文件锁，什么也不用做
	return nil
}
