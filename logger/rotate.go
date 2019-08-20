package logger

import (
	"errors"
	"fmt"
	"github.com/zero-miao/go-utils/email"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"syscall"
	"time"
)

// ==========================
// @Author : zero-miao
// @Date   : 2019-05-22 15:55
// @File   : rotate.go
// @Project: aotu/logger
// ==========================

// head in, tail out.
// when add item, head-1, and check tail equal to head.
// example:
//   length=3; 依次加入 1,2,3,4. 变化如下:
//   items = [x, x, x], head=0, tail=0, size=0; // add 1
//   items = [1, x, x], head=1, tail=0, size=1; // add 2
//   items = [1, 2, x], head=2, tail=0, size=2; // add 3
//   items = [1, 2, 3], head=0, tail=0, size=3; // add 4
//   items = [4, 2, 3], head=1, tail=1, size=3; // add 4
type Deque struct {
	items []string
	cap   int // 容量
	size  int // 实际元素个数
	head  int // 指向下一个要填入 item 的位置. 如果 size < cap, 说明在第一圈. 不会返回任何元素.
	mu    sync.Mutex
}

func (d *Deque) String() string {
	return fmt.Sprintf("Deque<size=%v, cap=%v>%v", d.size, d.cap, d.items)
}

func (d *Deque) Reset(cap int) {
	d.items = make([]string, cap)
	d.cap = cap
	d.size = 0
	d.head = 0
}

func (d *Deque) Add(item string) string {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.size < d.cap {
		d.items[d.head] = item
		d.size++
		d.head = (d.head + 1) % d.cap
		return ""
	} else {
		temp := d.items[d.head]
		d.items[d.head] = item
		d.head = (d.head + 1) % d.cap
		return temp
	}
}

type RotateFile struct {
	lock       sync.Mutex
	filename   string
	fp         *os.File
	lastRotate time.Time // 上一次切割时间.
	validFiles Deque     // 循环队列
	replica    int       // 队列长度
}

func (w *RotateFile) Reset(layout string, duration time.Duration) error {
	w.lock.Lock()
	defer w.lock.Unlock()
	if w.filename == "" {
		return errors.New("no filename")
	}
	if w.fp != nil {
		err := w.fp.Close()
		if err != nil {
			return err
		}
		w.fp = nil
	}
	f, err := os.OpenFile(w.filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	w.fp = f
	w.validFiles.Reset(w.replica)
	fn := w.filename + ".*"
	matches, err := filepath.Glob(fn)
	if err != nil {
		return err
	}
	sort.Strings(matches) // 字典序从小到大(即日期从早到晚)
	lastItem := ""
	for _, item := range matches {
		lastItem = item
		ftd := w.validFiles.Add(item)
		if ftd != "" {
			err := os.Remove(ftd)
			return err
		}
	}
	if lastItem != "" && layout != "" {
		lastItem = lastItem[len(w.filename)+1:]
		t, err := time.Parse(layout, lastItem)
		if err == nil {
			w.lastRotate = t.Add(duration)
		}
	}
	return nil
}

func (w *RotateFile) String() string {
	return fmt.Sprintf("fn=%s, last_rotate=%v, history=%v", w.filename, w.lastRotate.Format(time.RFC3339), w.validFiles.String())
}

// 考虑两个协程同时切割
func (w *RotateFile) Rotate(layout string, duration time.Duration) error {
	w.lock.Lock()
	// 退出时, 重置 w.fp
	defer func() {
		// Close existing file if open
		errs := make([]string, 0)
		if w.fp != nil {
			err := w.fp.Sync()
			if err != nil {
				errs = append(errs, fmt.Sprintf("[w.fp.sync] err=%v; w=%v", err, w))
			}
			err = w.fp.Close()
			if err != nil {
				errs = append(errs, fmt.Sprintf("[w.fp.close] err=%v; w=%v", err, w))
			}
			w.fp = nil
		}
		f, err := os.OpenFile(w.filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err == nil {
			w.fp = f
		} else {
			errs = append(errs, fmt.Sprintf("[OpenFile] err=%v; w=%v", err, w))
		}
		w.lock.Unlock()
		if len(errs) > 0 {
			email.SendServerMail("rotate quit error", fmt.Sprintf(body, time.Now().Format(time.RFC3339), syscall.Getpid(), w, layout, duration, errs), "text/plain")
		}
	}()

	if layout == "" {
		layout = "2006-01-02"
		duration = time.Hour * 24
	}

	now := time.Now()
	if fs, err := os.Stat(w.filename); err == nil {
		if fs.Size() > 0 {
			fmt.Printf("%s, %v [rotate file] w=%v, size=%v\n", now.Format(time.RFC3339), syscall.Getpid(), w, fs.Size())
			w.lastRotate = now.UTC().Round(duration)
			lastFile := w.filename + "." + w.lastRotate.Add(-1*duration).Format(layout)
			err = os.Rename(w.filename, lastFile)
			if err != nil {
				return errors.New(fmt.Sprintf("[rename] err=%v; src=%s, target=%s, w=%v", err, w.filename, lastFile, w))
			}
			ftd := w.validFiles.Add(lastFile)
			if ftd != "" {
				err := os.Remove(ftd)
				if err != nil {
					w.validFiles.Reset(w.replica)
					fn := w.filename + ".*"
					matches, err := filepath.Glob(fn)
					if err != nil {
						return err
					}
					sort.Strings(matches) // 字典序从小到大(即日期从早到晚)
					for _, item := range matches {
						ftd := w.validFiles.Add(item)
						if ftd != "" {
							os.Remove(ftd)
						}
					}
					return errors.New(fmt.Sprintf("[remove] 重置 validFiles=%v", w))
				}
			}
		} else {
			fmt.Println(now.Format(time.RFC3339), syscall.Getpid(), "file empty", w.filename)
		}
	} else {
		return errors.New(fmt.Sprintf("[os.Stat] err=%v; w=%v", err, w))
	}
	return nil
}

func (w *RotateFile) Write(output []byte) (int, error) {
	//fmt.Println("write file")
	w.lock.Lock()
	defer w.lock.Unlock()
	size, err := w.fp.Write(output)
	return size, err
}

var body = `
now = %v, pid = %v

w = %v

layout=%s, duration=%s

err = %v
`

func RotateWriter(filename string, layout string, duration time.Duration, replica int) (*RotateFile, error) {
	w := &RotateFile{filename: filename, replica: replica}
	err := w.Reset(layout, duration)
	if err != nil {
		return nil, err
	}
	go func(w *RotateFile, layout string, duration time.Duration) {
		// 刚运行时, 发现有文件, 则认为应该接着往里写.
		//err := w.Rotate(layout, duration)
		//if err != nil {
		//	email.SendServerMail("rotate error", fmt.Sprintf(body, time.Now().Format(time.RFC3339), syscall.Getpid(), w, layout, duration, err), "text/plain")
		//}
		now := time.Now()
		timeToSleep := GetMinMoreThan(now, duration).Sub(now)
		time.Sleep(timeToSleep)
		for {
			err := w.Rotate(layout, duration)
			if err != nil {
				//fmt.Println(syscall.Getpid(), err)
				email.SendServerMail("rotate error", fmt.Sprintf(body, time.Now().Format(time.RFC3339), syscall.Getpid(), w, layout, duration, err), "text/plain")
			}
			time.Sleep(duration)
		}
	}(w, layout, duration)
	return w, nil
}

func GetMinMoreThan(t time.Time, duration time.Duration) time.Time {
	newt := t.Round(duration)
	if newt.Before(t) {
		newt = newt.Add(1 * duration)
	}
	return newt
}
