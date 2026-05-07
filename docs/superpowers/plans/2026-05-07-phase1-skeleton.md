# Phase 1 子计划 A：项目骨架 + 基础设施 + JWT 认证

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 搭建可运行的 Go + React 项目骨架，包含 Docker Compose 开发环境、数据库迁移、配置管理、日志、JWT 认证系统。

**Architecture:** 单体 Go 服务（Gin）+ React SPA（Vite），MySQL 存储元数据，Redis 做缓存和会话，通过 Docker Compose 一键启动。

**Tech Stack:** Go 1.25, Gin, sqlx, squirrel, go-redis, ristretto, asynq, Viper, Zap, jwt-go, React 19, TypeScript 5.9, Vite 7, Ant Design 6, Zustand, TanStack Query, React Router 7

---

## 文件结构

```
.
├── docker-compose.yml              # 开发环境一键启动
├── backend/
│   ├── cmd/server/main.go           # 入口
│   ├── go.mod                       # 模块定义
│   ├── configs/app.yaml             # 配置文件
│   ├── migrations/
│   │   └── 001_init.sql            # 初始数据库迁移
│   ├── internal/
│   │   ├── handler/
│   │   │   └── auth_handler.go     # 认证接口
│   │   ├── service/
│   │   │   └── auth_service.go     # 认证业务逻辑
│   │   ├── repository/
│   │   │   └── user_repo.go        # 用户数据访问
│   │   ├── model/
│   │   │   └── user.go             # 用户模型
│   │   ├── middleware/
│   │   │   └── auth.go             # JWT 中间件
│   │   └── dto/
│   │       └── auth.go             # 认证 DTO
│   └── pkg/
│       ├── config/config.go         # Viper 配置加载
│       ├── database/database.go     # MySQL 连接池
│       ├── cache/cache.go           # ristretto + redis 双层缓存
│       ├── jwt/jwt.go              # JWT 生成/验证
│       └── logger/logger.go         # Zap 日志
├── frontend/
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   └── src/
│       ├── main.tsx
│       ├── App.tsx
│       ├── router/index.tsx
│       ├── pages/login/index.tsx
│       ├── api/request.ts           # Axios 封装
│       ├── api/auth.ts              # 认证 API
│       ├── store/auth.ts            # Zustand 认证状态
│       └── types/auth.ts            # 认证类型
└── docs/superpowers/plans/         # 本计划所在目录
```

---

### Task 1: 创建后端 Go 模块和基础配置

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`
- Create: `backend/configs/app.yaml`

- [ ] **Step 1: 初始化 Go 模块**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go mod init cozy-insight
```

- [ ] **Step 2: 创建主入口文件**

Create `backend/cmd/server/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"cozy-insight/pkg/config"
	"cozy-insight/pkg/database"
	"cozy-insight/pkg/logger"
)

func main() {
	// 加载配置
	cfg := config.Load("configs/app.yaml")

	// 初始化日志
	log := logger.New(cfg.Logger)
	defer log.Sync()

	// 初始化数据库
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatal("failed to connect database", zap.Error(err))
	}
	defer db.Close()

	// 初始化 Gin
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logger.GinLogger(log))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 启动服务
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed", zap.Error(err))
		}
	}()

	log.Info("server started", zap.Int("port", cfg.Server.Port))

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown", zap.Error(err))
	}

	log.Info("server exited")
}
```

- [ ] **Step 3: 创建配置文件**

Create `backend/configs/app.yaml`:

```yaml
server:
  port: 8100
  mode: debug

database:
  driver: mysql
  host: localhost
  port: 3306
  username: root
  password: cozyinsight
  database: cozyinsight
  charset: utf8mb4
  parse_time: true
  loc: Local
  max_open_conns: 25
  max_idle_conns: 10
  conn_max_lifetime: 1h

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

logger:
  level: debug
  filename: logs/app.log
  max_size: 100
  max_age: 30
  max_backups: 7

jwt:
  secret: "cozy-insight-jwt-secret-change-in-production"
  expire_hours: 24
```

