package logger

import (
	"os"
	"sync"
	"github.com/jonboulle/clockwork"
	"time"
)

var clock clockwork.Clock = clockwork.NewRealClock()

type RotateHandler  func(path string, didRotate bool)


// RotateFile describes a RotateFile that gets rotated daily
type RotateFile struct {
	sync.Mutex

	rotateType int  // 0. 不滚动  1.按天滚动 2.按小时滚动
	pathFormat string

	checkExist bool

	// info about currently opened RotateFile
	curKey     int
	path    string
	RotateFile    *os.File
	onClose func(path string, didRotate bool)
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
	f.curKey = -1

	return err
}

func (f *RotateFile) open(filePath string, newKey int) error {
	f.path = filePath
	f.curKey = newKey

	var err error
	// would be easier to open with os.O_APPEND but Seek() doesn't work in that case
	flag := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	f.RotateFile, err = os.OpenFile(f.path, flag, 0755)

	return err
}

func (f *RotateFile) GetFileName(t time.Time) string {
	var fname string
	switch f.rotateType {
	case 1:
		fname = t.Format(f.pathFormat)
	case 2:
		fname = t.Format(f.pathFormat)
	default:
		fname = f.pathFormat
	}
	return fname
}

// rotate on new day
func (f *RotateFile) reopenIfNeeded() error {
	t := clock.Now().UTC()
	var newKey int
	var fname string
	if f.checkExist {
		fname = f.GetFileName(t)
		_, err := os.Stat(fname)

		if err==nil &&  f.RotateFile!=nil{
			return  nil
		}

		//文件被强制删除了，需要关闭后重新创建
		if 	os.IsPermission(err) && f.RotateFile!=nil {
			f.close(true)
		}

	}else {
		switch f.rotateType {
		case 1:
			if t.YearDay() == f.curKey {
				return nil
			}
			err := f.close(true)
			if err != nil {
				return err
			}
			fname = f.GetFileName(t)
			newKey = t.YearDay()
		case 2:
			if t.Hour() == f.curKey {
				return nil
			}
			err := f.close(true)
			if err != nil {
				return err
			}
			fname = f.GetFileName(t)
			newKey = t.Hour()
		default:
			fname = f.GetFileName(t)
			if f.curKey == 1 {
				return  nil
			}
			newKey = 1
		}
	}
	return f.open(fname, newKey)
}


func NewRotateFile(rotateType string, filePath string, checkExsit bool,  rotateHandler RotateHandler) (*RotateFile, error) {
	 rt := 0
	 var pathFormat string
	 var typeMap map[string]int = map[string]int{
	 	"day":1,
	 	"hour":2,
	 	"none":0,
	 }
	rt = typeMap[rotateType]

	if rt == 1 {
		pathFormat = filePath+".2006-01-02"
	}else if rt==2 {
		pathFormat = filePath+".2006-01-02_15"
	}else{
		pathFormat = filePath
	}

	f := &RotateFile{
		checkExist:checkExsit,
		rotateType: rt,
		pathFormat: pathFormat,
	}

	err := ensureDir(filePath)
	if err != nil {
		return nil, err
	}

	// force early failure if we can't open the RotateFile
	// note that we don't set onClose yet so that it won't get called due to
	// opening/closing the RotateFile
	err = f.reopenIfNeeded()
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

func (f *RotateFile) write(d []byte, flush bool) (int, error) {
	err := f.reopenIfNeeded()
	if err != nil {
		return 0, err
	}

	n, err := f.RotateFile.Write(d)
	if err != nil {
		return  n, err
	}
	if flush {
		err = f.RotateFile.Sync()
	}
	return n, err
}

// Write writes data to a RotateFile
func (f *RotateFile) Write(d []byte) (int, error) {
	f.Lock()
	defer f.Unlock()
	n, err := f.write(d, false)
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