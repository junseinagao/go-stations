package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/TechBowl-japan/go-stations/model"
	"github.com/mattn/go-sqlite3"
)

// A TODOService implements CRUD of TODO entities.
type TODOService struct {
	db *sql.DB
}

// NewTODOService returns new TODOService.
func NewTODOService(db *sql.DB) *TODOService {
	return &TODOService{
		db: db,
	}
}

// CreateTODO creates a TODO on DB.
func (s *TODOService) CreateTODO(ctx context.Context, subject, description string) (*model.TODO, error) {
	const (
		insert  = `INSERT INTO todos(subject, description) VALUES(?, ?)`
		confirm = `SELECT subject, description, created_at, updated_at FROM todos WHERE id = ?`
	)
	var todo model.TODO
	if subject == "" {
		return &todo, sqlite3.ErrConstraint
	}

	stmt, err := s.db.PrepareContext(ctx, insert)
	result, err := stmt.ExecContext(ctx, subject, description)
	todo.ID, err = result.LastInsertId()
	row := s.db.QueryRowContext(ctx, confirm, todo.ID)
	err = row.Scan(&todo.Subject, &todo.Description, &todo.CreatedAt, &todo.UpdatedAt)

	return &todo, err
}

// ReadTODO reads TODOs on DB.
func (s *TODOService) ReadTODO(ctx context.Context, prevID, size int64) ([]*model.TODO, error) {
	const (
		read       = `SELECT id, subject, description, created_at, updated_at FROM todos ORDER BY id DESC LIMIT ?`
		readWithID = `SELECT id, subject, description, created_at, updated_at FROM todos WHERE id < ? ORDER BY id DESC LIMIT ?`
	)
	todos := make([]*model.TODO, 0)

	// * size がない場合は、空で返却する
	if size == 0 {
		return todos, nil
	}

	switch prevID {
	case 0:
		rows, err := s.db.QueryContext(ctx, read, size)
		defer rows.Close()
		if err != nil {
			return todos, err
		}
		for rows.Next() {
			var todo model.TODO
			err = rows.Scan(&todo.ID, &todo.Subject, &todo.Description, &todo.CreatedAt, &todo.UpdatedAt)
			if err != nil {
				return todos, err
			}
			todos = append(todos, &todo)
		}
		return todos, err
	default:
		rows, err := s.db.QueryContext(ctx, readWithID, prevID, size)
		defer rows.Close()
		if err != nil {
			return todos, err
		}
		for rows.Next() {
			var todo model.TODO
			err = rows.Scan(&todo.ID, &todo.Subject, &todo.Description, &todo.CreatedAt, &todo.UpdatedAt)
			if err != nil {
				return todos, err
			}
			todos = append(todos, &todo)
		}
		return todos, err
	}
}

// UpdateTODO updates the TODO on DB.
func (s *TODOService) UpdateTODO(ctx context.Context, id int64, subject, description string) (*model.TODO, error) {
	const (
		update  = `UPDATE todos SET subject = ?, description = ? WHERE id = ?`
		confirm = `SELECT subject, description, created_at, updated_at FROM todos WHERE id = ?`
	)

	var todo model.TODO
	todo.ID = id

	if subject == "" {
		return &todo, sqlite3.ErrConstraint
	}
	stmt, err := s.db.PrepareContext(ctx, update)
	if err != nil {
		return &todo, err
	}
	result, err := stmt.ExecContext(ctx, subject, description, id)
	if err != nil {
		return &todo, err
	}
	rowId, err := result.RowsAffected()
	if rowId == 0 {
		return &todo, &model.ErrNotFound{
			When: time.Now(),
			What: "Updated row not found",
		}
	}
	if err != nil {
		return &todo, err
	}

	row := s.db.QueryRowContext(ctx, confirm, id)
	err = row.Scan(&todo.Subject, &todo.Description, &todo.CreatedAt, &todo.UpdatedAt)

	return &todo, err
}

// DeleteTODO deletes TODOs on DB by ids.
func (s *TODOService) DeleteTODO(ctx context.Context, ids []int64) error {
	const deleteFmt = `DELETE FROM todos WHERE id IN (?%s)`
	delete := fmt.Sprintf(deleteFmt, strings.Repeat(",?", len(ids)-1))

	// * ids がない場合は、nil で返却する
	if len(ids) == 0 {
		return nil
	}

	strIds := make([]interface{}, len(ids))
	for i, id := range ids {
		strIds[i] = fmt.Sprintf("%d", id)
	}
	stmt, err := s.db.PrepareContext(ctx, delete)
	if err != nil {
		return err
	}
	result, err := stmt.ExecContext(ctx, strIds...)
	if err != nil {
		return err
	}
	nums, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if nums == 0 {
		return &model.ErrNotFound{
			When: time.Now(),
			What: "Deleted row not found",
		}
	}
	return err
}
