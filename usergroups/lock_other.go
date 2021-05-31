// +build !linux

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
)

// openFileWithLock opens a file but does not lock it on non-Linux.
func openFileWithLock(path string) (f *os.File, close func(), err error) {
	f, err = os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, nil, err
	}
	close = func() {
		f.Close()
	}
	return f, close, nil
}
