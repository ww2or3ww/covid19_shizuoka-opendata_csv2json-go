package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

type (
	LogLv int
)

const (
	_ LogLv = iota
	Debug
	Info
	Warn
	Error
)

var logLvStrings = [...]string{"", "DEBUG", "INFO", "WARN", "ERROR"}

var logLv LogLv = Debug
var fileMethodLength int = 15
var logFormatString string = ",%5s,%" + strconv.Itoa(fileMethodLength) + "s(%3d),%s"

func LogInitialize(logLvIn LogLv, fileMethodLengthIn int) {
	logLv = logLvIn
	fileMethodLength = fileMethodLengthIn
	logFormatString = ",%5s,%" + strconv.Itoa(fileMethodLength) + "s(%3d),%s"

	writer := io.MultiWriter(os.Stdout)
	log.SetFlags(log.Ldate | log.Lmicroseconds)
	log.SetOutput(writer)

	Infof("=== Log Initialzed | Lv=%s | Fmt=%s ===", logLvStrings[logLv], logFormatString)
}

func TypeName(value interface{}) {
	fmt.Println(reflect.TypeOf(value))
}

func Debugf(format string, args ...interface{}) {
	logPrintln(Debug, fmt.Sprintf(format, args...))
}

func Debugs(data ...interface{}) {
	logPrintln(Debug, makeStringFromInterfaces(data...))
}

func Infof(format string, args ...interface{}) {
	logPrintln(Info, fmt.Sprintf(format, args...))
}

func Infos(data ...interface{}) {
	logPrintln(Info, fmt.Sprintf("%v", makeStringFromInterfaces(data...)))
}

func Warnf(format string, args ...interface{}) {
	logPrintln(Warn, fmt.Sprintf(format, args...))
}

func Warns(data ...interface{}) {
	logPrintln(Warn, makeStringFromInterfaces(data...))
}

func Errorf(format string, args ...interface{}) {
	logPrintln(Error, fmt.Sprintf(format, args...))
}

func Errors(data ...interface{}) {
	logPrintln(Error, makeStringFromInterfaces(data...))
}

func makeStringFromInterfaces(dataList ...interface{}) string {
	var slc []string
	for _, data := range dataList {
		slc = append(slc, fmt.Sprintf("%v", data))
	}
	return strings.Join(slc, ",")
}

func logPrintln(logLvIn LogLv, text string) {
	if logLvIn < logLv {
		return
	}
	pc, path, lineNo, _ := runtime.Caller(2)
	fileNameWithoutExt := filepath.Base(path[:len(path)-len(filepath.Ext(path))])
	funcForPcName := runtime.FuncForPC(pc).Name()
	methodWithoutPackage := funcForPcName[strings.Index(funcForPcName, ".")+1:]
	fileMethod := fileNameWithoutExt + ":" + methodWithoutPackage
	log.Println(fmt.Sprintf(logFormatString, logLvStrings[logLvIn], fileMethod, lineNo, text))
}