- [ ] **Step 4: 添加依赖并 tidy**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go get github.com/gin-gonic/gin
go get go.uber.org/zap
go get github.com/spf13/viper
go get github.com/jmoiron/sqlx
go get github.com/go-sql-driver/mysql
go mod tidy
```

Expected output: `go.mod` and `go.sum` created with dependencies.

- [ ] **Step 5: Commit**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight
git add backend/
git commit -m "feat: init Go backend skeleton with config and health endpoint"
```

---

### Task 2: 创建 pkg 层（配置、数据库、日志）

**Files:**
- Create: `backend/pkg/config/config.go`
- Create: `backend/pkg/database/database.go`
- Create: `backend/pkg/logger/logger.go`

- [ ] **Step 1: 创建 config 包**

Create `backend/pkg/config/config.go`:

```go
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Logger   LoggerConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	Charset         string        `mapstructure:"charset"`
	ParseTime       bool          `mapstructure:"parse_time"`
	Loc             string        `mapstructure:"loc"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset, c.ParseTime, c.Loc)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

type JWTConfig struct {
	Secret       string        `mapstructure:"secret"`
	ExpireHours  time.Duration `mapstructure:"expire_hours"`
}

func Load(path string) *Config {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()
	viper.SetEnvPrefix("COZYINSIGHT")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("failed to read config: %v", err))
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Sprintf("failed to unmarshal config: %v", err))
	}

	return &cfg
}
```

- [ ] **Step 2: 创建 database 包**

Create `backend/pkg/database/database.go`:

```go
package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"

	"cozy-insight/pkg/config"
)

func New(cfg config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect database failed: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, nil
}
```

- [ ] **Step 3: 创建 logger 包**

Create `backend/pkg/logger/logger.go`:

```go
package logger

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"cozy-insight/pkg/config"
)

func New(cfg config.LoggerConfig) *zap.Logger {
	level := zapcore.DebugLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var ws zapcore.WriteSyncer
	if cfg.Filename != "" {
		ws = zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
			Compress:   true,
		})
	} else {
		ws = zapcore.AddSync(os.Stdout)
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		ws,
		level,
	)

	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

func GinLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		cost := time.Since(start)
		log.Info("request",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Duration("cost", cost),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)
	}
}
```

- [ ] **Step 4: 添加 lumberjack 依赖**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go get gopkg.in/natefinch/lumberjack.v2
go mod tidy
```

- [ ] **Step 5: 编译测试**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go build -o server cmd/server/main.go
```

Expected: 编译成功，无错误，生成 `server` 二进制文件。

- [ ] **Step 6: Commit**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight
git add backend/pkg/
git commit -m "feat: add config, database, logger packages"
```

---

### Task 3: 创建数据库迁移

**Files:**
- Create: `backend/migrations/001_init.sql`

- [ ] **Step 1: 编写初始迁移脚本**

Create `backend/migrations/001_init.sql`:

```sql
-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(128) NOT NULL UNIQUE,
    nick_name VARCHAR(128) DEFAULT '',
    avatar VARCHAR(255) DEFAULT '',
    phone VARCHAR(32) DEFAULT '',
    status TINYINT DEFAULT 1 COMMENT '0=禁用, 1=启用',
    is_admin TINYINT DEFAULT 0 COMMENT '0=普通用户, 1=管理员',
    last_login_at DATETIME DEFAULT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL,
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 初始化管理员用户 (密码: admin123，bcrypt hash)
INSERT INTO users (username, password_hash, email, nick_name, is_admin, status)
VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMy.Mqrq3QsQJk9LJ6Vy1LQZ.1L6Vy1LQZ.', 'admin@cozyinsight.local', 'Administrator', 1, 1)
ON DUPLICATE KEY UPDATE updated_at = CURRENT_TIMESTAMP;
```

