package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func Error(v ...interface{}) {
	fmt.Fprintln(os.Stderr, format("Info", ""), v)
}

func Errorf(f string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format("Info", " ")+f+"\n", v)
}

func Info(v ...interface{}) {
	fmt.Fprintln(os.Stdout, format("Info", ""), v)
}

func Infof(f string, v ...interface{}) {
	fmt.Fprintf(os.Stdout, format("Info", " ")+f+"\n", v)
}

func format(prefix, suffix string) string {
	_, file, line, _ := runtime.Caller(2)
	file = filepath.Base(file)
	return fmt.Sprintf("%s %s:%d:%s", prefix, file, line, suffix)
}
