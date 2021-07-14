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

//This function will run in background consuming and displaying the FFMPEG status information
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

	// The context to control when the transcoding process should stop
	ctx, cancel := context.WithCancel(context.Background())
	t := NewTranscoder(ctx)
	go monitor(ctx, t)
	// Some example parameters
	args := []string{
		"-y", "-i", "my_file.mov",
		"-c:v", "h264",
		"-c:a", "mp3",
		"-hls_time",
		"1000",
		"-hls_wrap",
		"100",
		"output.mp4"}
	// Send now transcoding job using the "Submit" method
	uuid, err := t.Submit(args)
	log.Println("NEW PROCESS", uuid, err)
	// You can stop it at any moment
	//err = t.Stop(uuid)

	args[11] = "output2.mp4"
	// Send another job
	uuid, err = t.Submit(args)
	log.Println("OTHER PROCESS", uuid, err)
	// To stop the processing when a signal is received
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	s := <-c
	log.Println("SIGNAL", s)
	//Once we receive the signal to stop, we cancel all transcoding jobs and wait until all is stopped
	cancel()
	t.Wait()
}
```