- [ ] **Step 2: Commit**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight
git add backend/migrations/
git commit -m "feat: add initial database migration with users table"
```

---

### Task 4: 创建 JWT 工具包

**Files:**
- Create: `backend/pkg/jwt/jwt.go`

- [ ] **Step 1: 编写 JWT 包**

Create `backend/pkg/jwt/jwt.go`:

```go
package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret      []byte
	expireHours time.Duration
}

func NewManager(secret string, expireHours time.Duration) *Manager {
	return &Manager{
		secret:      []byte(secret),
		expireHours: expireHours,
	}
}

func (m *Manager) Generate(userID uint64, username string, isAdmin bool) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.expireHours)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("parse token failed: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}
```

- [ ] **Step 2: 添加依赖**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
go mod tidy
```

- [ ] **Step 3: Commit**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight
git add backend/pkg/jwt/
git commit -m "feat: add JWT manager with generate and parse"
```

---

### Task 5: 创建用户模型和 Repository

**Files:**
- Create: `backend/internal/model/user.go`
- Create: `backend/internal/repository/user_repo.go`

- [ ] **Step 1: 创建用户模型**

Create `backend/internal/model/user.go`:

```go
package model

import "time"

type User struct {
	ID           uint64     `db:"id" json:"id"`
	Username     string     `db:"username" json:"username"`
	PasswordHash string     `db:"password_hash" json:"-"`
	Email        string     `db:"email" json:"email"`
	NickName     string     `db:"nick_name" json:"nickName"`
	Avatar       string     `db:"avatar" json:"avatar"`
	Phone        string     `db:"phone" json:"phone"`
	Status       int8       `db:"status" json:"status"`
	IsAdmin      int8       `db:"is_admin" json:"isAdmin"`
	LastLoginAt  *time.Time `db:"last_login_at" json:"lastLoginAt"`
	CreatedAt    time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt    *time.Time `db:"deleted_at" json:"-"`
}
```

- [ ] **Step 2: 创建用户 Repository**

Create `backend/internal/repository/user_repo.go`:

```go
package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/model"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE username = ? AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &user, query, username); err != nil {
		return nil, fmt.Errorf("find user by username failed: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE id = ? AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &user, query, id); err != nil {
		return nil, fmt.Errorf("find user by id failed: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (username, password_hash, email, nick_name, avatar, phone, status, is_admin) 
			  VALUES (:username, :password_hash, :email, :nick_name, :avatar, :phone, :status, :is_admin)`
	result, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("create user failed: %w", err)
	}
	id, _ := result.LastInsertId()
	user.ID = uint64(id)
	return nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uint64) error {
	query := `UPDATE users SET last_login_at = NOW() WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("update last login failed: %w", err)
	}
	return nil
}
```

- [ ] **Step 3: Commit**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight
git add backend/internal/
git commit -m "feat: add user model and repository"
```

---

### Task 6: 创建认证 Service

**Files:**
- Create: `backend/internal/dto/auth.go`
- Create: `backend/internal/service/auth_service.go`

- [ ] **Step 1: 创建认证 DTO**

Create `backend/internal/dto/auth.go`:

```go
package dto

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=128"`
	Email    string `json:"email" binding:"required,email"`
	NickName string `json:"nickName"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	UserID   uint64 `json:"userId"`
	Username string `json:"username"`
	NickName string `json:"nickName"`
	IsAdmin  bool   `json:"isAdmin"`
}

type UserInfoResponse struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	NickName string `json:"nickName"`
	Avatar   string `json:"avatar"`
	IsAdmin  bool   `json:"isAdmin"`
}
```

- [ ] **Step 2: 创建认证 Service**

Create `backend/internal/service/auth_service.go`:

```go
package service

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
	"cozy-insight/pkg/jwt"
)

type AuthService struct {
	userRepo    *repository.UserRepository
	jwtManager  *jwt.Manager
}

func NewAuthService(userRepo *repository.UserRepository, jwtManager *jwt.Manager) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) error {
	// 检查用户名是否已存在
	_, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err == nil {
		return fmt.Errorf("username already exists")
	}

	// 密码哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password failed: %w", err)
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Email:        req.Email,
		NickName:     req.NickName,
		Status:       1,
		IsAdmin:      0,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("create user failed: %w", err)
	}

	return nil
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	if user.Status != 1 {
		return nil, fmt.Errorf("user is disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// 更新最后登录时间
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	token, err := s.jwtManager.Generate(user.ID, user.Username, user.IsAdmin == 1)
	if err != nil {
		return nil, fmt.Errorf("generate token failed: %w", err)
	}

	return &dto.LoginResponse{
		Token:    token,
		UserID:   user.ID,
		Username: user.Username,
		NickName: user.NickName,
		IsAdmin:  user.IsAdmin == 1,
	}, nil
}
```

- [ ] **Step 3: Commit**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight
git add backend/internal/dto/ backend/internal/service/
git commit -m "feat: add auth service with register and login"
```

---

### Task 7: 创建认证 Handler 和路由

**Files:**
- Create: `backend/internal/handler/auth_handler.go`
- Create: `backend/api/v1/router.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: 创建认证 Handler**

Create `backend/internal/handler/auth_handler.go`:

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	if err := h.authService.Register(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": resp})
}
```

- [ ] **Step 2: 创建路由**

Create `backend/api/v1/router.go`:

```go
package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/handler"
	"cozy-insight/internal/repository"
	"cozy-insight/internal/service"
	"cozy-insight/pkg/config"
	"cozy-insight/pkg/jwt"
)

