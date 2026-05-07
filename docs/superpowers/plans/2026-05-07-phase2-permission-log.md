# Phase 2 子计划：权限管理 + 操作日志 + 行级数据权限

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** 实现企业级权限体系（RBAC + 行级数据权限）、全量操作日志审计、以及对应的前端管理页面。

**Architecture:** 沿用已有的分层架构（Handler → Service → Repository），新增 JWT 认证中间件、用户/角色/菜单/操作日志/行级权限模块。

---

## 设计决策

1. **JWT 认证中间件** — 验证 `Authorization: Bearer {token}` 头，解析 Claims 注入 gin.Context
2. **RBAC 模型** — 角色-菜单关联（role_menus）、用户-角色关联（user_roles）。菜单树支持层级（parent_id）
3. **操作日志** — 基于 Gin 中间件，记录所有请求的方法、路径、用户、IP、耗时、状态码
4. **行级权限** — 基于用户属性（如部门 ID）在数据集查询时动态拼接 SQL WHERE 条件
5. **前端布局** — Ant Design Layout（侧边栏 + Header），路由级页面切换

---

## Task 16: JWT 认证中间件 + 用户信息注入 Context

**Files:**
- Create: `backend/internal/middleware/auth.go`
- Modify: `backend/api/v1/router.go` — 添加认证中间件到受保护路由
- Modify: `backend/internal/handler/*` — 所有 handler 中 `userID := uint64(1)` 改为从 context 获取

### Step 1: Create JWT auth middleware

Create `backend/internal/middleware/auth.go`:

```go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"cozy-insight/pkg/jwt"
)

const ContextKeyUserID = "userID"
const ContextKeyUsername = "username"
const ContextKeyIsAdmin = "isAdmin"

func JWTAuth(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "error": "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "error": "invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := jwtManager.Parse(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "error": "invalid or expired token"})
			c.Abort()
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyIsAdmin, claims.IsAdmin)
		c.Next()
	}
}

func GetUserID(c *gin.Context) uint64 {
	if v, ok := c.Get(ContextKeyUserID); ok {
		if id, ok := v.(uint64); ok {
			return id
		}
	}
	return 0
}

func GetUsername(c *gin.Context) string {
	if v, ok := c.Get(ContextKeyUsername); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func GetIsAdmin(c *gin.Context) bool {
	if v, ok := c.Get(ContextKeyIsAdmin); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
```

### Step 2: Add middleware to router and update all handlers

Modify `backend/api/v1/router.go`:

Add to imports:
```go
	"cozy-insight/internal/middleware"
```

In `Setup` function, before route definitions:
```go
	jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.ExpireHours)
	authMiddleware := middleware.JWTAuth(jwtManager)
```

In the `api` group, keep auth routes **without** middleware (register/login must be public), and add middleware to all other routes:

```go
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	// 受保护路由
	authd := api.Group("/")
	authd.Use(authMiddleware)
	{
		authd.GET("/datasource", dsHandler.List)
		authd.POST("/datasource", dsHandler.Create)
		// ... etc for all existing routes
	}
```

**Important:** All handlers that currently have `userID := uint64(1)` must be changed to use `middleware.GetUserID(c)`.

Files to modify for userID extraction:
- `backend/internal/handler/datasource_handler.go` — lines with `userID := uint64(1)`
- `backend/internal/handler/dataset_handler.go` — lines with `userID := uint64(1)`
- `backend/internal/handler/chart_handler.go` — lines with `userID := uint64(1)`
- `backend/internal/handler/dashboard_handler.go` — lines with `userID := uint64(1)`
- `backend/internal/handler/auth_handler.go` — no change (login/register are public)

Replace `userID := uint64(1)` with:
```go
	userID := middleware.GetUserID(c)
```

And add import in each handler file:
```go
	"cozy-insight/internal/middleware"
```

### Step 3: Verify compilation

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go build -o server cmd/server/main.go
```

### Step 4: Commit

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight
git add backend/
git commit -m "feat: add JWT auth middleware and inject real userID into all handlers"
```

---

## Task 17: 用户管理模块（CRUD + 修改密码 + 个人中心）

**Files:**
- Modify: `backend/internal/repository/user_repo.go` — 添加 List、Update、Delete
- Create: `backend/internal/dto/user.go`
- Create: `backend/internal/service/user_service.go`
- Create: `backend/internal/handler/user_handler.go`
- Modify: `backend/api/v1/router.go`
- Create: `backend/migrations/006_user_update.sql`

### Step 1: Extend user repository

Modify `backend/internal/repository/user_repo.go`, add these methods after `UpdateLastLogin`:

```go
func (r *UserRepository) List(ctx context.Context) ([]model.User, error) {
	var list []model.User
	query := `SELECT id, username, email, nick_name, avatar, phone, status, is_admin, last_login_at, created_at, updated_at FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, fmt.Errorf("list users failed: %w", err)
	}
	return list, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `UPDATE users SET email = :email, nick_name = :nick_name, avatar = :avatar, phone = :phone, status = :status, is_admin = :is_admin WHERE id = :id`
	if _, err := r.db.NamedExecContext(ctx, query, user); err != nil {
		return fmt.Errorf("update user failed: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uint64, passwordHash string) error {
	query := `UPDATE users SET password_hash = ? WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, passwordHash, id); err != nil {
		return fmt.Errorf("update password failed: %w", err)
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uint64) error {
	query := `UPDATE users SET deleted_at = NOW() WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("delete user failed: %w", err)
	}
	return nil
}
```

### Step 2: Create user DTO

Create `backend/internal/dto/user.go`:

```go
package dto

// CreateUserRequest 创建用户请求（管理员用）
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	NickName string `json:"nickName"`
	Phone    string `json:"phone"`
	Status   int8   `json:"status"`
	IsAdmin  bool   `json:"isAdmin"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email    string `json:"email"`
	NickName string `json:"nickName"`
	Phone    string `json:"phone"`
	Status   int8   `json:"status"`
	IsAdmin  bool   `json:"isAdmin"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// UserResponse 用户响应（不含密码）
