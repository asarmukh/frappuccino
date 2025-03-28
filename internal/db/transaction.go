package db

import (
	"database/sql"
	"log"
)

// TransactionManager управляет транзакциями
type TransactionManager struct {
	db *sql.DB
	tx *sql.Tx
}

// NewTransactionManager создает новый менеджер транзакций
func NewTransactionManager(db *sql.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// Begin начинает новую транзакцию
func (tm *TransactionManager) Begin() error {
	tx, err := tm.db.Begin()
	if err != nil {
		return err
	}
	tm.tx = tx
	return nil
}

// Commit фиксирует транзакцию
func (tm *TransactionManager) Commit() error {
	if tm.tx == nil {
		return nil
	}
	err := tm.tx.Commit()
	if err != nil {
		log.Println("Commit error:", err)
	}
	tm.tx = nil
	return err
}

// Rollback откатывает транзакцию
func (tm *TransactionManager) Rollback() {
	if tm.tx != nil {
		_ = tm.tx.Rollback()
		log.Println("Transaction rolled back")
		tm.tx = nil
	}
}

// Exec выполняет SQL-запрос внутри транзакции
func (tm *TransactionManager) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tm.tx.Exec(query, args...)
}

// QueryRow выполняет SQL-запрос, который возвращает одну строку
func (tm *TransactionManager) QueryRow(query string, args ...interface{}) *sql.Row {
	return tm.tx.QueryRow(query, args...)
}

// Query возвращает строки
func (tm *TransactionManager) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tm.tx.Query(query, args...)
}

// WithTransaction выполняет функцию внутри транзакции
func (tm *TransactionManager) WithTransaction(fn func(tx *TransactionManager) error) (err error) {
	err = tm.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tm.Rollback()
			log.Println("Transaction panicked and rolled back:", p)
			panic(p)
		} else if err != nil {
			tm.Rollback()
		} else {
			err = tm.Commit()
		}
	}()
	return fn(tm)
}