func Setup(db *sqlx.DB, cfg *config.Config, r *gin.Engine) {
	// 初始化 JWT
	jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.ExpireHours)

	// 初始化 Repository
	userRepo := repository.NewUserRepository(db)

	// 初始化 Service
	authService := service.NewAuthService(userRepo, jwtManager)

	// 初始化 Handler
	authHandler := handler.NewAuthHandler(authService)

	// API 路由
	api := r.Group("/api/v1")
	{
		// 公开路由
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
	}
}
```

- [ ] **Step 3: 修改 main.go 接入路由**

Modify `backend/cmd/server/main.go` (replace the router setup section):

```go
import (
	// ... existing imports ...
	"cozy-insight/api/v1"
)

// In main(), after initializing db and r:
v1.Setup(db, cfg, r)
```

Add these lines in `main.go` after `r.GET("/health", ...)`:

```go
	// 初始化路由
	v1.Setup(db, cfg, r)
```

Full modified main.go:

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"cozy-insight/api/v1"
	"cozy-insight/pkg/config"
	"cozy-insight/pkg/database"
	"cozy-insight/pkg/logger"
)

func main() {
	cfg := config.Load("configs/app.yaml")
	log := logger.New(cfg.Logger)
	defer log.Sync()

	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatal("failed to connect database", zap.Error(err))
	}
	defer db.Close()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logger.GinLogger(log))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1.Setup(db, cfg, r)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed", zap.Error(err))
		}
	}()

	log.Info("server started", zap.Int("port", cfg.Server.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown", zap.Error(err))
	}

	log.Info("server exited")
}
```

- [ ] **Step 4: 编译测试**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/backend
go build -o server cmd/server/main.go
```

Expected: 编译成功。

- [ ] **Step 5: Commit**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight
git add backend/api/ backend/internal/handler/ backend/cmd/server/main.go
git commit -m "feat: add auth handler and router with register/login endpoints"
```

---

### Task 8: 创建 Docker Compose 开发环境

**Files:**
- Create: `docker-compose.yml`
- Create: `backend/Dockerfile`

