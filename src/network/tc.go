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
	//
	//Add tc base rule to limit client download bandwidth (server upload)
	//
	err := execute("tc", "qdisc", "add", "dev", "wg0", "root", "handle", "1:", "htb")
	if err != nil {
		return err
	}

	//
	//Add tc base rule to limit client upload bandwidth (server download)
	//
	err = execute("tc", "qdisc", "add", "dev", "wg0", "ingress")
	if err != nil {
		return err
	}

	return nil
}

func TcAddRules(ipNet string, downloadSpeed int, uploadSpeed int, id int) error {
	_downloadSpeed := strconv.FormatInt(int64(downloadSpeed), 10) + "mbit"
	_uploadSpeed := strconv.FormatInt(int64(uploadSpeed), 10) + "mbit"
	prio := strconv.FormatInt(int64(id), 10)
	classId := "1:" + strconv.FormatInt(int64(id), 10)

	//
	// Limit download bandwidth
	//
	err := execute("tc", "class", "add", "dev", "wg0", "parent", "1:", "classid", classId, "htb", "rate", "2mbit", "ceil", _downloadSpeed)
	if err != nil {
		return err
	}

	err = execute("tc", "filter", "add", "dev", "wg0", "protocol", "ip", "parent", "1:", "prio", prio, "u32", "match", "ip", "src", ipNet, "flowid", classId)
	if err != nil {
		return err
	}

	err = execute("tc", "filter", "add", "dev", "wg0", "protocol", "ip", "parent", "1:", "prio", prio, "u32", "match", "ip", "dst", ipNet, "flowid", classId)
	if err != nil {
		return err
	}

	//
	// Limit upload bandwidth
	//
	err = execute("tc", "filter", "add", "dev", "wg0", "protocol", "ip", "ingress", "prio", prio, "u32", "match", "ip", "src", ipNet, "action", "police", "rate", _uploadSpeed, "burst", "5mbit")
	if err != nil {
		return err
	}

	err = execute("tc", "filter", "add", "dev", "wg0", "protocol", "ip", "ingress", "prio", prio, "u32", "match", "ip", "dst", ipNet, "action", "police", "rate", _uploadSpeed, "burst", "5mbit")
	if err != nil {
		return err
	}

	return nil
}

func TcDelRules(id int) error {
	prio := strconv.FormatInt(int64(id), 10)
	classId := "1:" + strconv.FormatInt(int64(id), 10)

	err := execute("tc", "filter", "del", "dev", "wg0", "parent", classId, "prio", prio)
	if err != nil {
		return err
	}

	err = execute("tc", "filter", "del", "dev", "wg0", "ingress", "prio", prio)
	if err != nil {
		return err
	}

	err = execute("tc", "class", "del", "dev", "wg0", "parent", classId)
	if err != nil {
		return err
	}

	return nil
}

func execute(name string, args ...string) (err error) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	err = cmd.Run()
	if err != nil {
		logrus.Errorf("Command: '%v' Error: '%v' '%v'", cmd.String(), err.Error(), stdErr.String())
		return err
	}

	logrus.Debugf("Command: '%v' Out: '%v'", cmd.String(), stdOut.String())

	return nil
}
