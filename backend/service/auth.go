package service

import (
	"database/sql"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"itsyourturnring/config"
	"itsyourturnring/database"
	"itsyourturnring/model"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// Register 用户注册
func (s *AuthService) Register(req *model.RegisterRequest) (*model.TokenResponse, error) {
	db := database.GetDB()

	// 检查用户名是否已存在
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", req.Username).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if req.Email != "" {
		err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", req.Email).Scan(&count)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New("邮箱已被注册")
		}
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 插入用户
	result, err := db.Exec(`
		INSERT INTO users (username, password, email, phone, role)
		VALUES (?, ?, ?, ?, 'user')`,
		req.Username, string(hashedPassword), req.Email, req.Phone)
	if err != nil {
		return nil, err
	}

	userID, _ := result.LastInsertId()

	// 获取用户信息
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// 生成Token
	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &model.TokenResponse{
		Token: token,
		User:  user,
	}, nil
}

// Login 用户登录
func (s *AuthService) Login(req *model.LoginRequest) (*model.TokenResponse, error) {
	db := database.GetDB()

	var user model.User
	var password string
	err := db.QueryRow(`
		SELECT id, username, password, email, phone, avatar, role, created_at, updated_at
		FROM users WHERE username = ?`, req.Username).Scan(
		&user.ID, &user.Username, &password, &user.Email, &user.Phone,
		&user.Avatar, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password)); err != nil {
		return nil, errors.New("密码错误")
	}

	// 生成Token
	token, err := s.GenerateToken(&user)
	if err != nil {
		return nil, err
	}

	return &model.TokenResponse{
		Token: token,
		User:  &user,
	}, nil
}

// GetUserByID 根据ID获取用户
func (s *AuthService) GetUserByID(id int64) (*model.User, error) {
	db := database.GetDB()

	var user model.User
	err := db.QueryRow(`
		SELECT id, username, email, phone, avatar, role, created_at, updated_at
		FROM users WHERE id = ?`, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Phone,
		&user.Avatar, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GenerateToken 生成JWT Token
func (s *AuthService) GenerateToken(user *model.User) (string, error) {
	cfg := config.GetConfig()

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * time.Duration(cfg.JWT.ExpireHours)).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// ValidateToken 验证Token
func (s *AuthService) ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	cfg := config.GetConfig()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, errors.New("invalid token")
}

// UpdateUser 更新用户信息
func (s *AuthService) UpdateUser(userID int64, email, phone, avatar string) error {
	db := database.GetDB()

	_, err := db.Exec(`
		UPDATE users SET email = ?, phone = ?, avatar = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, email, phone, avatar, userID)
	return err
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	db := database.GetDB()

	// 获取当前密码
	var currentPassword string
	err := db.QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&currentPassword)
	if err != nil {
		return err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(currentPassword), []byte(oldPassword)); err != nil {
		return errors.New("原密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE users SET password = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		string(hashedPassword), userID)
	return err
}
