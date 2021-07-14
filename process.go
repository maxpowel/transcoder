package transcoder

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
)


func splitFunction(data []byte, atEOF bool) (advance int, token []byte, spliterror error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0:i], nil
	}
	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		// We have a cr terminated line
		return i + 1, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}

//Run a command and stream the otput
func run(ctx context.Context, status chan<- string, error chan <- error, command string, args []string) {
	cmd := exec.CommandContext(ctx, command, args...)

	out, err := cmd.StderrPipe()
	if err != nil {
		error<-err
		return
	}

	err = cmd.Start()
	if err != nil {
		error<-err
		return
	}

	scanner := bufio.NewScanner(out)
	scanner.Split(splitFunction)
	var lastLine string
	for scanner.Scan() {
		lastLine = scanner.Text()
		select {
		case <-ctx.Done():
			log.Println("Stopping command", command)
			error<-nil
			return
		case status <- lastLine:
		}
	}
	error <- fmt.Errorf("program exited: %s", lastLine)
}