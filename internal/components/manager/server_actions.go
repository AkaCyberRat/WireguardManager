package manager

type Server struct{}

// Update server

type UpdateServerParams struct{}

func (m *Manager) UpdateServer(params UpdatePeerParams) <-chan Result[Server]

// Check server

type CheckServerParams struct{}

func (m *Manager) CheckServer(params AddPeerParams) <-chan Result[Server]
