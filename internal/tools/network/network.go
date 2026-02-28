package network

import (
	"WireguardManager/internal/core"
)

type NetworkTool interface {
	EnableServer(peer *core.Server) error
	DisableServer() error

	EnablePeer(peer *core.Peer) error
	DisablePeer(peer *core.Peer) error

	GeneratePublicKey(privateKey string) (string, error)
	GeneratePrivateKey() (string, error)
}

type Tool struct{}

func NewNetworkTool(port int) *Tool {
	return nil
}

func (t *Tool) EnableServer(server *core.Server) error {
	return nil
}

func (t *Tool) DisableServer() error {
	return nil
}

func (t *Tool) EnablePeer(peer *core.Peer) error {
	return nil
}

func (t *Tool) DisablePeer(peer *core.Peer) error {
	return nil
}

func (t *Tool) GeneratePublicKey(privateKey string) (string, error) {
	return "", nil
}

func (t *Tool) GeneratePrivateKey() (string, error) {
	return "", nil

}

func GeneratePublicKey(privateKey string) (string, error) {
	return "", nil
}

func GeneratePrivateKey() (string, error) {
	return "", nil
}
