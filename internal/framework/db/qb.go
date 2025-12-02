package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// QueryBuilder is a simple query builder for common SQL operations
type QueryBuilder struct {
	table   string
	columns []string
	where   []string
	args    []any
	orderBy string
	limit   *int
	offset  *int
}

// NewQB creates a new QueryBuilder for the given table
func NewQB(table string) QueryBuilder {
	return QueryBuilder{
		table:   table,
		columns: []string{"*"},
		where:   make([]string, 0),
		args:    make([]any, 0),
	}
}

// Select sets the columns to select
func (qb QueryBuilder) Select(columns ...string) QueryBuilder {
	qb.columns = columns
	return qb
}

// WhereEq adds an equality condition
func (qb QueryBuilder) WhereEq(column string, value any) QueryBuilder {
	qb.where = append(qb.where, fmt.Sprintf("%s = ?", column))
	qb.args = append(qb.args, value)
	return qb
}

// WhereIn adds an IN condition
func (qb QueryBuilder) WhereIn(column string, values []any) QueryBuilder {
	if len(values) == 0 {
		return qb
	}
	placeholders := strings.Repeat("?,", len(values))
	placeholders = placeholders[:len(placeholders)-1]
	qb.where = append(qb.where, fmt.Sprintf("%s IN (%s)", column, placeholders))
	qb.args = append(qb.args, values...)
	return qb
}

// OrderBy sets the ORDER BY clause
func (qb QueryBuilder) OrderBy(expr string) QueryBuilder {
	qb.orderBy = expr
	return qb
}

// Limit sets the LIMIT
func (qb QueryBuilder) Limit(n int) QueryBuilder {
	qb.limit = &n
	return qb
}

// Offset sets the OFFSET
func (qb QueryBuilder) Offset(n int) QueryBuilder {
	qb.offset = &n
	return qb
}

// BuildSelect builds a SELECT query
func (qb QueryBuilder) BuildSelect() (string, []any) {
	var sb strings.Builder
	sb.WriteString("SELECT ")
	sb.WriteString(strings.Join(qb.columns, ", "))
	sb.WriteString(" FROM ")
	sb.WriteString(qb.table)

	if len(qb.where) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(qb.where, " AND "))
	}

	if qb.orderBy != "" {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(qb.orderBy)
	}

	if qb.limit != nil {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", *qb.limit))
	}

	if qb.offset != nil {
		sb.WriteString(fmt.Sprintf(" OFFSET %d", *qb.offset))
	}

	return sb.String(), qb.args
}

// BuildCount builds a COUNT query
func (qb QueryBuilder) BuildCount() (string, []any) {
	var sb strings.Builder
	sb.WriteString("SELECT COUNT(*) FROM ")
	sb.WriteString(qb.table)

	if len(qb.where) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(qb.where, " AND "))
	}

	return sb.String(), qb.args
}

// Helper Functions (10 common patterns)

// FindByID retrieves a single row by ID
func FindByID(ctx context.Context, table string, id any, dest ...any) error {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", table)
	err := DB.QueryRowContext(ctx, query, id).Scan(dest...)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	return err
}

// FindOneBy retrieves a single row matching conditions
func FindOneBy(ctx context.Context, table string, conditions map[string]any, dest ...any) error {
	qb := NewQB(table)
	for col, val := range conditions {
		qb = qb.WhereEq(col, val)
	}
	qb = qb.Limit(1)

	query, args := qb.BuildSelect()
	err := DB.QueryRowContext(ctx, query, args...).Scan(dest...)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	return err
}

// FindAll retrieves all rows from a table
func FindAll(ctx context.Context, table string, orderBy string) (*sql.Rows, error) {
	qb := NewQB(table)
	if orderBy != "" {
		qb = qb.OrderBy(orderBy)
	}
	query, args := qb.BuildSelect()
	return DB.QueryContext(ctx, query, args...)
}

// FindManyBy retrieves multiple rows matching conditions
func FindManyBy(ctx context.Context, table string, conditions map[string]any, orderBy string, limit, offset int) (*sql.Rows, error) {
	qb := NewQB(table)
	for col, val := range conditions {
		qb = qb.WhereEq(col, val)
	}
	if orderBy != "" {
		qb = qb.OrderBy(orderBy)
	}
	if limit > 0 {
		qb = qb.Limit(limit)
	}
	if offset > 0 {
		qb = qb.Offset(offset)
	}

	query, args := qb.BuildSelect()
	return DB.QueryContext(ctx, query, args...)
}

// Insert inserts a new row and returns the last insert ID
func Insert(ctx context.Context, table string, data map[string]any) (int64, error) {
	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]any, 0, len(data))

	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		values = append(values, val)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	result, err := DB.ExecContext(ctx, query, values...)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// UpdateByID updates a row by ID
func UpdateByID(ctx context.Context, table string, id any, data map[string]any) error {
	if len(data) == 0 {
		return nil
	}

	sets := make([]string, 0, len(data))
	values := make([]any, 0, len(data)+1)

	for col, val := range data {
		sets = append(sets, fmt.Sprintf("%s = ?", col))
		values = append(values, val)
	}
	values = append(values, id)

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = ?",
		table,
		strings.Join(sets, ", "),
	)

	result, err := DB.ExecContext(ctx, query, values...)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteByID deletes a row by ID
func DeleteByID(ctx context.Context, table string, id any) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)
	result, err := DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// SoftDeleteByID soft deletes a row by setting deleted_at
func SoftDeleteByID(ctx context.Context, table string, id any) error {
	query := fmt.Sprintf("UPDATE %s SET deleted_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL", table)
	result, err := DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// CountWhere counts rows matching conditions
func CountWhere(ctx context.Context, table string, conditions map[string]any) (int, error) {
	qb := NewQB(table)
	for col, val := range conditions {
		qb = qb.WhereEq(col, val)
	}

	query, args := qb.BuildCount()
	var count int
	err := DB.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// ExistsWhere checks if any row exists matching conditions
func ExistsWhere(ctx context.Context, table string, conditions map[string]any) (bool, error) {
	count, err := CountWhere(ctx, table, conditions)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
