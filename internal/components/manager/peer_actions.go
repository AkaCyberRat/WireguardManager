package manager

import (
	"WireguardManager/internal/tools/wg"
	"context"
	"net"
)

type Peer struct{}

// Add peer

type AddPeerParams struct {
	// Ip is the IP address of peer (e.g. '11.0.0.2')
	Ip net.IP
	// PublicKey is the public key of peer
	PublicKey string
	// PreSharedKey is the pre-shared key of peer (optional)
	PreSharedKey *string
}

func (m *Manager) AddPeerAsync(ctx context.Context, params AddPeerParams) <-chan Result[Peer] {
	return executeTaskAsync(m, ctx, params, addPeerTask)
}

func addPeerTask(params AddPeerParams, resultCh chan<- Result[Peer], ctx taskContext) {
	ctx.WgTool.AddWgPeer(wg.PeerParams{})
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