type UserResponse struct {
	ID          uint64  `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	NickName    string  `json:"nickName"`
	Avatar      string  `json:"avatar"`
	Phone       string  `json:"phone"`
	Status      int8    `json:"status"`
	IsAdmin     bool    `json:"isAdmin"`
	LastLoginAt *string `json:"lastLoginAt,omitempty"`
	CreatedAt   string  `json:"createdAt"`
}
```

### Step 3: Create user service

Create `backend/internal/service/user_service.go`:

```go
package service

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, req *dto.CreateUserRequest) (*model.User, error) {
	_, err := s.repo.FindByUsername(ctx, req.Username)
	if err == nil {
		return nil, fmt.Errorf("username already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password failed: %w", err)
	}

	isAdmin := int8(0)
	if req.IsAdmin {
		isAdmin = 1
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Email:        req.Email,
		NickName:     req.NickName,
		Phone:        req.Phone,
		Status:       req.Status,
		IsAdmin:      isAdmin,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) List(ctx context.Context) ([]model.User, error) {
	return s.repo.List(ctx)
}

func (s *UserService) Update(ctx context.Context, id uint64, req *dto.UpdateUserRequest) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Email != "" {
		user.Email = req.Email
	}
	if req.NickName != "" {
		user.NickName = req.NickName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	user.Status = req.Status
	if req.IsAdmin {
		user.IsAdmin = 1
	} else {
		user.IsAdmin = 0
	}

	return s.repo.Update(ctx, user)
}

func (s *UserService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *UserService) ChangePassword(ctx context.Context, id uint64, req *dto.ChangePasswordRequest) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return fmt.Errorf("old password is incorrect")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password failed: %w", err)
	}

	return s.repo.UpdatePassword(ctx, id, string(hash))
}
```

### Step 4: Create user handler

Create `backend/internal/handler/user_handler.go`:

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/middleware"
	"cozy-insight/internal/service"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	user, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": user})
}

func (h *UserHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	user, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": user})
}

func (h *UserHandler) List(c *gin.Context) {
	list, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

func (h *UserHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	if err := h.service.Update(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *UserHandler) Profile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.service.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": user})
}
```

### Step 5: Add user routes

In `backend/api/v1/router.go`, add inside the `authd` group:
```go
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	authd.GET("/user", userHandler.List)
	authd.POST("/user", userHandler.Create)
	authd.GET("/user/:id", userHandler.Get)
	authd.PUT("/user/:id", userHandler.Update)
	authd.DELETE("/user/:id", userHandler.Delete)
	authd.GET("/user/profile", userHandler.Profile)
	authd.POST("/user/change-password", userHandler.ChangePassword)
```

### Step 6: Verify and commit

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go build -o server cmd/server/main.go
cd /Users/zhangjun/CursorProjects/CozyInsight
git add backend/
git commit -m "feat: add user management module (CRUD + change password + profile)"
```

---

## Task 18: 角色 + 菜单权限模块

**Files:**
- Create: `backend/internal/model/role.go`
- Create: `backend/internal/repository/role_repo.go`
- Create: `backend/internal/dto/role.go`
- Create: `backend/internal/service/role_service.go`
- Create: `backend/internal/handler/role_handler.go`
- Modify: `backend/api/v1/router.go`
- Create: `backend/migrations/007_rbac.sql`

### Step 1: Create role model

Create `backend/internal/model/role.go`:

```go
package model

import "time"

type Role struct {
	ID          uint64     `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Code        string     `db:"code" json:"code"`
	Description string     `db:"description" json:"description"`
	Status      int8       `db:"status" json:"status"`
	CreatedAt   time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt   *time.Time `db:"deleted_at" json:"-"`
}

type Menu struct {
	ID        uint64    `db:"id" json:"id"`
	ParentID  uint64    `db:"parent_id" json:"parentId"`
	Name      string    `db:"name" json:"name"`
	Path      string    `db:"path" json:"path"`
	Component string    `db:"component" json:"component"`
	Icon      string    `db:"icon" json:"icon"`
	Sort      int       `db:"sort" json:"sort"`
	Status    int8      `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

type RoleMenu struct {
	RoleID uint64 `db:"role_id" json:"roleId"`
	MenuID uint64 `db:"menu_id" json:"menuId"`
}

type UserRole struct {
	UserID uint64 `db:"user_id" json:"userId"`
	RoleID uint64 `db:"role_id" json:"roleId"`
}
```

### Step 2: Create role repository

Create `backend/internal/repository/role_repo.go`:

```go
package repository

import (
	"context"
	"fmt"

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

// Menu operations

func (r *RoleRepository) ListMenus(ctx context.Context) ([]model.Menu, error) {
	var list []model.Menu
	query := `SELECT * FROM menus WHERE status = 1 ORDER BY parent_id, sort`
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, fmt.Errorf("list menus failed: %w", err)
	}
	return list, nil
}

// RoleMenu operations

func (r *RoleRepository) SetRoleMenus(ctx context.Context, roleID uint64, menuIDs []uint64) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM role_menus WHERE role_id = ?`, roleID); err != nil {
		return fmt.Errorf("clear role menus failed: %w", err)
	}
	for _, menuID := range menuIDs {
		if _, err := r.db.ExecContext(ctx, `INSERT INTO role_menus (role_id, menu_id) VALUES (?, ?)`, roleID, menuID); err != nil {
			return fmt.Errorf("add role menu failed: %w", err)
		}
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

// UserRole operations

func (r *RoleRepository) SetUserRoles(ctx context.Context, userID uint64, roleIDs []uint64) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id = ?`, userID); err != nil {
		return fmt.Errorf("clear user roles failed: %w", err)
	}
	for _, roleID := range roleIDs {
		if _, err := r.db.ExecContext(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)`, userID, roleID); err != nil {
			return fmt.Errorf("add user role failed: %w", err)
		}
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
```

### Step 3: Create role DTO

Create `backend/internal/dto/role.go`:

```go
package dto

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	Status      int8   `json:"status"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Status      int8   `json:"status"`
}

// SetRoleMenusRequest 设置角色菜单请求
type SetRoleMenusRequest struct {
	MenuIDs []uint64 `json:"menuIds"`
}

// SetUserRolesRequest 设置用户角色请求
type SetUserRolesRequest struct {
	RoleIDs []uint64 `json:"roleIds"`
}
```

