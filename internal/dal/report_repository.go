package dal

import (
	"database/sql"
	"frappuccino/models"
)

type ReportRepositoryInterface interface {
	TotalSales() (float64, error)
}

type ReportRepository struct {
	db *sql.DB
}

func NewReportRepository(_db *sql.DB) ReportRepository {
	return ReportRepository{db: _db}
}

func (r ReportRepository) TotalSales() (float64, error) {
	var orderClosed models.Order
	var totalSales float64
	queryTotalSales := `SELECT total_amount FROM orders WHERE status = 'closed'`
	rows, err := r.db.Query(queryTotalSales)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&orderClosed.TotalAmount); err != nil {
			return 0, err
		}
		totalSales += orderClosed.TotalAmount
	}

	if err := rows.Err(); err != nil {
		return 0, err
	}
	return totalSales, nil
}
