package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"sync"
)

type severity int

const (
	sInfo severity = iota
	sWarning
	sError
	sFatal
)

const (
	flags = log.Ldate | log.Ltime | log.Lshortfile
)

var (
	logLock       sync.Mutex
	defaultLogger *Logger
)

// Logger ...
type Logger struct {
	infoLog     *log.Logger
	warningLog  *log.Logger
	errorLog    *log.Logger
	fatalLog    *log.Logger
	closers     []io.Closer
	initialized bool
}

func (l *Logger) output(s severity, depth int, txt string) {
	logLock.Lock()
	defer logLock.Unlock()
	switch s {
	case sInfo:
		l.infoLog.Output(3+depth, txt)
	case sWarning:
		l.warningLog.Output(3+depth, txt)
	case sError:
		l.errorLog.Output(3+depth, txt)
	case sFatal:
		l.fatalLog.Output(3+depth, txt)
	default:
		panic(fmt.Sprintln("unrecognized severity: ", s))
	}
}

// Info ...
func Info(v ...interface{}) {
	defaultLogger.output(sInfo, 0, fmt.Sprint(v...))
}

// Infof ...
func Infof(format string, v ...interface{}) {
	defaultLogger.output(sInfo, 0, fmt.Sprintf(format, v...))
}

// Warning ...
func Warning(v ...interface{}) {
	defaultLogger.output(sWarning, 0, fmt.Sprint(v...))
}

// Warningf ...
func Warningf(format string, v ...interface{}) {
	defaultLogger.output(sWarning, 0, fmt.Sprintf(format, v...))
}

// Error ...
func Error(v ...interface{}) {
	defaultLogger.output(sError, 0, fmt.Sprint(v...))
}

// Errorf ...
func Errorf(format string, v ...interface{}) {
	defaultLogger.output(sError, 0, fmt.Sprintf(format, v...))
}

// Fatal ...
func Fatal(v ...interface{}) {
	defaultLogger.output(sFatal, 0, fmt.Sprint(v...))
	defaultLogger.Close()
	os.Exit(1)
}

// Fatalf ...
func Fatalf(format string, v ...interface{}) {
	defaultLogger.output(sFatal, 0, fmt.Sprintf(format, v...))
	defaultLogger.Close()
	os.Exit(1)
}

// Close ...
func (l *Logger) Close() {
	logLock.Lock()
	defer logLock.Unlock()
	for _, c := range l.closers {
		if err := c.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close log %v: %v\n", c, err)
		}
	}
}

// Init ...
func Init(infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer,
	fatalHandle io.Writer) {
	defaultLogger = &Logger{
		infoLog:    log.New(infoHandle, "INFO: ", flags),
		warningLog: log.New(warningHandle, "WARNING: ", flags),
		errorLog:   log.New(errorHandle, "ERROR: ", flags),
		fatalLog:   log.New(fatalHandle, "FATAL: ", flags),
	}
}

// FileForSaving ...
func FileForSaving(fileName string) *os.File {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", file, ":", err)
	}
	return file
}

// CustomError ..
type CustomError struct {
	Message    string
	StackTrace string
}

// ErrorStruct presents custom errors
type ErrorStruct struct {
	CustomError
}

// WrapError return custom error struct
func WrapError(messagef string) CustomError {
	return CustomError{
		Message:    messagef,
		StackTrace: string(debug.Stack()),
	}
}

// Return ...
func Return(err error) error {
	return ErrorStruct{CustomError: WrapError(err.Error())}
}

// Need implement issue, you can do not use it
func (err CustomError) Error() string {
	return err.Message
}

// Example
// if err := step(); err != nil {
//   errx := errorx.New(http.StatusInternalServerError, "Operation failed")
//   errx.Wrap(err)
// }
