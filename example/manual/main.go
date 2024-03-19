package main

import (
	"log"
	"time"

	"github.com/haormj/fileshredder"
)

func main() {
	fs, err := fileshredder.NewFileShredder(
		fileshredder.GlobPath("/path/to/glob/"),
		fileshredder.Interval(time.Minute),
		fileshredder.MaxAge(time.Hour),
		fileshredder.MaxSize(10*1024*1024),
		fileshredder.MaxCount(100),
	)
	if err != nil {
		log.Fatalln(err)
	}

	// 调用者自己决定调用时机
	if err := fs.MillRunOnce(); err != nil {
		log.Fatalln(err)
	}
}
