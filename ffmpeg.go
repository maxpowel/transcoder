package main

import (
	"context"
	"github.com/google/uuid"
	"log"
	"regexp"
	"strings"
	"sync"
)

//A single ffmpeg process. It just listens for a stop event or new ffmpeg data.
func ffmpeg(ctx context.Context, wg *sync.WaitGroup, uuid uuid.UUID, status chan<-*StreamStatus, error chan <- error, args []string) {
	defer wg.Done()

	log.Println("Booting ffmpeg process")
	ch := make(chan string)
	defer close(ch)
	streamStatus := &StreamStatus{}
	streamStatus.Uuid = uuid
	go run(ctx, ch, error,"ffmpeg", args)
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down ffmepg process, waiting for result")
			return
		case line := <-ch:
			//We are only interested on the transcoding status information
			if strings.HasPrefix(line, "frame") {
				parseLine(line, streamStatus)
				status <- streamStatus
			}
		}

	}
}


type StreamStatus struct {
	Uuid    uuid.UUID `json:"uuid"`
	Frame   string    `json:"frame"`
	Fps     string    `json:"fps"`
	Quality string    `json:"quality"`
	Time    string    `json:"time"`
	Bitrate string    `json:"bitrate"`
	Speed   string    `json:"speed"`
	Error   error     `json:"error"`
}

func parseLine(line string, status *StreamStatus) {
	// Stream
	// frame=  267 fps= 45 q=-1.0 size=N/A time=00:00:10.65 bitrate=N/A speed= 1.8x
	//r, err := regexp.Compile("frame=([0-9]+) fps=([0-9]+) q=([0-9]+\\.[0-9]+) size=    ([0-9]+[A-Za-z]+) time=([0-9]+:[0-9]+:[0-9]+)\\.[0-9]+ bitrate=N/A speed=([0-9]+\\.[0-9]+)x")
	//res := make([]string, 7)
	m := regexp.MustCompile("[ ]{2,}")
	line = m.ReplaceAllString(line, "")
	line = strings.ReplaceAll(line, "= ", "=")
	elems := strings.Split(line, " ")
	for _, part := range elems {
		p := strings.Split(part, "=")
		switch p[0] {
		case "frame":
			status.Frame = p[1]
		case "fps":
			status.Fps = p[1]
		case "q":
			status.Quality = p[1]
		case "time":
			status.Time = p[1]
		case "bitrate":
			status.Bitrate = p[1]
		case "speed":
			status.Speed = p[1]
		}
	}
}