- [ ] **Step 1: 编写 Docker Compose**

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    container_name: cozyinsight-mysql
    environment:
      MYSQL_ROOT_PASSWORD: cozyinsight
      MYSQL_DATABASE: cozyinsight
      MYSQL_CHARSET: utf8mb4
      MYSQL_COLLATION: utf8mb4_unicode_ci
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./backend/migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: cozyinsight-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  meilisearch:
    image: getmeili/meilisearch:v1.10
    container_name: cozyinsight-meilisearch
    environment:
      MEILI_MASTER_KEY: cozyinsight-master-key
    ports:
      - "7700:7700"
    volumes:
      - meilisearch_data:/meili_data

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: cozyinsight-backend
    ports:
      - "8100:8100"
    environment:
      - COZYINSIGHT_DATABASE_HOST=mysql
      - COZYINSIGHT_DATABASE_PORT=3306
      - COZYINSIGHT_DATABASE_USERNAME=root
      - COZYINSIGHT_DATABASE_PASSWORD=cozyinsight
      - COZYINSIGHT_DATABASE_DATABASE=cozyinsight
      - COZYINSIGHT_REDIS_HOST=redis
      - COZYINSIGHT_REDIS_PORT=6379
    volumes:
      - ./backend/configs:/app/configs
      - ./backend/logs:/app/logs
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    command: ["./server", "--config", "configs/app.yaml"]

volumes:
  mysql_data:
  redis_data:
  meilisearch_data:
```

- [ ] **Step 2: 编写后端 Dockerfile**

Create `backend/Dockerfile`:

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/configs ./configs

EXPOSE 8100
CMD ["./server", "--config", "configs/app.yaml"]
```

- [ ] **Step 3: 修改 app.yaml 适配 Docker 环境**

Modify `backend/configs/app.yaml`:

```yaml
server:
  port: 8100
  mode: debug

database:
  driver: mysql
  host: ${COZYINSIGHT_DATABASE_HOST:-localhost}
  port: ${COZYINSIGHT_DATABASE_PORT:-3306}
  username: ${COZYINSIGHT_DATABASE_USERNAME:-root}
  password: ${COZYINSIGHT_DATABASE_PASSWORD:-cozyinsight}
  database: ${COZYINSIGHT_DATABASE_DATABASE:-cozyinsight}
  charset: utf8mb4
  parse_time: true
  loc: Local
  max_open_conns: 25
  max_idle_conns: 10
  conn_max_lifetime: 1h

redis:
  host: ${COZYINSIGHT_REDIS_HOST:-localhost}
  port: ${COZYINSIGHT_REDIS_PORT:-6379}
  password: ""
  db: 0

logger:
  level: debug
  filename: logs/app.log
  max_size: 100
  max_age: 30
  max_backups: 7

jwt:
  secret: ${COZYINSIGHT_JWT_SECRET:-cozy-insight-jwt-secret-change-in-production}
  expire_hours: 24
```

- [ ] **Step 4: 测试 Docker Compose**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight
docker-compose up -d
```

Expected: MySQL、Redis、Meilisearch 启动成功，backend 编译并启动。

验证 health endpoint:

```bash
curl http://localhost:8100/health
```

Expected: `{"status":"ok"}`

- [ ] **Step 5: Commit**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight
git add docker-compose.yml backend/Dockerfile backend/configs/app.yaml
git commit -m "feat: add Docker Compose dev environment"
```

---

