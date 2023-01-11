package shell

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func Run(command string) (string, error) {
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
