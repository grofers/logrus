package logrus

import (
	"bufio"
	"io"
	"runtime"
	"bytes"
	"errors"
)

const (
	MaxScanTokenSize = 256 * 1024
)

func (logger *Logger) Writer() *io.PipeWriter {
	return NewEntry(logger).WriterLevelDynamic()
}

func (logger *Logger) WriterLevel(level Level) *io.PipeWriter {
	return NewEntry(logger).WriterLevel(level)
}

func (entry *Entry) Writer() *io.PipeWriter {
	return entry.WriterLevelDynamic()
}

func (entry *Entry) WriterLevel(level Level) *io.PipeWriter {
	reader, writer := io.Pipe()

	printFunc := entry.get_level_function(level)

	go entry.writerScanner(reader, printFunc)
	runtime.SetFinalizer(writer, writerFinalizer)

	return writer
}

func (entry *Entry) WriterLevelDynamic() *io.PipeWriter {
	reader, writer := io.Pipe()

	go entry.writerScannerDynamic(reader)
	runtime.SetFinalizer(writer, writerFinalizer)

	return writer
}

func (entry *Entry) writerScanner(reader *io.PipeReader, printFunc func(args ...interface{})) {
	err := errors.New("Token error")
	for err != nil {
		err = entry._writerScanner(reader, printFunc)
	}
	reader.Close()
}

func (entry *Entry) _writerScanner(reader *io.PipeReader, printFunc func(args ...interface{})) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(nil, MaxScanTokenSize)
	for scanner.Scan() {
		printFunc(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		entry.Errorf("Error while reading from Writer: %s", err)
		return err
	}
	return nil
}

func (entry *Entry) writerScannerDynamic(reader *io.PipeReader) {
	err := errors.New("Token error")
	for err != nil {
		err = entry._writerScannerDynamic(reader)
	}
	reader.Close()
}

func (entry *Entry) _writerScannerDynamic(reader *io.PipeReader) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(nil, MaxScanTokenSize)
	for scanner.Scan() {
		txt := scanner.Text()
		printFunc := entry.get_level_function(get_level(txt))
		printFunc(txt)
	}
	if err := scanner.Err(); err != nil {
		entry.Errorf("Error while reading from Writer: %s", err)
		return err
	}
	return nil
}

func writerFinalizer(writer *io.PipeWriter) {
	writer.Close()
}

func (entry *Entry) get_level_function(level Level) func(args ...interface{}) {
	var printFunc func(args ...interface{})

	switch level {
	case DebugLevel:
		printFunc = entry.Debug
	case InfoLevel:
		printFunc = entry.Info
	case WarnLevel:
		printFunc = entry.Warn
	case ErrorLevel:
		printFunc = entry.Error
	case FatalLevel:
		printFunc = entry.Fatal
	case PanicLevel:
		printFunc = entry.Panic
	default:
		printFunc = entry.Print
	}

	return printFunc
}

func get_level(line string) Level {
	var lvl string
	line_b := []byte(line)
	x := bytes.IndexByte(line_b, '[')
	if x >= 0 {
		y := bytes.IndexByte(line_b[x:], ']')
		if y >= 0 {
			lvl = string(line_b[x+1 : x+y])
		}
	}
	level, err := ParseLevel(lvl)
	if err != nil {
		level = InfoLevel
	}
	return level
}
