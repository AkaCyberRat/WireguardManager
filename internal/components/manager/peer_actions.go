package manager

import "context"

type Peer struct{}

// Add peer

type AddPeerParams struct{}

func (m *Manager) AddPeerAsync(ctx context.Context, params AddPeerParams) <-chan Result[Peer] {
	task := NewGenericTask(AddPeerParams{}, addPeerTask)

	return task.WaitAsync()
}

func addPeerTask(params AddPeerParams, resultCh chan<- Result[Peer], ctx taskContext) {

}

// Update peer

type UpdatePeerParams struct{}

func (m *Manager) UpdatePeer(ctx context.Context, params UpdatePeerParams) Result[Peer]

// Check peer

type CheckPeerParams struct {
}

func (m *Manager) CheckPeer(ctx context.Context, params AddPeerParams) <-chan Result[Peer]

// Delete peer

type DeletePeerParams struct{}

func (m *Manager) DeletePeer(ctx context.Context, params AddPeerParams) <-chan Result[Peer]