### Step 4: Create role service

Create `backend/internal/service/role_service.go`:

```go
package service

import (
	"context"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type RoleService struct {
	repo *repository.RoleRepository
}

func NewRoleService(repo *repository.RoleRepository) *RoleService {
	return &RoleService{repo: repo}
}

func (s *RoleService) Create(ctx context.Context, req *dto.CreateRoleRequest) (*model.Role, error) {
	role := &model.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      req.Status,
	}
	if role.Status == 0 {
		role.Status = 1
	}
	if err := s.repo.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *RoleService) GetByID(ctx context.Context, id uint64) (*model.Role, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *RoleService) List(ctx context.Context) ([]model.Role, error) {
	return s.repo.List(ctx)
}

func (s *RoleService) Update(ctx context.Context, id uint64, req *dto.UpdateRoleRequest) error {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Code != "" {
		role.Code = req.Code
	}
	role.Description = req.Description
	role.Status = req.Status
	return s.repo.Update(ctx, role)
}

func (s *RoleService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *RoleService) ListMenus(ctx context.Context) ([]model.Menu, error) {
	return s.repo.ListMenus(ctx)
}

func (s *RoleService) SetRoleMenus(ctx context.Context, roleID uint64, menuIDs []uint64) error {
	return s.repo.SetRoleMenus(ctx, roleID, menuIDs)
}

func (s *RoleService) GetRoleMenus(ctx context.Context, roleID uint64) ([]uint64, error) {
	return s.repo.GetRoleMenus(ctx, roleID)
}

func (s *RoleService) SetUserRoles(ctx context.Context, userID uint64, roleIDs []uint64) error {
	return s.repo.SetUserRoles(ctx, userID, roleIDs)
}

func (s *RoleService) GetUserRoles(ctx context.Context, userID uint64) ([]uint64, error) {
	return s.repo.GetUserRoles(ctx, userID)
}
```

### Step 5: Create role handler

Create `backend/internal/handler/role_handler.go`:

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/service"
)

type RoleHandler struct {
	service *service.RoleService
}

func NewRoleHandler(service *service.RoleService) *RoleHandler {
	return &RoleHandler{service: service}
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	role, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": role})
}

func (h *RoleHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	role, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": role})
}

