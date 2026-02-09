package shell

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func RunExecWithTimeout(command string) (string, error) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	args := strings.Split(command, " ")

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	err := cmd.Run()
	strOut := stdOut.String()
	strErr := stdErr.String()

	if err != nil {
		logrus.Errorf("Command: '%v' Error: '%v' '%v'", command, err.Error(), strErr)
		return strOut, err
	}

	logrus.Tracef("Command: '%v' Out: '%v'", cmd.String(), strOut)
	return strOut, nil
}

func RunBlockingExec(cmd string, args ...any) error {
	strArgs := make([]string, len(args))
	for i, a := range args {
		switch v := a.(type) {
		case string:
			strArgs[i] = v
		case fmt.Stringer:
			strArgs[i] = v.String()
		default:
			strArgs[i] = fmt.Sprint(v)
		}
	}

	c := exec.Command(cmd, strArgs...)
	out, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v failed: %v\n%s", cmd, strArgs, err, out)
	}
	return nil
}
