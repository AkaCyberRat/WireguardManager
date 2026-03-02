package manager

import (
	"WireguardManager/internal/tools"
	"WireguardManager/internal/tools/ipt"
	"WireguardManager/internal/tools/tc"
	"WireguardManager/internal/tools/wg"

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
		Field: 1,
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

type taskContext struct {
	Field int
}
