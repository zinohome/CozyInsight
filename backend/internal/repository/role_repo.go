package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/model"
)

type RoleRepository struct {
	db *sqlx.DB
}

func NewRoleRepository(db *sqlx.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) Create(ctx context.Context, role *model.Role) error {
	query := `INSERT INTO roles (name, code, description, status) VALUES (:name, :code, :description, :status)`
	result, err := r.db.NamedExecContext(ctx, query, role)
	if err != nil {
		return fmt.Errorf("create role failed: %w", err)
	}
	id, _ := result.LastInsertId()
	role.ID = uint64(id)
	return nil
}

func (r *RoleRepository) FindByID(ctx context.Context, id uint64) (*model.Role, error) {
	var role model.Role
	query := `SELECT * FROM roles WHERE id = ? AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &role, query, id); err != nil {
		return nil, fmt.Errorf("find role failed: %w", err)
	}
	return &role, nil
}

func (r *RoleRepository) List(ctx context.Context) ([]model.Role, error) {
	var list []model.Role
	query := `SELECT * FROM roles WHERE deleted_at IS NULL ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, fmt.Errorf("list roles failed: %w", err)
	}
	return list, nil
}

func (r *RoleRepository) Update(ctx context.Context, role *model.Role) error {
	query := `UPDATE roles SET name = :name, code = :code, description = :description, status = :status WHERE id = :id`
	if _, err := r.db.NamedExecContext(ctx, query, role); err != nil {
		return fmt.Errorf("update role failed: %w", err)
	}
	return nil
}

func (r *RoleRepository) Delete(ctx context.Context, id uint64) error {
	query := `UPDATE roles SET deleted_at = NOW() WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("delete role failed: %w", err)
	}
	return nil
}

func (r *RoleRepository) ListMenus(ctx context.Context) ([]model.Menu, error) {
	var list []model.Menu
	query := `SELECT * FROM menus WHERE status = 1 ORDER BY parent_id, sort`
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, fmt.Errorf("list menus failed: %w", err)
	}
	return list, nil
}

func (r *RoleRepository) SetRoleMenus(ctx context.Context, roleID uint64, menuIDs []uint64) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM role_menus WHERE role_id = ?`, roleID); err != nil {
		return fmt.Errorf("clear role menus failed: %w", err)
	}
	if len(menuIDs) == 0 {
		return nil
	}
	placeholders := make([]string, len(menuIDs))
	args := make([]interface{}, 0, len(menuIDs)*2)
	for i, menuID := range menuIDs {
		placeholders[i] = "(?, ?)"
		args = append(args, roleID, menuID)
	}
	query := "INSERT INTO role_menus (role_id, menu_id) VALUES " + strings.Join(placeholders, ", ")
	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("add role menus failed: %w", err)
	}
	return nil
}

func (r *RoleRepository) GetRoleMenus(ctx context.Context, roleID uint64) ([]uint64, error) {
	var ids []uint64
	query := `SELECT menu_id FROM role_menus WHERE role_id = ?`
	if err := r.db.SelectContext(ctx, &ids, query, roleID); err != nil {
		return nil, fmt.Errorf("get role menus failed: %w", err)
	}
	return ids, nil
}

func (r *RoleRepository) SetUserRoles(ctx context.Context, userID uint64, roleIDs []uint64) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id = ?`, userID); err != nil {
		return fmt.Errorf("clear user roles failed: %w", err)
	}
	if len(roleIDs) == 0 {
		return nil
	}
	placeholders := make([]string, len(roleIDs))
	args := make([]interface{}, 0, len(roleIDs)*2)
	for i, roleID := range roleIDs {
		placeholders[i] = "(?, ?)"
		args = append(args, userID, roleID)
	}
	query := "INSERT INTO user_roles (user_id, role_id) VALUES " + strings.Join(placeholders, ", ")
	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("add user roles failed: %w", err)
	}
	return nil
}

func (r *RoleRepository) GetUserRoles(ctx context.Context, userID uint64) ([]uint64, error) {
	var ids []uint64
	query := `SELECT role_id FROM user_roles WHERE user_id = ?`
	if err := r.db.SelectContext(ctx, &ids, query, userID); err != nil {
		return nil, fmt.Errorf("get user roles failed: %w", err)
	}
	return ids, nil
}
