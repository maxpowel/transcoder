# transcoder
I wanted to play with goroutines in Go so I created this program.
Basically, this program uses `Context`, `WaitGroup` and `goroutines` to handle different
FFMPEG processes in background and the same time. It also parses the output and provides
the status and errors through channels.

This is an example about how to use it:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

//
func monitor(ctx context.Context, t *Transcoder) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Stop monitor")
			return
		case line := <-t.Status:
			fmt.Println("DATA", line)

		case line := <-t.Error:
			fmt.Println("ERROR", line)
		}
	}
}

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	t := NewTranscoder(ctx)
	go monitor(ctx, t)
	args := []string{
		"-y", "-i", "my_file.mov",
		"-c:v", "h264",
		"-c:a", "mp3",
		"-hls_time",
		"1000",
		"-hls_wrap",
		"100",
		"output.mp4"}
	uuid, err := t.Submit(args)
	log.Println("NEW PROCESS", uuid, err)
	args[11] = "output2.mp4"
	uuid, err = t.Submit(args)
	log.Println("OTHER PROCESS", uuid, err)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	s := <-c
	log.Println("SIGNAL", s)
	cancel()
	t.Wait()
}
```