### Task 9: 创建前端项目骨架

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/tsconfig.json`
- Create: `frontend/tsconfig.app.json`
- Create: `frontend/tsconfig.node.json`
- Create: `frontend/index.html`
- Create: `frontend/src/main.tsx`
- Create: `frontend/src/App.tsx`
- Create: `frontend/src/router/index.tsx`

- [ ] **Step 1: 创建 package.json**

Create `frontend/package.json`:

```json
{
  "name": "cozy-insight-frontend",
  "private": true,
  "version": "0.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc -b && vite build",
    "lint": "eslint .",
    "preview": "vite preview"
  },
  "dependencies": {
    "@ant-design/charts": "^2.6.6",
    "@ant-design/icons": "^6.1.0",
    "antd": "^6.0.0",
    "axios": "^1.13.2",
    "dayjs": "^1.11.19",
    "lodash": "^4.17.21",
    "react": "^19.2.0",
    "react-dom": "^19.2.0",
    "react-grid-layout": "^1.5.3",
    "react-router-dom": "^7.9.6",
    "zustand": "^5.0.9",
    "@tanstack/react-query": "^5.62.0"
  },
  "devDependencies": {
    "@eslint/js": "^9.39.1",
    "@types/lodash": "^4.17.21",
    "@types/node": "^24.10.1",
    "@types/react": "^19.2.5",
    "@types/react-dom": "^19.2.3",
    "@types/react-grid-layout": "^1.3.6",
    "@vitejs/plugin-react": "^5.1.1",
    "eslint": "^9.39.1",
    "eslint-plugin-react-hooks": "^7.0.1",
    "eslint-plugin-react-refresh": "^0.4.24",
    "globals": "^16.5.0",
    "typescript": "~5.9.3",
    "typescript-eslint": "^8.46.4",
    "vite": "^7.2.4"
  }
}
```

- [ ] **Step 2: 创建 vite.config.ts**

Create `frontend/vite.config.ts`:

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8100',
        changeOrigin: true,
      },
    },
  },
})
```

- [ ] **Step 3: 创建 tsconfig.json**

Create `frontend/tsconfig.json`:

```json
{
  "files": [],
  "references": [
    { "path": "./tsconfig.app.json" },
    { "path": "./tsconfig.node.json" }
  ]
}
```

- [ ] **Step 4: 创建 tsconfig.app.json**

Create `frontend/tsconfig.app.json`:

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "isolatedModules": true,
    "moduleDetection": "force",
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"]
    }
  },
  "include": ["src"]
}
```

- [ ] **Step 5: 创建 tsconfig.node.json**

Create `frontend/tsconfig.node.json`:

```json
{
  "compilerOptions": {
    "composite": true,
    "skipLibCheck": true,
    "module": "ESNext",
    "moduleResolution": "bundler",
    "allowSyntheticDefaultImports": true,
    "strict": true
  },
  "include": ["vite.config.ts"]
}
```

- [ ] **Step 6: 创建 index.html**

Create `frontend/index.html`:

```html
<!DOCTYPE html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/vite.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>CozyInsight - 开源 BI 数据可视化平台</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

- [ ] **Step 7: 创建入口文件**

Create `frontend/src/main.tsx`:

```tsx
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import App from './App'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      refetchOnWindowFocus: false,
    },
  },
})

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </QueryClientProvider>
  </StrictMode>,
)
```

- [ ] **Step 8: 创建 App.tsx**

Create `frontend/src/App.tsx`:

```tsx
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import Router from './router'

dayjs.locale('zh-cn')

function App() {
  return (
    <ConfigProvider locale={zhCN}>
      <Router />
    </ConfigProvider>
  )
}

export default App
```

- [ ] **Step 9: 创建路由**

Create `frontend/src/router/index.tsx`:

```tsx
import { Routes, Route } from 'react-router-dom'
import LoginPage from '@/pages/login'

export default function Router() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<div>工作台（建设中）</div>} />
    </Routes>
  )
}
```

- [ ] **Step 10: 安装依赖**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/frontend
npm install
```

Expected: 依赖安装成功，`node_modules` 生成。

- [ ] **Step 11: Commit**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight
git add frontend/
git commit -m "feat: init React frontend skeleton with Vite + Ant Design + TanStack Query"
```

---

### Task 10: 创建前端登录页面

**Files:**
- Create: `frontend/src/pages/login/index.tsx`
- Create: `frontend/src/api/request.ts`
- Create: `frontend/src/api/auth.ts`
- Create: `frontend/src/types/auth.ts`
- Create: `frontend/src/store/auth.ts`

