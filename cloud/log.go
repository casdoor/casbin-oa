// Copyright 2021 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cloud

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"

	"github.com/casbin/casbin-oa/util"
)

var (
	errorLoggerOnce sync.Once
	errorLogger     *log.Logger
)

// getProjectRoot returns the absolute path of the project root, by walking up from the
// current working directory until a project marker is found, so that the log file always
// lands in the same place no matter which directory the program is started from.
func getProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("getProjectRoot() error: %s\n", err.Error())
		return "."
	}

	wd, err = filepath.Abs(wd)
	if err != nil {
		fmt.Printf("getProjectRoot() error: %s\n", err.Error())
		return "."
	}

	dir := wd
	for {
		if util.FileExist(filepath.Join(dir, "go.mod")) || util.FileExist(filepath.Join(dir, "conf", "app.conf")) {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return wd
		}

		dir = parent
	}
}

// getErrorLogPath returns the absolute path of the error log file: [project root]/logs/cloud_error.log
func getErrorLogPath() string {
	return filepath.Join(getProjectRoot(), "logs", "cloud_error.log")
}

func getErrorLogger() *log.Logger {
	errorLoggerOnce.Do(func() {
		path := getErrorLogPath()

		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			fmt.Printf("getErrorLogger() error: failed to create dir: [%s]: %s\n", filepath.Dir(path), err.Error())
			return
		}

		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Printf("getErrorLogger() error: failed to open file: [%s]: %s\n", path, err.Error())
			return
		}

		fmt.Printf("error log file: [%s]\n", path)
		errorLogger = log.New(f, "", 0)
	})

	return errorLogger
}

// logError prints the error to the console and appends it to the error log file,
// so that the caller can keep running instead of panicking.
func logError(f string, v ...interface{}) {
	line := fmt.Sprintf("[%s] ERROR: %s", util.GetCurrentTimeFormatted(), fmt.Sprintf(f, v...))
	fmt.Println(line)

	logger := getErrorLogger()
	if logger != nil {
		logger.Println(line)
	}
}

// recoverError recovers an unexpected panic and logs it as an error,
// so that the caller's loop keeps running.
func recoverError(name string) {
	r := recover()
	if r == nil {
		return
	}

	logError("%s() panicked: %v, stack:\n%s", name, r, debug.Stack())
}
