package main
/*
Here we have two kind of concurrent jobs: the ffmpeg and the transcoder manager.
The ffmpeg jobs just run the ffmpeg command and stream the status.
The transcoder is in charge of managing all these processes (killing them, reading from them, creating them...).
This is why we have two WaitGroup (because the ffmpeg processes should die in first place, then the transcoder).
*/
import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"sync"
)

func NewTranscoder(ctx context.Context) *Transcoder {
	t := Transcoder{}
	t.Ctx = ctx
	t.TranscoderWaitGroup = sync.WaitGroup{}
	t.ProcessesWaitGroup = sync.WaitGroup{}
	t.Status = make(chan *StreamStatus)
	t.Error = make(chan error)
	t.ProcessesCancels = make(map[uuid.UUID]context.CancelFunc)
	t.TranscoderWaitGroup.Add(1)
	go t.run()
	return &t
}

type Transcoder struct {
	Ctx context.Context
	TranscoderWaitGroup sync.WaitGroup
	ProcessesWaitGroup sync.WaitGroup
	ProcessesCancels map[uuid.UUID]context.CancelFunc
	Status chan *StreamStatus
	Error chan error
}

func (t *Transcoder) Submit(args []string) (uuid.UUID, error) {
	t.ProcessesWaitGroup.Add(1)
	processUuid, err := uuid.NewUUID()
	if err != nil {
		return processUuid, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.ProcessesCancels[processUuid] = cancel
	go ffmpeg(ctx, &t.ProcessesWaitGroup, processUuid, t.Status, t.Error, args)
	return processUuid, nil
}

func (t *Transcoder) Stop(processUuid uuid.UUID) error {
	if cancel, ok := t.ProcessesCancels[processUuid]; ok {
		cancel()
		delete(t.ProcessesCancels, processUuid)
		return nil
	} else {
		return fmt.Errorf("process not found")
	}

}

func (t *Transcoder) run() {
	defer t.TranscoderWaitGroup.Done()
	log.Println("Transcoder ready, waiting for shutdown signal")
	<-t.Ctx.Done()
	log.Println("Transcoder shutting down...")
	for processUuid, cancel := range t.ProcessesCancels {
		log.Println("Stopping encoding:", processUuid)
		cancel()
	}
	log.Println("Waiting for all processes to stop...")
	t.ProcessesWaitGroup.Wait()
	log.Println("All processed stopped, bye")
}

func (t *Transcoder) Wait() {
	t.ProcessesWaitGroup.Wait()
}
