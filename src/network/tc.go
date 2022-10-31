package network

import (
	"bytes"
	"context"
	"os/exec"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

func TcConfigure() error {
	var err error

	//
	//Add tc base rules to limit client download bandwidth (server upload)
	//
	err = Execute(time.Second, "tc", "qdisc", "add", "dev", "wg0", "root", "handle", "1:", "htb")
	if err != nil {
		return err
	}

	err = Execute(time.Second, "tc", "class", "add", "dev", "wg0", "parent", "1:", "classid", "1:1", "htb", "rate", "5mbit", "ceil", "5mbit")
	if err != nil {
		return err
	}

	//
	//Add tc rules to limit client upload bandwidth (server download)
	//
	err = Execute(time.Second, "tc", "qdisc", "add", "dev", "wg0", "ingress")
	if err != nil {
		return err
	}

	return nil
}

func TcAddRules(ipNet string, downloadSpeed int, uploadSpeed int, prio int, classId int) error {
	var err error

	downloadRate := strconv.FormatInt(int64(downloadSpeed), 10) + "mbit"
	//uploadRate := strconv.FormatInt(int64(uploadSpeed), 10) + "mbit"
	classIdValue := "1:" + strconv.FormatInt(int64(classId), 10)
	prioValue := strconv.FormatInt(int64(prio), 10)

	//
	//Add tc base rules to limit client download bandwidth (server upload)
	//
	err = Execute(time.Second, "tc", "class", "add", "dev", "wg0", "parent", "1:1", "classid", classIdValue, "htb", "rate", "4mbit", "ceil", downloadRate)
	if err != nil {
		return err
	}

	err = Execute(time.Second, "tc", "filter", "add", "dev", "wg0", "protocol", "ip", "parent", "1:", "prio", prioValue, "u32", "match", "ip", "src", ipNet, "flowid", classIdValue)
	if err != nil {
		return err
	}

	err = Execute(time.Second, "tc", "filter", "add", "dev", "wg0", "protocol", "ip", "parent", "1:", "prio", prioValue, "u32", "match", "ip", "dst", ipNet, "flowid", classIdValue)
	if err != nil {
		return err
	}

	//
	//Add tc rules to limit client upload bandwidth (server download)
	//
	err = Execute(time.Second, "tc", "filter", "add", "dev", "wg0", "protocol", "ip", "ingress", "prio", prioValue, "u32", "match", "ip", "src", ipNet, "action", "police", "rate", uploadRate, "burst", uploadRate)
	if err != nil {
		return err
	}

	err = Execute(time.Second, "tc", "filter", "add", "dev", "wg0", "protocol", "ip", "ingress", "prio", prioValue, "u32", "match", "ip", "dst", ipNet, "action", "police", "rate", uploadRate, "burst", uploadRate)
	if err != nil {
		return err
	}

	// err = Execute(time.Second, "tc", "filter", "add", "dev", "wg0", "protocol", "ip", "ingress", "prio", prioValue, "u32", "match", "ip", "src", ipNet, "flowid", classIdValue)
	// if err != nil {
	// 	return err
	// }

	// err = Execute(time.Second, "tc", "filter", "add", "dev", "wg0", "protocol", "ip", "ingress", "prio", prioValue, "u32", "match", "ip", "dst", ipNet, "flowid", classIdValue)
	// if err != nil {
	// 	return err
	// }

	return nil
}

// func TcDelRules(ipNet string, downloadSpeed int, uploadSpeed int, prio int, classId int) error {
// 	var err error

// 	downloadRate := strconv.FormatInt(int64(downloadSpeed), 10) + "mbit"
// 	uploadRate := strconv.FormatInt(int64(uploadSpeed), 10) + "mbit"
// 	classIdValue := "1:" + strconv.FormatInt(int64(classId), 10)
// 	prioValue := strconv.FormatInt(int64(prio), 10)

// 	err = execute(time.Second, )

// }

func Execute(duration time.Duration, name string, args ...string) (err error) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}

	var cmd *exec.Cmd
	if duration == 0 {
		cmd = exec.Command(name, args...)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		cmd = exec.CommandContext(ctx, name, args...)
	}

	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	err = cmd.Run()
	if err != nil {

		stdErrStr := stdErr.String()
		if len(stdErrStr) > 0 {
			logrus.Errorf("Command: '%v' HandledError: '%v' ErrorOut: '%v'", cmd.String(), err.Error(), stdErr.String())
		} else {
			logrus.Errorf("Command: '%v' HandledError: '%v'", cmd.String(), err.Error())
		}

		return err
	}

	logrus.Debugf("Command: '%v' Out: '%v'", cmd.String(), stdOut.String())

	return nil
}