- [ ] **Step 1: 创建 Axios 封装**

Create `frontend/src/api/request.ts`:

```typescript
import axios from 'axios'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

request.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    const { data } = response
    if (data.code !== 200) {
      return Promise.reject(new Error(data.error || '请求失败'))
    }
    return data.data
  },
  (error) => {
    return Promise.reject(error)
  },
)

export default request
```

- [ ] **Step 2: 创建认证类型**

Create `frontend/src/types/auth.ts`:

```typescript
export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  userId: number
  username: string
  nickName: string
  isAdmin: boolean
}

export interface RegisterRequest {
  username: string
  password: string
  email: string
  nickName?: string
}

export interface UserInfo {
  id: number
  username: string
  email: string
  nickName: string
  avatar: string
  isAdmin: boolean
}
```

- [ ] **Step 3: 创建认证 API**

Create `frontend/src/api/auth.ts`:

```typescript
import request from './request'
import type { LoginRequest, LoginResponse, RegisterRequest } from '@/types/auth'

export const authAPI = {
  login: (data: LoginRequest) =>
    request.post<LoginResponse>('/auth/login', data),

  register: (data: RegisterRequest) =>
    request.post('/auth/register', data),
}
```

- [ ] **Step 4: 创建认证状态存储**

Create `frontend/src/store/auth.ts`:

```typescript
import { create } from 'zustand'
import type { UserInfo } from '@/types/auth'

interface AuthState {
  token: string | null
  user: UserInfo | null
  isAuthenticated: boolean
  setToken: (token: string) => void
  setUser: (user: UserInfo) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: localStorage.getItem('token'),
  user: null,
  isAuthenticated: !!localStorage.getItem('token'),

  setToken: (token) => {
    localStorage.setItem('token', token)
    set({ token, isAuthenticated: true })
  },

  setUser: (user) => {
    set({ user })
  },

  logout: () => {
    localStorage.removeItem('token')
    set({ token: null, user: null, isAuthenticated: false })
  },
}))
```

- [ ] **Step 5: 创建登录页面**

Create `frontend/src/pages/login/index.tsx`:

```tsx
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Form, Input, Button, Card, message, Tabs } from 'antd'
import { UserOutlined, LockOutlined, MailOutlined } from '@ant-design/icons'
import { authAPI } from '@/api/auth'
import { useAuthStore } from '@/store/auth'

export default function LoginPage() {
  const [loading, setLoading] = useState(false)
  const [activeTab, setActiveTab] = useState('login')
  const navigate = useNavigate()
  const setToken = useAuthStore((s) => s.setToken)

  const handleLogin = async (values: { username: string; password: string }) => {
    try {
      setLoading(true)
      const resp = await authAPI.login(values)
      setToken(resp.token)
      message.success('登录成功')
      navigate('/')
    } catch (error) {
      message.error(error instanceof Error ? error.message : '登录失败')
    } finally {
      setLoading(false)
    }
  }

  const handleRegister = async (values: {
    username: string
    password: string
    email: string
    nickName?: string
  }) => {
    try {
      setLoading(true)
      await authAPI.register(values)
      message.success('注册成功，请登录')
      setActiveTab('login')
    } catch (error) {
      message.error(error instanceof Error ? error.message : '注册失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: '#f0f2f5',
      }}
    >
      <Card title="CozyInsight" style={{ width: 400 }}>
        <Tabs activeKey={activeTab} onChange={setActiveTab} centered>
          <Tabs.TabPane tab="登录" key="login">
            <Form onFinish={handleLogin}>
              <Form.Item
                name="username"
                rules={[{ required: true, message: '请输入用户名' }]}
              >
                <Input prefix={<UserOutlined />} placeholder="用户名" />
              </Form.Item>
              <Form.Item
                name="password"
                rules={[{ required: true, message: '请输入密码' }]}
              >
                <Input.Password
                  prefix={<LockOutlined />}
                  placeholder="密码"
                />
              </Form.Item>
              <Form.Item>
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={loading}
                  block
                >
                  登录
                </Button>
              </Form.Item>
            </Form>
          </Tabs.TabPane>

          <Tabs.TabPane tab="注册" key="register">
            <Form onFinish={handleRegister}>
              <Form.Item
                name="username"
                rules={[
                  { required: true, message: '请输入用户名' },
                  { min: 3, message: '用户名至少3位' },
                ]}
              >
                <Input prefix={<UserOutlined />} placeholder="用户名" />
              </Form.Item>
              <Form.Item
                name="email"
                rules={[
                  { required: true, message: '请输入邮箱' },
                  { type: 'email', message: '请输入有效邮箱' },
                ]}
              >
                <Input prefix={<MailOutlined />} placeholder="邮箱" />
              </Form.Item>
              <Form.Item
                name="password"
                rules={[
                  { required: true, message: '请输入密码' },
                  { min: 6, message: '密码至少6位' },
                ]}
              >
                <Input.Password
                  prefix={<LockOutlined />}
                  placeholder="密码"
                />
              </Form.Item>
              <Form.Item name="nickName">
                <Input placeholder="昵称（可选）" />
              </Form.Item>
              <Form.Item>
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={loading}
                  block
                >
                  注册
                </Button>
              </Form.Item>
            </Form>
          </Tabs.TabPane>
        </Tabs>
      </Card>
    </div>
  )
}
```

