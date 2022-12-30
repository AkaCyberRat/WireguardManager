package repository

type TransactionManager interface {
	Transaction(txFunc InTransaction) (out interface{}, err error)

	GetPeerRepository() PeerRepository
}

type InTransaction func(tm TransactionManager) (interface{}, error)
