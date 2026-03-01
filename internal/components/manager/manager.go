package manager

import (
	"fmt"
)

type Manager struct {
	taskChan    chan task
	closingChan chan struct{}
}

func NewManager() (Manager, error) {

	return Manager{}, nil
}

// Use[T] возвращает <-chan Result[T] для вызывающей стороны
func Use[T any](m *Manager, c Command) <-chan Result[T] {
	resultCh := make(chan Result[T], 1)
	t := task{
		command: c,
		handle:  make(chan any, 1),
	}

	select {
	case m.taskChan <- t:
		go func() {
			defer close(resultCh)

			v := <-t.handle
			if r, ok := v.(Result[T]); ok {
				resultCh <- r
				return
			}

			resultCh <- Result[T]{Error: fmt.Errorf("unexpected result type")}
		}()

	case <-m.closingChan:
		var zero T
		resultCh <- Result[T]{Ok: zero, Error: fmt.Errorf("manager closed")}
		close(resultCh)
	}

	return resultCh
}

func (m *Manager) Close() {
	close(m.closingChan)
	close(m.taskChan)
}

func managerWorker() {}

// Helpers

type task struct {
	command Command
	handle  chan any
}

type Command struct {
	Params any
	Type   CommandType
}

type CommandType int

const (
	UpdateServerCommand CommandType = iota
	CheckServerCommand
	AddPeerCommand
	UpdatePeerCommand
	CheckPeerCommand
	RemovePeerCommand
)

type Result[T any] struct {
	Ok    T
	Error error
}
