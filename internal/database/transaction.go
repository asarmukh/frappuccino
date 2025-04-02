package database

import (
	"context"
	"database/sql"
	"log"
)

// TxFunc — функция, выполняемая в транзакции
type TxFunc func(tx *sql.Tx) error

// WithTransaction выполняет переданную функцию в рамках транзакции
func WithTransaction(db *sql.DB, fn TxFunc) error {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Если функция завершится с ошибкой, откатываем транзакцию
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			log.Printf("Recovered from panic in transaction: %v", p)
		}
	}()

	err = fn(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
