// +build linux

/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"os"
	"syscall"
)

// openFileWithLock opens the file at path by acquiring an exclive write lock.
// The returned close() function should be called to release the lock and close the file.
func openFileWithLock(path string) (f *os.File, close func(), err error) {
	f, err = os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, nil, err
	}
	// This would lock the file for exclusive writing.
	// If another process is holding and exclusive lock, the call will block
	// until the lock is released.
	lock := syscall.Flock_t{Type: syscall.F_WRLCK}
	if err := syscall.FcntlFlock(f.Fd(), syscall.F_SETLKW, &lock); err != nil {
		f.Close()
		return nil, nil, err
	}
	close = func() {
		// This function should be called once operations with the file are finished.
		// It unlocks the file and closes it.
		unlock := syscall.Flock_t{Type: syscall.F_UNLCK}
		syscall.FcntlFlock(f.Fd(), syscall.F_SETLK, &unlock)
		f.Close()
	}
	return f, close, nil
}
