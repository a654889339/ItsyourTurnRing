package service

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"itsyourturnring/config"
	"itsyourturnring/database"
	"itsyourturnring/model"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// Register 用户注册
func (s *AuthService) Register(req *model.RegisterRequest) (*model.TokenResponse, error) {
	db := database.GetDB()

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", req.Username).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	if req.Email != "" {
		err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", req.Email).Scan(&count)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New("邮箱已被注册")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	result, err := db.Exec(`
		INSERT INTO users (username, password, email, phone, role)
		VALUES (?, ?, ?, ?, 'user')`,
		req.Username, string(hashedPassword), req.Email, req.Phone)
	if err != nil {
		return nil, err
	}

	userID, _ := result.LastInsertId()

	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

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

// ==================== 微信登录 ====================

// WechatLogin 微信小程序登录
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
	reqURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appID, appSecret, code)

	resp, err := httpClient.Get(reqURL)
	if err != nil {
		return "", fmt.Errorf("请求微信API失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取微信响应失败: %w", err)
	}

	var result struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析微信响应失败: %w", err)
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("微信API错误: %d %s", result.ErrCode, result.ErrMsg)
	}
	if result.OpenID == "" {
		return "", errors.New("微信返回空openid")
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

	username := safeUsername("wx", openid)
	dummyPwd, _ := bcrypt.GenerateFromPassword([]byte(openid), bcrypt.DefaultCost)
	result, err := db.Exec(`
		INSERT INTO users (username, password, nickname, wechat_openid, role)
		VALUES (?, ?, ?, ?, 'user')`, username, string(dummyPwd), "", openid)
	if err != nil {
		return nil, fmt.Errorf("创建微信用户失败: %w", err)
	}
	userID, _ := result.LastInsertId()
	return s.GetUserByID(userID)
}

// ==================== 支付宝登录 ====================

// AlipayLogin 支付宝小程序登录
func (s *AuthService) AlipayLogin(authCode string) (*model.TokenResponse, error) {
	cfg := config.GetConfig()
	if cfg.AlipayMP.AppID == "" {
		return nil, errors.New("支付宝小程序未配置")
	}

	alipayUID, err := s.getAlipayUserID(cfg, authCode)
	if err != nil {
		return nil, fmt.Errorf("获取支付宝用户信息失败: %w", err)
	}

	user, err := s.findOrCreateByAlipay(alipayUID)
	if err != nil {
		return nil, err
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &model.TokenResponse{Token: token, User: user}, nil
}

func (s *AuthService) getAlipayUserID(cfg *config.Config, authCode string) (string, error) {
	if cfg.AlipayMP.PrivateKey == "" {
		return "", errors.New("支付宝私钥未配置，无法完成登录")
	}

	params := map[string]string{
		"app_id":     cfg.AlipayMP.AppID,
		"method":     "alipay.system.oauth.token",
		"format":     "JSON",
		"charset":    "utf-8",
		"sign_type":  "RSA2",
		"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
		"version":    "1.0",
		"grant_type": "authorization_code",
		"code":       authCode,
	}

	signStr := buildAlipaySignString(params)
	sign, err := rsaSignSHA256(signStr, cfg.AlipayMP.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("支付宝签名失败: %w", err)
	}
	params["sign"] = sign

	formValues := url.Values{}
	for k, v := range params {
		formValues.Set(k, v)
	}

	resp, err := httpClient.PostForm("https://openapi.alipay.com/gateway.do", formValues)
	if err != nil {
		return "", fmt.Errorf("请求支付宝API失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取支付宝响应失败: %w", err)
	}

	var result struct {
		Response struct {
			UserID      string `json:"user_id"`
			AccessToken string `json:"access_token"`
			Code        string `json:"code"`
			Msg         string `json:"msg"`
			SubCode     string `json:"sub_code"`
			SubMsg      string `json:"sub_msg"`
		} `json:"alipay_system_oauth_token_response"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析支付宝响应失败: %w", err)
	}

	if result.Response.Code != "" && result.Response.Code != "10000" {
		log.Printf("支付宝OAuth失败: code=%s, subCode=%s, subMsg=%s",
			result.Response.Code, result.Response.SubCode, result.Response.SubMsg)
		return "", fmt.Errorf("支付宝授权失败: %s", result.Response.SubMsg)
	}

	if result.Response.UserID == "" {
		return "", errors.New("支付宝返回空user_id")
	}

	return result.Response.UserID, nil
}

