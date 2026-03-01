package manager

import "context"

type Peer struct{}

// Add peer

type AddPeerParams struct{}

func (m *Manager) AddPeer(ctx context.Context, params AddPeerParams) <-chan Result[Peer] {
	return Use[Peer](m, Command{Params: params, Type: AddPeerCommand})
}

// Update peer

type UpdatePeerParams struct{}

func (m *Manager) UpdatePeer(ctx context.Context, params UpdatePeerParams) <-chan Result[Peer]

// Check peer

type CheckPeerParams struct {
}

func (m *Manager) CheckPeer(ctx context.Context, params AddPeerParams) <-chan Result[Peer]

// Delete peer

type DeletePeerParams struct{}

func (m *Manager) DeletePeer(ctx context.Context, params AddPeerParams) <-chan Result[Peer]
