/*
Copyright 2016 Google PostInc. All Rights Reserved.
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

// Package log offers simple cross platform logging for Windows and Linux.
// Available logging endpoints are event log (Windows), syslog (Linux), and
// an io.Writer.
package log

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"io"
	"log"
	"os"
	"sync"
)

type severity int

// Severity levels.
const (
	sInfo severity = iota
	sWarning
	sError
	sFatal
)

// Severity tags.
const (
	tagInfo    = "INFO : "
	tagWarning = "WARN : "
	tagError   = "ERROR: "
	tagFatal   = "FATAL: "
)

const (
	flags    = log.Ldate | log.Lmicroseconds | log.Lshortfile
	initText = "? "
)

var (
	logLock       sync.Mutex
	defaultLogger *Logger
)

// initialize resets defaultLogger.  Which allows tests to reset environment.
func initialize() {
	defaultLogger = &Logger{
		infoLog:    log.New(os.Stderr, initText+tagInfo, flags),
		warningLog: log.New(os.Stderr, initText+tagWarning, flags),
		errorLog:   log.New(os.Stderr, initText+tagError, flags),
		fatalLog:   log.New(os.Stderr, initText+tagFatal, flags),
	}
}

func init() {
	initialize()
}

// Init sets up logging and should be called before log functions, usually in
// the caller's main(). Default log functions can be called before Init(), but log
// output will only go to stderr (along with a warning).
// The first call to Init populates the default log and returns the
// generated log, subsequent calls to Init will only return the generated
// log.
// If the logFile passed in also satisfies io.Closer, logFile.Close will be called
// when closing the log.

type Config struct {
	Name      string    // log name
	Verbose   bool      // verbose log to stdout/stderr
	LogWriter io.Writer // writer to write log
	LogFile   string    // filename to write log
	SentryDsn string    // log to sentry
	Exclusive bool      // do not use as default log
}

func (cfg Config) Init() *Logger {
	logw := io.Writer(nil)
	if cfg.LogWriter != nil {
		logw = cfg.LogWriter
	} else if cfg.LogFile != "" {
		f, err := os.Create(cfg.LogFile)
		if err != nil {
			panic(err)
		}
		logw = f
	}

	makeLog := func(level severity, w ...io.Writer) *log.Logger {
		var tag string
		var a []io.Writer
		var v io.Writer
		if logw != nil {
			a = append(a, logw)
		}
		switch level {
		case sInfo:
			v = os.Stdout
			tag = tagInfo
		case sWarning:
			v = os.Stdout
			tag = tagWarning
		case sError:
			v = os.Stderr
			tag = tagError
		case sFatal:
			v = os.Stderr
			tag = tagFatal
		}
		if cfg.Verbose {
			a = append(a, v)
		}
		a = append(a, w...)
		return log.New(io.MultiWriter(a...), tag, flags)
	}

	l := Logger{
		infoLog:    makeLog(sInfo, sentryInfoLog),
		warningLog: makeLog(sWarning, sentryWarnLog),
		errorLog:   makeLog(sError, sentryErrorLog),
		fatalLog:   makeLog(sFatal, sentryFatalLog),
	}

	l.closers = append(l.closers, sentryFatalLog)
	if logw != nil {
		if c, ok := logw.(io.Closer); ok && c != nil {
			l.closers = append(l.closers, c)
		}
	}

	l.initialized = true

	logLock.Lock()
	defer logLock.Unlock()
	if !cfg.Exclusive && !defaultLogger.initialized {
		defaultLogger = &l
	}

	if !cfg.Exclusive && cfg.SentryDsn != "" {
		err := sentry.Init(sentry.ClientOptions{Dsn: cfg.SentryDsn})
		if err != nil {
			l.Warningf("failed to connect Sentry: %v", err.Error())
		}
	}

	return &l
}

// Close closes the default log.
func Close() {
	defaultLogger.Close()
}

// A Logger represents an active logging object. Multiple loggers can be used
// simultaneously even if they are using the same same writers.
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
		panic(fmt.Sprintln("unrecognized severity:", s))
	}
}

// Close closes all the underlying log writers, which will flush any cached logs.
// Any errors from closing the underlying log writers will be printed to stderr.
// Once Close is called, all future calls to the log will panic.
func (l *Logger) Close() {
	logLock.Lock()
	defer logLock.Unlock()

	if !l.initialized {
		return
	}

	for _, c := range l.closers {
		if err := c.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close log %v: %v\n", c, err)
		}
	}

	if l == defaultLogger {
		initialize()
	}
}

// Info logs with the Info severity.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Info(v ...interface{}) {
	l.output(sInfo, 0, fmt.Sprint(v...))
}

// InfoDepth acts as Info but uses depth to determine which call frame to log.
// InfoDepth(0, "msg") is the same as Info("msg").
func (l *Logger) InfoDepth(depth int, v ...interface{}) {
	l.output(sInfo, depth, fmt.Sprint(v...))
}

// Infoln logs with the Info severity.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Infoln(v ...interface{}) {
	l.output(sInfo, 0, fmt.Sprintln(v...))
}

// Infof logs with the Info severity.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Infof(format string, v ...interface{}) {
	l.output(sInfo, 0, fmt.Sprintf(format, v...))
}

// Warning logs with the Warning severity.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Warning(v ...interface{}) {
	l.output(sWarning, 0, fmt.Sprint(v...))
}

// WarningDepth acts as Warning but uses depth to determine which call frame to log.
// WarningDepth(0, "msg") is the same as Warning("msg").
func (l *Logger) WarningDepth(depth int, v ...interface{}) {
	l.output(sWarning, depth, fmt.Sprint(v...))
}

// Warningln logs with the Warning severity.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Warningln(v ...interface{}) {
	l.output(sWarning, 0, fmt.Sprintln(v...))
}

// Warningf logs with the Warning severity.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Warningf(format string, v ...interface{}) {
	l.output(sWarning, 0, fmt.Sprintf(format, v...))
}

// Error logs with the ERROR severity.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Error(v ...interface{}) {
	l.output(sError, 0, fmt.Sprint(v...))
}

// ErrorDepth acts as Error but uses depth to determine which call frame to log.
// ErrorDepth(0, "msg") is the same as Error("msg").
func (l *Logger) ErrorDepth(depth int, v ...interface{}) {
	l.output(sError, depth, fmt.Sprint(v...))
}

// Errorln logs with the ERROR severity.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Errorln(v ...interface{}) {
	l.output(sError, 0, fmt.Sprintln(v...))
}

// Errorf logs with the Error severity.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.output(sError, 0, fmt.Sprintf(format, v...))
}

// Fatal logs with the Fatal severity, and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Fatal(v ...interface{}) {
	l.output(sFatal, 0, fmt.Sprint(v...))
	l.Close()
	os.Exit(1)
}

// FatalDepth acts as Fatal but uses depth to determine which call frame to log.
// FatalDepth(0, "msg") is the same as Fatal("msg").
func (l *Logger) FatalDepth(depth int, v ...interface{}) {
	l.output(sFatal, depth, fmt.Sprint(v...))
	l.Close()
	os.Exit(1)
}

// Fatalln logs with the Fatal severity, and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Fatalln(v ...interface{}) {
	l.output(sFatal, 0, fmt.Sprintln(v...))
	l.Close()
	os.Exit(1)
}

// Fatalf logs with the Fatal severity, and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.output(sFatal, 0, fmt.Sprintf(format, v...))
	l.Close()
	os.Exit(1)
}

// SetFlags sets the output flags for the log.
func SetFlags(flag int) {
	defaultLogger.infoLog.SetFlags(flag)
	defaultLogger.warningLog.SetFlags(flag)
	//defaultLogger.errorLog.SetFlags(flag)
	//defaultLogger.fatalLog.SetFlags(flag)
}

// Info uses the default log and logs with the Info severity.
// Arguments are handled in the manner of fmt.Print.
func Info(v ...interface{}) {
	defaultLogger.output(sInfo, 0, fmt.Sprint(v...))
}

// InfoDepth acts as Info but uses depth to determine which call frame to log.
// InfoDepth(0, "msg") is the same as Info("msg").
func InfoDepth(depth int, v ...interface{}) {
	defaultLogger.output(sInfo, depth, fmt.Sprint(v...))
}

// Infoln uses the default log and logs with the Info severity.
// Arguments are handled in the manner of fmt.Println.
func Infoln(v ...interface{}) {
	defaultLogger.output(sInfo, 0, fmt.Sprintln(v...))
}

// Infof uses the default log and logs with the Info severity.
// Arguments are handled in the manner of fmt.Printf.
func Infof(format string, v ...interface{}) {
	defaultLogger.output(sInfo, 0, fmt.Sprintf(format, v...))
}

// Warning uses the default log and logs with the Warning severity.
// Arguments are handled in the manner of fmt.Print.
func Warning(v ...interface{}) {
	defaultLogger.output(sWarning, 0, fmt.Sprint(v...))
}

// WarningDepth acts as Warning but uses depth to determine which call frame to log.
// WarningDepth(0, "msg") is the same as Warning("msg").
func WarningDepth(depth int, v ...interface{}) {
	defaultLogger.output(sWarning, depth, fmt.Sprint(v...))
}

// Warningln uses the default log and logs with the Warning severity.
// Arguments are handled in the manner of fmt.Println.
func Warningln(v ...interface{}) {
	defaultLogger.output(sWarning, 0, fmt.Sprintln(v...))
}

// Warningf uses the default log and logs with the Warning severity.
// Arguments are handled in the manner of fmt.Printf.
func Warningf(format string, v ...interface{}) {
	defaultLogger.output(sWarning, 0, fmt.Sprintf(format, v...))
}

// Error uses the default log and logs with the Error severity.
// Arguments are handled in the manner of fmt.Print.
func Error(v ...interface{}) {
	defaultLogger.output(sError, 0, fmt.Sprint(v...))
}

// ErrorDepth acts as Error but uses depth to determine which call frame to log.
// ErrorDepth(0, "msg") is the same as Error("msg").
func ErrorDepth(depth int, v ...interface{}) {
	defaultLogger.output(sError, depth, fmt.Sprint(v...))
}

// Errorln uses the default log and logs with the Error severity.
// Arguments are handled in the manner of fmt.Println.
func Errorln(v ...interface{}) {
	defaultLogger.output(sError, 0, fmt.Sprintln(v...))
}

// Errorf uses the default log and logs with the Error severity.
// Arguments are handled in the manner of fmt.Printf.
func Errorf(format string, v ...interface{}) {
	defaultLogger.output(sError, 0, fmt.Sprintf(format, v...))
}

// Fatalln uses the default log, logs with the Fatal severity,
// and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Print.
func Fatal(v ...interface{}) {
	defaultLogger.output(sFatal, 0, fmt.Sprint(v...))
	defaultLogger.Close()
	os.Exit(1)
}

// FatalDepth acts as Fatal but uses depth to determine which call frame to log.
// FatalDepth(0, "msg") is the same as Fatal("msg").
func FatalDepth(depth int, v ...interface{}) {
	defaultLogger.output(sFatal, depth, fmt.Sprint(v...))
	defaultLogger.Close()
	os.Exit(1)
}

// Fatalln uses the default log, logs with the Fatal severity,
// and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Println.
func Fatalln(v ...interface{}) {
	defaultLogger.output(sFatal, 0, fmt.Sprintln(v...))
	defaultLogger.Close()
	os.Exit(1)
}

// Fatalf uses the default log, logs with the Fatal severity,
// and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Printf.
func Fatalf(format string, v ...interface{}) {
	defaultLogger.output(sFatal, 0, fmt.Sprintf(format, v...))
	defaultLogger.Close()
	os.Exit(1)
}
