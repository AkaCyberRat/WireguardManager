package manager

import (
	"WireguardManager/internal/tools"
	"WireguardManager/internal/tools/ipt"
	"WireguardManager/internal/tools/tc"
	"WireguardManager/internal/tools/wg"
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

type Manager struct {
	wgTool      wg.Tool
	tcTool      tc.Tool
	iptTool     ipt.Tool
	taskChan    chan ExecutableTask[taskContext]
	closingChan chan struct{}
}

func NewManager() (Manager, error) {
	var empty Manager

	wgTool, err := wg.NewTool()
	if err != nil {
		return empty, tools.FailedToCreateTool("wg", err)
	}

	tcTool, err := tc.NewTool()
	if err != nil {
		return empty, tools.FailedToCreateTool("tc", err)
	}

	iptTool, err := ipt.NewTool()
	if err != nil {
		return empty, tools.FailedToCreateTool("ipt", err)
	}

	manager := Manager{
		wgTool:      wgTool,
		tcTool:      tcTool,
		iptTool:     iptTool,
		taskChan:    make(chan ExecutableTask[taskContext], 1),
		closingChan: make(chan struct{}),
	}

	go manager.managerWorker()

	return manager, nil
}

func (m *Manager) Close() {
	close(m.closingChan)
	close(m.taskChan)
}

func (m *Manager) managerWorker() {
	ctx := taskContext{
		WgTool:  m.wgTool,
		TcTool:  m.tcTool,
		IptTool: m.iptTool,
	}

	for {
		select {
		case task := <-m.taskChan:
			// TODO: Add timout
			task.Execute(ctx)

		case <-m.closingChan:
			logrus.Info("Manager worker finished")
			return
		}
	}
}

func (m *Manager) sendTask(ctx context.Context, task ExecutableTask[taskContext]) error {
	select {
	case <-ctx.Done():
		return ErrorSendCanceled
	case m.taskChan <- task:
		return nil
	}
}

func executeTaskAsync[TParams any, TOut any](m *Manager, ctx context.Context, params TParams, execFunc TaskFunc[TParams, TOut, taskContext]) <-chan Result[TOut] {
	task := NewGenericTask(params, execFunc)

	err := m.sendTask(ctx, &task)
	if err != nil {
		task.Fail(err)
	}

	return task.WaitAsync()
}

type taskContext struct {
	WgTool  wg.Tool
	TcTool  tc.Tool
	IptTool ipt.Tool
}

var ErrorSendCanceled = fmt.Errorf("task sending was canceled by context")