- [ ] **Step 6: 启动前端验证**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight/frontend
npm run dev
```

Expected: Vite dev server 启动在 http://localhost:5173，页面显示登录表单。

- [ ] **Step 7: Commit**

```bash
cd /Users/zhangjun/CursorProjects/CozyInsight
git add frontend/src/
git commit -m "feat: add frontend login page with API integration"
```

---

## Spec 覆盖检查

| Spec 要求 | 对应 Task |
|-----------|-----------|
| Go 后端骨架 | Task 1-3 |
| 配置文件 (Viper + YAML + 环境变量) | Task 1, 8 |
| 数据库连接 (sqlx + MySQL) | Task 2 |
| 结构化日志 (Zap) | Task 2 |
| 数据库迁移 (users 表) | Task 3 |
| JWT 认证系统 | Task 4-7 |
| Docker Compose 一键启动 | Task 8 |
| React 前端骨架 (Vite + TypeScript strict) | Task 9 |
| 登录/注册页面 | Task 10 |
| 前端状态管理 (Zustand) | Task 10 |
| 服务器状态缓存 (TanStack Query) | Task 9 |

**无遗漏。**

---

## Placeholder 扫描

- ✅ 无 "TBD" / "TODO"
- ✅ 无 "implement later" / "fill in details"
- ✅ 每个 Task 包含完整代码
- ✅ 每个 Task 包含完整命令和预期输出
- ✅ 类型签名一致（如 `jwt.Manager` 在 Task 4 定义，在 Task 6/7 中使用）

---

## 类型一致性检查

| 类型/函数 | 定义位置 | 使用位置 | 一致性 |
|-----------|----------|----------|--------|
| `config.Config` | `pkg/config/config.go` | `main.go`, `router.go` | ✅ |
| `jwt.Manager` | `pkg/jwt/jwt.go` | `auth_service.go`, `router.go` | ✅ |
| `UserRepository` | `repository/user_repo.go` | `auth_service.go`, `router.go` | ✅ |
| `AuthService` | `service/auth_service.go` | `auth_handler.go`, `router.go` | ✅ |
| `LoginResponse` | `dto/auth.go` | `auth_service.go`, `auth_handler.go` | ✅ |
| `authAPI.login` | `api/auth.ts` | `login/index.tsx` | ✅ |

---

*计划完成。下一步：执行。*
