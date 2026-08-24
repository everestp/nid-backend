// modules/handle/repository/handle_repository.go
package repository

import (
	"database/sql"
	"nid-backend/models"
)

type HandleRepository struct {
	db *sql.DB
}

func NewHandleRepository(db *sql.DB) *HandleRepository {
	return &HandleRepository{db: db}
}

func (r *HandleRepository) Create(userID, name string) (*models.Handle, error) {
	query := `INSERT INTO handles (user_id, name, status, is_primary) VALUES ($1, $2, 'active', true) RETURNING id, user_id, name, status, is_primary, created_at`
	h := &models.Handle{}
	err := r.db.QueryRow(query, userID, name).Scan(&h.ID, &h.UserID, &h.Name, &h.Status, &h.Primary, &h.CreatedAt)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (r *HandleRepository) FindByName(name string) (*models.Handle, error) {
	query := `SELECT id, user_id, name, status, is_primary, created_at FROM handles WHERE name = $1`
	h := &models.Handle{}
	err := r.db.QueryRow(query, name).Scan(&h.ID, &h.UserID, &h.Name, &h.Status, &h.Primary, &h.CreatedAt)
	if err != nil {
		return nil, err
	}
	return h, nil
}
