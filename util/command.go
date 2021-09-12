package util

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sync"
)

var (
	bufferPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer([]byte{})
		},
	}
)

func ExecCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, name, args...)

	stdOut := bufferPool.Get().(*bytes.Buffer)
	stdErr := bufferPool.Get().(*bytes.Buffer)

	defer func() {
		stdOut.Reset()
		stdErr.Reset()
		bufferPool.Put(stdErr)
		bufferPool.Put(stdOut)
	}()

	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("command %s %v start failed, error: %s %s", path, args, err.Error(), stdErr.String())
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("command %s %v exec failed, error: %s %s", path, args, err.Error(), stdErr.String())
	}

	return stdOut.Bytes(), nil
}
