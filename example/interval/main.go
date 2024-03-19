package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/haormj/fileshredder"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

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

	if err := fs.Run(ctx); err != nil {
		log.Fatalln(err)
	}
}
