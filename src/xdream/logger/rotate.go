package logger

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type RotateHandler  func(path string, didRotate bool)

// RotateFile describes a RotateFile that gets rotated daily
type RotateFile struct {
	sync.Mutex
	pathFormat string

	// info about currently opened RotateFile
	day     int
	path    string
	RotateFile    *os.File
	onClose func(path string, didRotate bool)

	// position in the RotateFile of last Write or Write2, exposed for tests
	lastWritePos int64
}

func (f *RotateFile) close(didRotate bool) error {
	if f.RotateFile == nil {
		return nil
	}
	err := f.RotateFile.Close()
	f.RotateFile = nil
	if err == nil && f.onClose != nil {
		f.onClose(f.path, didRotate)
	}
	f.day = 0
	return err
}

func (f *RotateFile) open() error {
	t := time.Now().UTC()
	f.path = t.Format(f.pathFormat)
	f.day = t.YearDay()

	// we can't assume that the dir for the RotateFile already exists
	dir := filepath.Dir(f.path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	// would be easier to open with os.O_APPEND but Seek() doesn't work in that case
	flag := os.O_CREATE | os.O_WRONLY
	f.RotateFile, err = os.OpenFile(f.path, flag, 0644)
	if err != nil {
		return err
	}
	_, err = f.RotateFile.Seek(0, io.SeekEnd)
	return err
}

// rotate on new day
func (f *RotateFile) reopenIfNeeded() error {
	t := time.Now().UTC()
	if t.YearDay() == f.day {
		return nil
	}
	err := f.close(true)
	if err != nil {
		return err
	}
	return f.open()
}

// NewRotateFile creates a new RotateFile that will be rotated daily (at UTC midnight).
// pathFormat is RotateFile format accepted by time.Format that will be used to generate
// a name of the RotateFile. It should be unique in a given day e.g. 2006-01-02.txt.
// onClose is an optional function that will be called every time existing RotateFile
// is closed, either as a result calling Close or due to being rotated.
// didRotate will be true if it was closed due to rotation.
// If onClose() takes a long time, you should do it in a background goroutine
// (it blocks all other operations, including writes)
func NewRotateFile(pathFormat string, rotateHandler RotateHandler) (*RotateFile, error) {
	f := &RotateFile{
		pathFormat: pathFormat,
	}
	// force early failure if we can't open the RotateFile
	// note that we don't set onClose yet so that it won't get called due to
	// opening/closing the RotateFile
	err := f.reopenIfNeeded()
	if err != nil {
		return nil, err
	}
	err = f.close(false)
	if err != nil {
		return nil, err
	}
	f.onClose = rotateHandler
	return f, nil
}

// Close closes the RotateFile
func (f *RotateFile) Close() error {
	f.Lock()
	defer f.Unlock()
	return f.close(false)
}

func (f *RotateFile) write(d []byte, flush bool) (int64, int, error) {
	err := f.reopenIfNeeded()
	if err != nil {
		return 0, 0, err
	}
	f.lastWritePos, err = f.RotateFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, 0, err
	}
	n, err := f.RotateFile.Write(d)
	if err != nil {
		return 0, n, err
	}
	if flush {
		err = f.RotateFile.Sync()
	}
	return f.lastWritePos, n, err
}

// Write writes data to a RotateFile
func (f *RotateFile) Write(d []byte) (int, error) {
	f.Lock()
	defer f.Unlock()
	_, n, err := f.write(d, false)
	return n, err
}

// Flush flushes the RotateFile
func (f *RotateFile) Flush() error {
	return f.Sync()
}

func (f *RotateFile)Sync() error {
	f.Lock()
	defer f.Unlock()
	return f.RotateFile.Sync()
}