func buildAlipaySignString(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		if params[k] != "" {
			parts = append(parts, k+"="+params[k])
		}
	}
	return strings.Join(parts, "&")
}

func rsaSignSHA256(content, privateKeyStr string) (string, error) {
	privateKeyStr = strings.ReplaceAll(privateKeyStr, "-----BEGIN RSA PRIVATE KEY-----", "")
	privateKeyStr = strings.ReplaceAll(privateKeyStr, "-----END RSA PRIVATE KEY-----", "")
	privateKeyStr = strings.ReplaceAll(privateKeyStr, "-----BEGIN PRIVATE KEY-----", "")
	privateKeyStr = strings.ReplaceAll(privateKeyStr, "-----END PRIVATE KEY-----", "")
	privateKeyStr = strings.ReplaceAll(privateKeyStr, "\n", "")
	privateKeyStr = strings.ReplaceAll(privateKeyStr, "\r", "")
	privateKeyStr = strings.TrimSpace(privateKeyStr)

	keyBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return "", fmt.Errorf("base64解码私钥失败: %w", err)
	}

	var privateKey *rsa.PrivateKey

	// 尝试 PEM 解码
	if block, _ := pem.Decode([]byte("-----BEGIN PRIVATE KEY-----\n" + privateKeyStr + "\n-----END PRIVATE KEY-----")); block != nil {
		keyBytes = block.Bytes
	}

	// 先尝试 PKCS8 格式
	key, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		// 再尝试 PKCS1 格式
		privateKey, err = x509.ParsePKCS1PrivateKey(keyBytes)
		if err != nil {
			return "", fmt.Errorf("解析私钥失败(PKCS1/PKCS8均不匹配): %w", err)
		}
	} else {
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return "", errors.New("私钥类型不是RSA")
		}
	}

	h := sha256.New()
	h.Write([]byte(content))
	hashed := h.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		return "", fmt.Errorf("RSA签名失败: %w", err)
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func (s *AuthService) findOrCreateByAlipay(alipayUID string) (*model.User, error) {
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
		return &user, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	username := safeUsername("ali", alipayUID)
	dummyPwd, _ := bcrypt.GenerateFromPassword([]byte(alipayUID), bcrypt.DefaultCost)
	result, err := db.Exec(`
		INSERT INTO users (username, password, nickname, alipay_userid, role)
		VALUES (?, ?, ?, ?, 'user')`, username, string(dummyPwd), "", alipayUID)
	if err != nil {
		return nil, fmt.Errorf("创建支付宝用户失败: %w", err)
	}
	id, _ := result.LastInsertId()
	return s.GetUserByID(id)
}

// ==================== 更新用户资料 ====================

// UpdateProfile 更新用户昵称和头像
func (s *AuthService) UpdateProfile(userID int64, nickname, avatar string) (*model.User, error) {
	db := database.GetDB()

	if nickname != "" {
		if _, err := db.Exec("UPDATE users SET nickname = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			nickname, userID); err != nil {
			return nil, fmt.Errorf("更新昵称失败: %w", err)
		}
	}
	if avatar != "" {
		if _, err := db.Exec("UPDATE users SET avatar = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			avatar, userID); err != nil {
			return nil, fmt.Errorf("更新头像失败: %w", err)
		}
	}

	return s.GetUserByID(userID)
}

// ==================== 辅助函数 ====================

// safeUsername 生成安全的用户名，使用完整标识符避免碰撞，截断到50字符
func safeUsername(prefix, identifier string) string {
	username := prefix + "_" + identifier
	if len(username) > 50 {
		username = username[:50]
	}
	return username
}
