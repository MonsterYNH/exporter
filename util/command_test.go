package util

import (
	"context"
	"fmt"
	"testing"
)

func Test1(t *testing.T) {

	cmd := `date -d "$(awk -F. '{print $1}' /proc/uptime) second ago" +"%Y-%m-%d %H:%M:%S"`

	bytes, err := ExecCommand(context.Background(), "bash", "-c", cmd)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(bytes))
}
