package sqlite

import (
	"fmt"

	"WireguardManager/internal/repository"

	"gorm.io/gorm"
)

type TransactionManager struct {
	db         *gorm.DB
	txExecutor *gorm.DB
}

func NewSqliteTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

func (tm TransactionManager) Transaction(txFunc repository.InTransaction) (out interface{}, err error) {
	var tx *gorm.DB

	// Create transaction; new or from existing context(nested tx)
	if tm.txExecutor == nil {
		tx = tm.db.Begin()
	} else {
		tx = tm.txExecutor.Begin()
	}

	err = tx.Error
	if err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback().Error // TODO: Rollback err handling
			panic(p)                // re-throw panic after Rollback
		} else if err != nil {
			xerr := tx.Rollback().Error
			if xerr != nil {
				err = fmt.Errorf("tx err: %w \t rollback err: %w", err, xerr)
			}
		} else {
			err = tx.Commit().Error
		}
	}()

	contextTm := TransactionManager{
		db:         tm.db,
		txExecutor: tx,
	}

	out, err = txFunc(contextTm)
	return
}

func (tm TransactionManager) GetPeerRepository() repository.PeerRepository {
	if tm.txExecutor != nil {
		return NewSqlitePeerRepository(tm.txExecutor)
	}

	return NewSqlitePeerRepository(tm.db)
}
