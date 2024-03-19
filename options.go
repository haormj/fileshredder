package fileshredder

import (
	"time"
)

type Options struct {
	MaxAge   time.Duration
	MaxCount int64
	MaxSize  int64
	GlobPath string
	Interval time.Duration
}

type Option func(*Options)

func MaxAge(age time.Duration) Option {
	return func(o *Options) {
		o.MaxAge = age
	}
}

func MaxCount(count int64) Option {
	return func(o *Options) {
		o.MaxCount = count
	}
}

func MaxSize(size int64) Option {
	return func(o *Options) {
		o.MaxSize = size
	}
}

func GlobPath(path string) Option {
	return func(o *Options) {
		o.GlobPath = path
	}
}

func Interval(d time.Duration) Option {
	return func(o *Options) {
		o.Interval = d
	}
}

type MillRunOnceOptions struct {
	IsNotDelete IsNotDeleteFunc
}

type MillRunOnceOption func(*MillRunOnceOptions)

type IsNotDeleteFunc func(info *FileInfo) bool

func NewMillRunOnceOptions() MillRunOnceOptions {
	return MillRunOnceOptions{
		IsNotDelete: defaultIsNotDelete,
	}
}

func defaultIsNotDelete(info *FileInfo) bool {
	return false
}

func IsNotDelete(fn IsNotDeleteFunc) MillRunOnceOption {
	return func(o *MillRunOnceOptions) {
		o.IsNotDelete = fn
	}
}
