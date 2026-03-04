package manager

type Server struct{}

// Create server

type CreateServerParams struct{}

func (m *Manager) createServer(params UpdatePeerParams) <-chan Result[Server]

// Update server

type UpdateServerParams struct{}

func (m *Manager) UpdateServer(params UpdatePeerParams) <-chan Result[Server]

// Check server

type CheckServerParams struct{}

func (m *Manager) CheckServer(params AddPeerParams) <-chan Result[Server]