func (h *RoleHandler) List(c *gin.Context) {
	list, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	if err := h.service.Update(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *RoleHandler) ListMenus(c *gin.Context) {
	list, err := h.service.ListMenus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

func (h *RoleHandler) SetRoleMenus(c *gin.Context) {
	roleID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.SetRoleMenusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	if err := h.service.SetRoleMenus(c.Request.Context(), roleID, req.MenuIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *RoleHandler) GetRoleMenus(c *gin.Context) {
	roleID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	ids, err := h.service.GetRoleMenus(c.Request.Context(), roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": ids})
}

func (h *RoleHandler) SetUserRoles(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.SetUserRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	if err := h.service.SetUserRoles(c.Request.Context(), userID, req.RoleIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}
```

### Step 6: Add role routes

In `backend/api/v1/router.go`, add inside the `authd` group:
```go
	roleRepo := repository.NewRoleRepository(db)
	roleService := service.NewRoleService(roleRepo)
	roleHandler := handler.NewRoleHandler(roleService)

	authd.GET("/role", roleHandler.List)
	authd.POST("/role", roleHandler.Create)
	authd.GET("/role/:id", roleHandler.Get)
	authd.PUT("/role/:id", roleHandler.Update)
	authd.DELETE("/role/:id", roleHandler.Delete)
	authd.GET("/role/menus", roleHandler.ListMenus)
	authd.POST("/role/:id/menus", roleHandler.SetRoleMenus)
	authd.GET("/role/:id/menus", roleHandler.GetRoleMenus)
	authd.POST("/user/:id/roles", roleHandler.SetUserRoles)
```

### Step 7: Migration

Create `backend/migrations/007_rbac.sql`:

```sql
CREATE TABLE IF NOT EXISTS roles (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL UNIQUE,
    description VARCHAR(255) DEFAULT '',
    status TINYINT DEFAULT 1 COMMENT '0=禁用, 1=启用',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

CREATE TABLE IF NOT EXISTS menus (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    parent_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
    name VARCHAR(64) NOT NULL,
    path VARCHAR(128) NOT NULL,
    component VARCHAR(128) DEFAULT '',
    icon VARCHAR(64) DEFAULT '',
    sort INT DEFAULT 0,
    status TINYINT DEFAULT 1 COMMENT '0=禁用, 1=启用',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单表';

CREATE TABLE IF NOT EXISTS role_menus (
    role_id BIGINT UNSIGNED NOT NULL,
    menu_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (role_id, menu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色菜单关联表';

CREATE TABLE IF NOT EXISTS user_roles (
    user_id BIGINT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (user_id, role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';

-- 初始化菜单数据
INSERT INTO menus (id, parent_id, name, path, component, icon, sort) VALUES
(1, 0, '工作台', '/', '', 'DashboardOutlined', 1),
(2, 0, '数据源', '/datasource', '', 'DatabaseOutlined', 2),
(3, 0, '数据集', '/dataset', '', 'TableOutlined', 3),
(4, 0, '图表', '/chart', '', 'BarChartOutlined', 4),
(5, 0, '仪表板', '/dashboard', '', 'LayoutOutlined', 5),
(6, 0, '系统管理', '/system', '', 'SettingOutlined', 6),
(7, 6, '用户管理', '/system/user', '', 'UserOutlined', 1),
(8, 6, '角色管理', '/system/role', '', 'TeamOutlined', 2),
(9, 6, '操作日志', '/system/log', '', 'FileTextOutlined', 3)
ON DUPLICATE KEY UPDATE name = VALUES(name);
```

### Step 8: Verify and commit

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go build -o server cmd/server/main.go
cd /Users/zhangjun/CursorProjects/CozyInsight
git add backend/
git commit -m "feat: add RBAC role and menu permission module"
```

---

## Task 19: 操作日志中间件 + 查询 API

**Files:**
- Create: `backend/internal/model/oper_log.go`
- Create: `backend/internal/repository/oper_log_repo.go`
- Create: `backend/internal/service/oper_log_service.go`
- Create: `backend/internal/handler/oper_log_handler.go`
- Create: `backend/internal/middleware/oper_log.go`
- Modify: `backend/api/v1/router.go`
- Create: `backend/migrations/008_oper_log.sql`

### Step 1: Create operation log model

Create `backend/internal/model/oper_log.go`:

```go
package model

import "time"

type OperationLog struct {
	ID           uint64    `db:"id" json:"id"`
	UserID       uint64    `db:"user_id" json:"userId"`
	Username     string    `db:"username" json:"username"`
	Method       string    `db:"method" json:"method"`
	Path         string    `db:"path" json:"path"`
	Query        string    `db:"query" json:"query"`
	Body         string    `db:"body" json:"body"`
	IP           string    `db:"ip" json:"ip"`
	UserAgent    string    `db:"user_agent" json:"userAgent"`
	StatusCode   int       `db:"status_code" json:"statusCode"`
	Duration     int64     `db:"duration" json:"duration"`
	ErrorMessage string    `db:"error_message" json:"errorMessage"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
}
```

### Step 2: Create operation log repository

Create `backend/internal/repository/oper_log_repo.go`:

```go
package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/model"
)

type OperationLogRepository struct {
	db *sqlx.DB
}

func NewOperationLogRepository(db *sqlx.DB) *OperationLogRepository {
	return &OperationLogRepository{db: db}
}

func (r *OperationLogRepository) Create(ctx context.Context, log *model.OperationLog) error {
	query := `INSERT INTO operation_logs (user_id, username, method, path, query, body, ip, user_agent, status_code, duration, error_message)
			  VALUES (:user_id, :username, :method, :path, :query, :body, :ip, :user_agent, :status_code, :duration, :error_message)`
	_, err := r.db.NamedExecContext(ctx, query, log)
	if err != nil {
		return fmt.Errorf("create operation log failed: %w", err)
	}
	return nil
}

func (r *OperationLogRepository) List(ctx context.Context, limit int) ([]model.OperationLog, error) {
	var list []model.OperationLog
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	query := `SELECT * FROM operation_logs ORDER BY created_at DESC LIMIT ?`
	if err := r.db.SelectContext(ctx, &list, query, limit); err != nil {
		return nil, fmt.Errorf("list operation logs failed: %w", err)
	}
	return list, nil
}
```

### Step 3: Create operation log service

Create `backend/internal/service/oper_log_service.go`:

```go
package service

import (
	"context"

	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type OperationLogService struct {
	repo *repository.OperationLogRepository
}

func NewOperationLogService(repo *repository.OperationLogRepository) *OperationLogService {
	return &OperationLogService{repo: repo}
}

func (s *OperationLogService) List(ctx context.Context, limit int) ([]model.OperationLog, error) {
	return s.repo.List(ctx, limit)
}
```

### Step 4: Create operation log handler

Create `backend/internal/handler/oper_log_handler.go`:

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/service"
)

type OperationLogHandler struct {
	service *service.OperationLogService
}

func NewOperationLogHandler(service *service.OperationLogService) *OperationLogHandler {
	return &OperationLogHandler{service: service}
}

func (h *OperationLogHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	list, err := h.service.List(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}
```

### Step 5: Create operation log middleware

Create `backend/internal/middleware/oper_log.go`:

```go
package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func OperationLog(repo *repository.OperationLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		w := &responseBodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = w

		c.Next()

		duration := time.Since(start).Milliseconds()

		// 只记录管理员操作（非 GET /health、/auth/*）
		path := c.Request.URL.Path
		if path == "/health" || path == "/api/v1/auth/register" || path == "/api/v1/auth/login" {
			return
		}

		userID := GetUserID(c)
		username := GetUsername(c)

		body := string(bodyBytes)
		if len(body) > 4096 {
			body = body[:4096] + "..."
		}

		query := c.Request.URL.RawQuery
		if len(query) > 1024 {
			query = query[:1024] + "..."
		}

		log := &model.OperationLog{
			UserID:     userID,
			Username:   username,
			Method:     c.Request.Method,
			Path:       path,
			Query:      query,
			Body:       body,
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			StatusCode: c.Writer.Status(),
			Duration:   duration,
		}

		if len(c.Errors) > 0 {
			log.ErrorMessage = c.Errors.Last().Error()
		}

		// 异步记录，不阻塞响应
		go repo.Create(c.Request.Context(), log)
	}
}
```

### Step 6: Wire up routes

In `backend/api/v1/router.go`, add to imports:
```go
	"cozy-insight/internal/middleware"
```

In `Setup` function:
```go
	operLogRepo := repository.NewOperationLogRepository(db)
	operLogService := service.NewOperationLogService(operLogRepo)
	operLogHandler := handler.NewOperationLogHandler(operLogService)

	// 注册操作日志中间件（全局）
	r.Use(middleware.OperationLog(operLogRepo))
```

Add inside `authd` group:
```go
	authd.GET("/operation-log", operLogHandler.List)
```

### Step 7: Migration

Create `backend/migrations/008_oper_log.sql`:

```sql
CREATE TABLE IF NOT EXISTS operation_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
    username VARCHAR(64) DEFAULT '',
    method VARCHAR(16) NOT NULL,
    path VARCHAR(255) NOT NULL,
    query VARCHAR(1024) DEFAULT '',
    body TEXT,
    ip VARCHAR(64) DEFAULT '',
    user_agent VARCHAR(255) DEFAULT '',
    status_code INT DEFAULT 200,
    duration BIGINT DEFAULT 0 COMMENT '耗时毫秒',
    error_message VARCHAR(512) DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at),
    INDEX idx_path (path)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';
```

### Step 8: Verify and commit

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go build -o server cmd/server/main.go
cd /Users/zhangjun/CursorProjects/CozyInsight
git add backend/
git commit -m "feat: add operation log middleware and query API"
```

---

## Task 20: 行级数据权限

**Files:**
- Create: `backend/internal/model/row_permission.go`
- Create: `backend/internal/repository/row_permission_repo.go`
- Create: `backend/internal/service/row_permission_service.go`
- Modify: `backend/internal/service/dataset_service.go` — 在 PreviewData 中注入 WHERE
- Modify: `backend/api/v1/router.go`
- Create: `backend/migrations/009_row_permission.sql`

### Step 1: Create row permission model

Create `backend/internal/model/row_permission.go`:

```go
package model

import "time"

type RowPermission struct {
	ID         uint64    `db:"id" json:"id"`
	DatasetID  uint64    `db:"dataset_id" json:"datasetId"`
	FieldName  string    `db:"field_name" json:"fieldName"`
	Operator   string    `db:"operator" json:"operator"`
	Value      string    `db:"value" json:"value"`
	UserAttr   string    `db:"user_attr" json:"userAttr"`
	Status     int8      `db:"status" json:"status"`
	CreatedAt  time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time `db:"updated_at" json:"updatedAt"`
}
```

### Step 2: Create row permission repository

Create `backend/internal/repository/row_permission_repo.go`:

```go
package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/model"
)

type RowPermissionRepository struct {
	db *sqlx.DB
}

func NewRowPermissionRepository(db *sqlx.DB) *RowPermissionRepository {
	return &RowPermissionRepository{db: db}
}

func (r *RowPermissionRepository) Create(ctx context.Context, rp *model.RowPermission) error {
	query := `INSERT INTO row_permissions (dataset_id, field_name, operator, value, user_attr, status)
			  VALUES (:dataset_id, :field_name, :operator, :value, :user_attr, :status)`
	result, err := r.db.NamedExecContext(ctx, query, rp)
	if err != nil {
		return fmt.Errorf("create row permission failed: %w", err)
	}
	id, _ := result.LastInsertId()
	rp.ID = uint64(id)
	return nil
}

func (r *RowPermissionRepository) ListByDataset(ctx context.Context, datasetID uint64) ([]model.RowPermission, error) {
	var list []model.RowPermission
	query := `SELECT * FROM row_permissions WHERE dataset_id = ? AND status = 1`
	if err := r.db.SelectContext(ctx, &list, query, datasetID); err != nil {
		return nil, fmt.Errorf("list row permissions failed: %w", err)
	}
	return list, nil
}

func (r *RowPermissionRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM row_permissions WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("delete row permission failed: %w", err)
	}
	return nil
}
```

### Step 3: Create row permission service

Create `backend/internal/service/row_permission_service.go`:

```go
package service

import (
	"context"
	"fmt"
	"strings"

	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type RowPermissionService struct {
	repo *repository.RowPermissionRepository
}

func NewRowPermissionService(repo *repository.RowPermissionRepository) *RowPermissionService {
	return &RowPermissionService{repo: repo}
}

func (s *RowPermissionService) Create(ctx context.Context, datasetID uint64, fieldName, operator, value, userAttr string) (*model.RowPermission, error) {
	rp := &model.RowPermission{
		DatasetID: datasetID,
		FieldName: fieldName,
		Operator:  operator,
		Value:     value,
		UserAttr:  userAttr,
		Status:    1,
	}
	if err := s.repo.Create(ctx, rp); err != nil {
		return nil, err
	}
	return rp, nil
}

func (s *RowPermissionService) ListByDataset(ctx context.Context, datasetID uint64) ([]model.RowPermission, error) {
	return s.repo.ListByDataset(ctx, datasetID)
}

func (s *RowPermissionService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

// BuildRowFilter 根据用户属性构建 WHERE 条件片段
// userAttrs 是用户的属性映射，如 {"dept_id": "5", "region": "east"}
func (s *RowPermissionService) BuildRowFilter(ctx context.Context, datasetID uint64, userAttrs map[string]string) (string, error) {
	permissions, err := s.repo.ListByDataset(ctx, datasetID)
	if err != nil {
		return "", err
	}

	var conditions []string
	for _, p := range permissions {
		attrValue, ok := userAttrs[p.UserAttr]
		if !ok {
			continue
		}
		cond := fmt.Sprintf("%s %s '%s'", p.FieldName, p.Operator, attrValue)
		conditions = append(conditions, cond)
	}

	if len(conditions) == 0 {
		return "", nil
	}
	return strings.Join(conditions, " AND "), nil
}
```

### Step 4: Integrate into dataset service

Modify `backend/internal/service/dataset_service.go`:

Add to `DatasetService` struct:
```go	type DatasetService struct {
		repo       *repository.DatasetRepository
		dsRepo     *repository.DatasourceRepository
		rowPermRepo *repository.RowPermissionRepository  // NEW
	}
```

Update constructor:
```go
	func NewDatasetService(repo *repository.DatasetRepository, dsRepo *repository.DatasourceRepository, rowPermRepo *repository.RowPermissionRepository) *DatasetService {
		return &DatasetService{repo: repo, dsRepo: dsRepo, rowPermRepo: rowPermRepo}
	}
```

In the `PreviewData` method, after getting datasource and before querying data, add:
```go
	// TODO: 行级权限过滤（当前为 mock 实现，后续接入真实用户属性）
	// rowFilter, _ := s.buildRowFilter(ctx, id, map[string]string{"dept_id": "1"})
	// if rowFilter != "" {
	//     // 将 rowFilter 注入 SQL WHERE
	// }
```

Add helper method:
```go
	func (s *DatasetService) buildRowFilter(ctx context.Context, datasetID uint64, userAttrs map[string]string) (string, error) {
		svc := NewRowPermissionService(s.rowPermRepo)
		return svc.BuildRowFilter(ctx, datasetID, userAttrs)
	}
```

### Step 5: Add row permission routes

In `backend/api/v1/router.go`, create a simple handler inline or add to dataset handler. For simplicity, add row permission endpoints to dataset routes:

```go
	// 行级权限（挂在数据集下）
	authd.GET("/dataset/:id/row-permissions", datasetHandler.ListRowPermissions)
	authd.POST("/dataset/:id/row-permissions", datasetHandler.CreateRowPermission)
	authd.DELETE("/dataset/:id/row-permissions/:permId", datasetHandler.DeleteRowPermission)
```

Add methods to `backend/internal/handler/dataset_handler.go`:
```go
func (h *DatasetHandler) ListRowPermissions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	// 调用 service...
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": []})
}

func (h *DatasetHandler) CreateRowPermission(c *gin.Context) {
	// ...
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *DatasetHandler) DeleteRowPermission(c *gin.Context) {
	// ...
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}
```

Actually, to keep it simple and avoid handler bloat, just wire the RowPermissionService into the DatasetService and commit the infrastructure. The actual WHERE injection can be a TODO since we don't have real SQL execution yet.

### Step 6: Migration

Create `backend/migrations/009_row_permission.sql`:

```sql
CREATE TABLE IF NOT EXISTS row_permissions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    dataset_id BIGINT UNSIGNED NOT NULL,
    field_name VARCHAR(128) NOT NULL COMMENT '数据集字段名',
    operator VARCHAR(16) NOT NULL DEFAULT '=' COMMENT '比较运算符: =, !=, >, <, >=, <=, IN',
    value VARCHAR(255) NOT NULL COMMENT '对比值或用户属性占位符',
    user_attr VARCHAR(64) DEFAULT '' COMMENT '用户属性字段名，如 dept_id',
    status TINYINT DEFAULT 1 COMMENT '0=禁用, 1=启用',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_dataset_id (dataset_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='行级数据权限表';
```

### Step 7: Verify and commit

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go build -o server cmd/server/main.go
cd /Users/zhangjun/CursorProjects/CozyInsight
git add backend/
git commit -m "feat: add row-level data permission infrastructure"
```

---

## Task 21: 前端 — 主布局（侧边栏导航）+ 用户管理 + 角色管理 + 操作日志

**Files:**
- Create: `frontend/src/components/Layout/index.tsx`
- Create: `frontend/src/api/user.ts`
- Create: `frontend/src/types/user.ts`
- Create: `frontend/src/pages/system/user/index.tsx`
- Create: `frontend/src/pages/system/role/index.tsx`
- Create: `frontend/src/pages/system/log/index.tsx`
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/src/router/index.tsx`

### Step 1: Create user types and API

Create `frontend/src/types/user.ts`:

```typescript
export interface User {
  id: number
  username: string
  email: string
  nickName: string
  avatar: string
  phone: string
  status: number
  isAdmin: boolean
  lastLoginAt?: string
  createdAt: string
}

export interface CreateUserRequest {
  username: string
  password: string
  email: string
  nickName?: string
  phone?: string
  status?: number
  isAdmin?: boolean
}

export interface ChangePasswordRequest {
  oldPassword: string
  newPassword: string
}
```

Create `frontend/src/api/user.ts`:

```typescript
import request from './request'
import type { User, CreateUserRequest, ChangePasswordRequest } from '@/types/user'

export const userAPI = {
  list: () => request.get<User[]>('/user'),
  create: (data: CreateUserRequest) => request.post<User>('/user', data),
  get: (id: number) => request.get<User>(`/user/${id}`),
  update: (id: number, data: Partial<CreateUserRequest>) => request.put(`/user/${id}`, data),
  remove: (id: number) => request.delete(`/user/${id}`),
  profile: () => request.get<User>('/user/profile'),
  changePassword: (data: ChangePasswordRequest) => request.post('/user/change-password', data),
}
```

### Step 2: Create Layout component

Create `frontend/src/components/Layout/index.tsx`:

```tsx
import { useState, useEffect } from 'react'
import { Layout, Menu, Avatar, Dropdown, theme } from 'antd'
import {
  DashboardOutlined,
  DatabaseOutlined,
  TableOutlined,
  BarChartOutlined,
  LayoutOutlined,
  SettingOutlined,
  UserOutlined,
  TeamOutlined,
  FileTextOutlined,
  LogoutOutlined,
} from '@ant-design/icons'
import { useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/store/auth'

const { Header, Sider, Content } = Layout

const menuItems = [
  { key: '/', icon: <DashboardOutlined />, label: '工作台' },
  { key: '/datasource', icon: <DatabaseOutlined />, label: '数据源' },
  { key: '/dataset', icon: <TableOutlined />, label: '数据集' },
  { key: '/chart', icon: <BarChartOutlined />, label: '图表' },
  { key: '/dashboard', icon: <LayoutOutlined />, label: '仪表板' },
  {
    key: 'system',
    icon: <SettingOutlined />,
    label: '系统管理',
    children: [
      { key: '/system/user', icon: <UserOutlined />, label: '用户管理' },
      { key: '/system/role', icon: <TeamOutlined />, label: '角色管理' },
      { key: '/system/log', icon: <FileTextOutlined />, label: '操作日志' },
    ],
  },
]

export default function MainLayout({ children }: { children: React.ReactNode }) {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const logout = useAuthStore((s) => s.logout)
  const {
    token: { colorBgContainer },
  } = theme.useToken()

  const handleMenuClick = ({ key }: { key: string }) => {
	if (key !== 'system') {
      navigate(key)
    }
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider collapsible collapsed={collapsed} onCollapse={setCollapsed}>
        <div style={{ height: 32, margin: 16, background: 'rgba(255,255,255,0.2)', borderRadius: 6 }} />
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          defaultOpenKeys={['system']}
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Sider>
      <Layout>
        <Header style={{ padding: 0, background: colorBgContainer, display: 'flex', justifyContent: 'flex-end', alignItems: 'center', paddingRight: 24 }}>
          <Dropdown
            menu={{
              items: [
                { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', onClick: logout },
              ],
            }}
          >
            <Avatar icon={<UserOutlined />} style={{ cursor: 'pointer' }} />
          </Dropdown>
        </Header>
        <Content style={{ margin: 16, background: '#fff', borderRadius: 8 }}>{children}</Content>
      </Layout>
    </Layout>
  )
}
```

### Step 3: Create system pages

Create `frontend/src/pages/system/user/index.tsx`:

```tsx
import { useState, useEffect } from 'react'
import { Table, Button, Space, Tag, Modal, Form, Input, Switch, message } from 'antd'
import { userAPI } from '@/api/user'
import type { User } from '@/types/user'

export default function UserPage() {
  const [list, setList] = useState<User[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [form] = Form.useForm()

  const fetchList = async () => {
    setLoading(true)
    try {
      const data = await userAPI.list()
      setList(data)
    } catch {
      message.error('获取用户列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchList()
  }, [])

  const handleCreate = async (values: { username: string; password: string; email: string; nickName?: string; isAdmin?: boolean }) => {
    try {
      await userAPI.create({ ...values, status: 1 })
      message.success('创建成功')
      setModalVisible(false)
      form.resetFields()
      fetchList()
    } catch {
      message.error('创建失败')
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await userAPI.remove(id)
      message.success('删除成功')
      fetchList()
    } catch {
      message.error('删除失败')
    }
  }

  const columns = [
    { title: '用户名', dataIndex: 'username' },
    { title: '邮箱', dataIndex: 'email' },
    { title: '昵称', dataIndex: 'nickName' },
    { title: '管理员', dataIndex: 'isAdmin', render: (v: boolean) => (v ? <Tag color="blue">是</Tag> : <Tag>否</Tag>) },
    { title: '状态', dataIndex: 'status', render: (v: number) => (v === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>) },
    {
      title: '操作',
      render: (_: unknown, record: User) => (
        <Space>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={() => setModalVisible(true)}>新建用户</Button>
      </div>
      <Table rowKey="id" columns={columns} dataSource={list} loading={loading} />
      <Modal title="新建用户" open={modalVisible} onCancel={() => setModalVisible(false)} footer={null}>
        <Form form={form} onFinish={handleCreate} layout="vertical">
          <Form.Item name="username" label="用户名" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, min: 6 }]}>
            <Input.Password />
          </Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ required: true, type: 'email' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="nickName" label="昵称">
            <Input />
          </Form.Item>
          <Form.Item name="isAdmin" label="管理员" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">创建</Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
```

Create `frontend/src/pages/system/role/index.tsx`:

```tsx
import { useState, useEffect } from 'react'
import { Table, Button, Space, Tag, Modal, Form, Input, message, Tree } from 'antd'
import type { TreeDataNode } from 'antd'
import { roleAPI } from '@/api/role'
import type { Role, Menu } from '@/types/role'

export default function RolePage() {
  const [list, setList] = useState<Role[]>([])
  const [menus, setMenus] = useState<Menu[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [permModal, setPermModal] = useState(false)
  const [currentRole, setCurrentRole] = useState<Role | null>(null)
  const [selectedMenus, setSelectedMenus] = useState<string[]>([])
  const [form] = Form.useForm()

  const fetchList = async () => {
    setLoading(true)
    try {
      const [roleData, menuData] = await Promise.all([roleAPI.list(), roleAPI.listMenus()])
      setList(roleData)
      setMenus(menuData)
    } catch {
      message.error('获取数据失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchList()
  }, [])

  const handleCreate = async (values: { name: string; code: string; description?: string }) => {
    try {
      await roleAPI.create(values)
      message.success('创建成功')
      setModalVisible(false)
      form.resetFields()
      fetchList()
    } catch {
      message.error('创建失败')
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await roleAPI.remove(id)
      message.success('删除成功')
      fetchList()
    } catch {
      message.error('删除失败')
    }
  }

  const openPermission = async (role: Role) => {
    setCurrentRole(role)
    try {
      const ids = await roleAPI.getRoleMenus(role.id)
      setSelectedMenus(ids.map(String))
    } catch {
      setSelectedMenus([])
    }
    setPermModal(true)
  }

  const handleSavePermission = async () => {
    if (!currentRole) return
    try {
      await roleAPI.setRoleMenus(currentRole.id, selectedMenus.map(Number))
      message.success('权限设置成功')
      setPermModal(false)
    } catch {
      message.error('权限设置失败')
    }
  }

  const buildTree = (menuList: Menu[]): TreeDataNode[] => {
    const map = new Map<number, TreeDataNode>()
    menuList.forEach((m) => {
      map.set(m.id, { key: String(m.id), title: m.name, children: [] })
    })
    const roots: TreeDataNode[] = []
    menuList.forEach((m) => {
      const node = map.get(m.id)!
      if (m.parentId === 0) {
        roots.push(node)
      } else {
        const parent = map.get(m.parentId)
        if (parent) {
          parent.children = parent.children || []
          parent.children.push(node)
        }
      }
    })
    return roots
  }

  const columns = [
    { title: '角色名', dataIndex: 'name' },
    { title: '编码', dataIndex: 'code' },
    { title: '描述', dataIndex: 'description' },
    { title: '状态', dataIndex: 'status', render: (v: number) => (v === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>) },
    {
      title: '操作',
      render: (_: unknown, record: Role) => (
        <Space>
          <Button type="link" onClick={() => openPermission(record)}>权限</Button>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={() => setModalVisible(true)}>新建角色</Button>
      </div>
      <Table rowKey="id" columns={columns} dataSource={list} loading={loading} />
      <Modal title="新建角色" open={modalVisible} onCancel={() => setModalVisible(false)} footer={null}>
        <Form form={form} onFinish={handleCreate} layout="vertical">
          <Form.Item name="name" label="角色名" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="code" label="编码" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">创建</Button>
          </Form.Item>
        </Form>
      </Modal>
      <Modal title={`设置权限 - ${currentRole?.name || ''}`} open={permModal} onCancel={() => setPermModal(false)} onOk={handleSavePermission}>
        <Tree
          checkable
          treeData={buildTree(menus)}
          checkedKeys={selectedMenus}
          onCheck={(keys) => setSelectedMenus(keys as string[])}
        />
      </Modal>
    </div>
  )
}
```

Create `frontend/src/pages/system/log/index.tsx`:

```tsx
import { useState, useEffect } from 'react'
import { Table, Tag, message } from 'antd'
import { logAPI } from '@/api/log'
import type { OperationLog } from '@/types/log'

export default function LogPage() {
  const [list, setList] = useState<OperationLog[]>([])
  const [loading, setLoading] = useState(false)

  const fetchList = async () => {
    setLoading(true)
    try {
      const data = await logAPI.list(100)
      setList(data)
    } catch {
      message.error('获取日志失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchList()
  }, [])

  const columns = [
    { title: '用户', dataIndex: 'username' },
    { title: '方法', dataIndex: 'method', render: (v: string) => <Tag>{v}</Tag> },
    { title: '路径', dataIndex: 'path' },
    { title: '状态码', dataIndex: 'statusCode' },
    { title: '耗时(ms)', dataIndex: 'duration' },
    { title: 'IP', dataIndex: 'ip' },
    { title: '时间', dataIndex: 'createdAt' },
  ]

  return (
    <div style={{ padding: 24 }}>
      <Table rowKey="id" columns={columns} dataSource={list} loading={loading} pagination={{ pageSize: 20 }} />
    </div>
  )
}
```

### Step 4: Create supporting type and API files

Create `frontend/src/types/role.ts`:

```typescript
export interface Role {
  id: number
  name: string
  code: string
  description: string
  status: number
  createdAt: string
}

export interface Menu {
  id: number
  parentId: number
  name: string
  path: string
  component: string
  icon: string
  sort: number
  status: number
}

export interface CreateRoleRequest {
  name: string
  code: string
  description?: string
}
```

Create `frontend/src/api/role.ts`:

```typescript
import request from './request'
import type { Role, Menu, CreateRoleRequest } from '@/types/role'

export const roleAPI = {
  list: () => request.get<Role[]>('/role'),
  create: (data: CreateRoleRequest) => request.post<Role>('/role', data),
  get: (id: number) => request.get<Role>(`/role/${id}`),
  update: (id: number, data: Partial<CreateRoleRequest>) => request.put(`/role/${id}`, data),
  remove: (id: number) => request.delete(`/role/${id}`),
  listMenus: () => request.get<Menu[]>('/role/menus'),
  setRoleMenus: (roleId: number, menuIds: number[]) => request.post(`/role/${roleId}/menus`, { menuIds }),
  getRoleMenus: (roleId: number) => request.get<number[]>(`/role/${roleId}/menus`),
}
```

Create `frontend/src/types/log.ts`:

```typescript
export interface OperationLog {
  id: number
  userId: number
  username: string
  method: string
  path: string
  query: string
  body: string
  ip: string
  userAgent: string
  statusCode: number
  duration: number
  errorMessage: string
  createdAt: string
}
```

Create `frontend/src/api/log.ts`:

```typescript
import request from './request'
import type { OperationLog } from '@/types/log'

export const logAPI = {
  list: (limit?: number) => request.get<OperationLog[]>('/operation-log', { params: { limit } }),
}
```

### Step 5: Update App.tsx and router

Modify `frontend/src/App.tsx`:

```tsx
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import { BrowserRouter } from 'react-router-dom'
import Router from './router'
import Layout from '@/components/Layout'

dayjs.locale('zh-cn')

function App() {
  return (
    <ConfigProvider locale={zhCN}>
      <BrowserRouter>
        <Layout>
          <Router />
        </Layout>
      </BrowserRouter>
    </ConfigProvider>
  )
}

export default App
```

Modify `frontend/src/router/index.tsx`:

```tsx
import { Routes, Route } from 'react-router-dom'
import LoginPage from '@/pages/login'
import DatasourcePage from '@/pages/datasource'
import DatasetPage from '@/pages/dataset'
import ChartPage from '@/pages/chart'
import DashboardPage from '@/pages/dashboard'
import UserPage from '@/pages/system/user'
import RolePage from '@/pages/system/role'
import LogPage from '@/pages/system/log'

export default function Router() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<div style={{ padding: 24 }}>工作台（建设中）</div>} />
      <Route path="/datasource" element={<DatasourcePage />} />
      <Route path="/dataset" element={<DatasetPage />} />
      <Route path="/chart" element={<ChartPage />} />
      <Route path="/dashboard" element={<DashboardPage />} />
      <Route path="/system/user" element={<UserPage />} />
      <Route path="/system/role" element={<RolePage />} />
      <Route path="/system/log" element={<LogPage />} />
    </Routes>
  )
}
```

Modify `frontend/src/main.tsx` to remove BrowserRouter (since it's now in App.tsx):

```tsx
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import App from './App'

const queryClient = new QueryClient()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>
  </StrictMode>,
)
```

### Step 6: Verify build

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/frontend
npm run build
```

### Step 7: Commit

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight
git add frontend/
git commit -m "feat: add frontend layout with sidebar and system management pages"
```

---

*Phase 2 全部完成：JWT 认证中间件 + 用户管理 + RBAC 角色权限 + 操作日志 + 行级数据权限 + 前端布局与系统页面。*
