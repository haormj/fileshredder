package fileshredder

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type FileShredder struct {
	options Options
	quit    chan struct{}
}

func NewFileShredder(opts ...Option) (*FileShredder, error) {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	f := &FileShredder{
		options: options,
		quit:    make(chan struct{}),
	}

	return f, nil
}

func (f *FileShredder) Run(ctx context.Context) error {
	if f.options.Interval == 0 {
		return errors.New("interval is 0")
	}

	t := time.NewTicker(f.options.Interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-f.quit:
			return nil
		case <-t.C:
			if err := f.MillRunOnce(); err != nil {
				log.Printf("MillRunOnce error, err: %v\n", err)
			}
		}
	}
}

func (f *FileShredder) Close() error {
	select {
	case <-f.quit:
	default:
		close(f.quit)
	}

	return nil
}

func (f *FileShredder) MillRunOnce(opts ...MillRunOnceOption) error {
	options := NewMillRunOnceOptions()
	for _, o := range opts {
		o(&options)
	}

	// 如果没有配置任何限制，直接返回
	if f.options.MaxSize <= 0 && f.options.MaxAge <= 0 && f.options.MaxCount <= 0 {
		return nil
	}

	matches, err := filepath.Glob(f.options.GlobPath)
	if err != nil {
		return err
	}

	var fileInfos []FileInfo
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			return err
		}
		fileInfos = append(fileInfos, FileInfo{
			Path:     match,
			FileInfo: info,
		})
	}

	sort.Sort(fileInfoList(fileInfos))

	var remove []FileInfo
	var size int64

	if f.options.MaxSize > 0 {
		var remaining []FileInfo
		for _, fileInfo := range fileInfos {
			if fileInfo.IsDir() {
				s, err := getDirSize(fileInfo.Path)
				if err != nil {
					return err
				}
				size += int64(s)
			} else {
				size += fileInfo.Size()
			}
			if size > f.options.MaxSize {
				remove = append(remove, fileInfo)
			} else {
				remaining = append(remaining, fileInfo)
			}
		}
		fileInfos = remaining
	}

	if f.options.MaxAge > 0 {
		var remaining []FileInfo
		for i, fileInfo := range fileInfos {
			if time.Since(fileInfo.ModTime()) > f.options.MaxAge {
				remove = append(remove, fileInfos[i:]...)
				break
			}
			remaining = append(remaining, fileInfo)
		}
		fileInfos = remaining
	}

	if f.options.MaxCount > 0 && int64(len(fileInfos)) > f.options.MaxCount {
		remove = append(remove, fileInfos[f.options.MaxCount:]...)
		// 若后续还有其他策略的话，需要更新 fileInfos
		// fileInfos = fileInfos[:r.maxCount]
	}

	for _, fileInfo := range remove {
		if options.IsNotDelete(&fileInfo) {
			continue
		}
		if err := os.RemoveAll(fileInfo.Path); err != nil {
			return err
		}
	}

	return nil
}

type FileInfo struct {
	fs.FileInfo
	Path string
}

type fileInfoList []FileInfo

func (l fileInfoList) Len() int {
	return len(l)
}

func (l fileInfoList) Less(i, j int) bool {
	return l[i].ModTime().After(l[j].ModTime())
}

func (l fileInfoList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func getDirSize(path string) (uint64, error) {
	var totalSize uint64

	err := filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			totalSize += uint64(info.Size())
		}

		return nil
	})

	if err == nil {
		return totalSize, nil
	}

	if os.IsNotExist(err) {
		return 0, nil
	}

	return 0, err
}
