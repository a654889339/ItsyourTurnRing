package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
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
	var nickname sql.NullString
	err := db.QueryRow(`
		SELECT id, username, password, COALESCE(nickname,'') as nickname, email, phone, avatar, role, created_at, updated_at
		FROM users WHERE username = ?`, req.Username).Scan(
		&user.ID, &user.Username, &password, &nickname, &user.Email, &user.Phone,
		&user.Avatar, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	user.Nickname = nickname.String

	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password)); err != nil {
		return nil, errors.New("密码错误")
	}

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
	var nickname sql.NullString
	err := db.QueryRow(`
		SELECT id, username, COALESCE(nickname,'') as nickname, email, phone, avatar, role, created_at, updated_at
		FROM users WHERE id = ?`, id).Scan(
		&user.ID, &user.Username, &nickname, &user.Email, &user.Phone,
		&user.Avatar, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	user.Nickname = nickname.String

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

	var currentPassword string
	err := db.QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&currentPassword)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(currentPassword), []byte(oldPassword)); err != nil {
		return errors.New("原密码错误")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE users SET password = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		string(hashedPassword), userID)
	return err
}

// WechatLogin 微信小程序登录 - 用code换取openid，然后查找或创建用户
func (s *AuthService) WechatLogin(code string) (*model.TokenResponse, error) {
	cfg := config.GetConfig()
	if cfg.WechatMP.AppID == "" || cfg.WechatMP.AppSecret == "" {
		return nil, errors.New("微信小程序未配置")
	}

	openid, err := s.getWechatOpenID(cfg.WechatMP.AppID, cfg.WechatMP.AppSecret, code)
	if err != nil {
		return nil, fmt.Errorf("获取openid失败: %w", err)
	}

	user, err := s.findOrCreateByWechat(openid)
	if err != nil {
		return nil, err
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &model.TokenResponse{Token: token, User: user}, nil
}

func (s *AuthService) getWechatOpenID(appID, appSecret, code string) (string, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appID, appSecret, code)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("wechat error: %d %s", result.ErrCode, result.ErrMsg)
	}
	if result.OpenID == "" {
		return "", errors.New("empty openid")
	}
	return result.OpenID, nil
}

func (s *AuthService) findOrCreateByWechat(openid string) (*model.User, error) {
	db := database.GetDB()

	var user model.User
	var nickname sql.NullString
	err := db.QueryRow(`
		SELECT id, username, COALESCE(nickname,'') as nickname, email, phone, avatar, role, created_at, updated_at
		FROM users WHERE wechat_openid = ?`, openid).Scan(
		&user.ID, &user.Username, &nickname, &user.Email, &user.Phone,
		&user.Avatar, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err == nil {
		user.Nickname = nickname.String
		user.WechatOpenID = openid
		return &user, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	username := "wx_" + openid[:8]
	dummyPwd, _ := bcrypt.GenerateFromPassword([]byte(openid), bcrypt.DefaultCost)
	result, err := db.Exec(`
		INSERT INTO users (username, password, nickname, wechat_openid, role)
		VALUES (?, ?, ?, ?, 'user')`, username, string(dummyPwd), "", openid)
	if err != nil {
		return nil, err
	}
	userID, _ := result.LastInsertId()
	return s.GetUserByID(userID)
}

// AlipayLogin 支付宝小程序登录 - 用authCode换取user_id，然后查找或创建用户
func (s *AuthService) AlipayLogin(authCode string) (*model.TokenResponse, error) {
	cfg := config.GetConfig()
	if cfg.AlipayMP.AppID == "" {
		return nil, errors.New("支付宝小程序未配置")
	}

	userID, nickname, avatar, err := s.getAlipayUserInfo(authCode)
	if err != nil {
		return nil, fmt.Errorf("获取支付宝用户信息失败: %w", err)
	}

	user, err := s.findOrCreateByAlipay(userID, nickname, avatar)
	if err != nil {
		return nil, err
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &model.TokenResponse{Token: token, User: user}, nil
}

func (s *AuthService) getAlipayUserInfo(authCode string) (userID, nickname, avatar string, err error) {
	// 支付宝小程序通过 my.getAuthCode(scopes:'auth_user') 获取 authCode
	// 后端用 authCode 换 access_token, 再用 access_token 获取用户信息
	// 简化实现: 将 authCode 作为唯一标识（真实环境需调用支付宝 open API）
	cfg := config.GetConfig()
	_ = cfg

	// 第1步: 换取 access_token
	// 在真实环境中需要签名调用支付宝 API
	// 此处简化: 直接使用 authCode 作为 user_id 标识
	log.Printf("AlipayLogin: authCode=%s, appID=%s", authCode, cfg.AlipayMP.AppID)
	userID = "alipay_" + authCode
	if len(userID) > 32 {
		userID = userID[:32]
	}
	return userID, "", "", nil
}

func (s *AuthService) findOrCreateByAlipay(alipayUID, nickname, avatar string) (*model.User, error) {
	db := database.GetDB()

	var user model.User
	var nick sql.NullString
	err := db.QueryRow(`
		SELECT id, username, COALESCE(nickname,'') as nickname, email, phone, avatar, role, created_at, updated_at
		FROM users WHERE alipay_userid = ?`, alipayUID).Scan(
		&user.ID, &user.Username, &nick, &user.Email, &user.Phone,
		&user.Avatar, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err == nil {
		user.Nickname = nick.String
		user.AlipayUserID = alipayUID
		if nickname != "" && user.Nickname == "" {
			db.Exec("UPDATE users SET nickname = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", nickname, user.ID)
			user.Nickname = nickname
		}
		if avatar != "" && !user.Avatar.Valid {
			db.Exec("UPDATE users SET avatar = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", avatar, user.ID)
			user.Avatar = sql.NullString{String: avatar, Valid: true}
		}
		return &user, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	username := "ali_" + alipayUID[:8]
	dummyPwd, _ := bcrypt.GenerateFromPassword([]byte(alipayUID), bcrypt.DefaultCost)
	result, err := db.Exec(`
		INSERT INTO users (username, password, nickname, avatar, alipay_userid, role)
		VALUES (?, ?, ?, ?, ?, 'user')`, username, string(dummyPwd), nickname, avatar, alipayUID)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return s.GetUserByID(id)
}

// UpdateProfile 更新用户昵称和头像
func (s *AuthService) UpdateProfile(userID int64, nickname, avatar string) (*model.User, error) {
	db := database.GetDB()

	if nickname != "" {
		db.Exec("UPDATE users SET nickname = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", nickname, userID)
	}
	if avatar != "" {
		db.Exec("UPDATE users SET avatar = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", avatar, userID)
	}

	return s.GetUserByID(userID)
}
