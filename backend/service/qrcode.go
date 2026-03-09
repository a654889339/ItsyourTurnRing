package service

import (
	"fmt"
	"net/url"
	"strings"

	"itsyourturnring/config"
	"itsyourturnring/database"
	"itsyourturnring/model"
)

type QRCodeService struct{}

func NewQRCodeService() *QRCodeService {
	return &QRCodeService{}
}

// BuildSchemeURL 根据平台构建小程序 URL Scheme
func (s *QRCodeService) BuildSchemeURL(platform, page, params string) string {
	cfg := config.GetConfig()

	switch platform {
	case "wechat":
		appID := cfg.WechatMP.AppID
		if appID == "" {
			appID = "WECHAT_APPID"
		}
		u := fmt.Sprintf("weixin://dl/business/?appid=%s&path=%s", appID, url.QueryEscape(page))
		if params != "" {
			u += "&query=" + url.QueryEscape(params)
		}
		return u

	case "alipay":
		appID := cfg.AlipayMP.AppID
		if appID == "" {
			appID = "ALIPAY_APPID"
		}
		u := fmt.Sprintf("alipays://platformapi/startapp?appId=%s&page=%s", appID, url.QueryEscape(page))
		if params != "" {
			u += "&query=" + url.QueryEscape(params)
		}
		return u

	default:
		return ""
	}
}

// Create 创建二维码记录
func (s *QRCodeService) Create(req *model.QRCodeCreateRequest) (*model.QRCode, error) {
	db := database.GetDB()

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return nil, fmt.Errorf("名称不能为空")
	}
	if req.Platform != "wechat" && req.Platform != "alipay" {
		return nil, fmt.Errorf("平台必须是 wechat 或 alipay")
	}
	if req.Page == "" {
		return nil, fmt.Errorf("页面路径不能为空")
	}

	content := s.BuildSchemeURL(req.Platform, req.Page, req.Params)

	result, err := db.Exec(`
		INSERT INTO qr_codes (name, platform, page, params, content)
		VALUES (?, ?, ?, ?, ?)`,
		req.Name, req.Platform, req.Page, req.Params, content)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return s.GetByID(id)
}

// GetByID 根据ID获取二维码
func (s *QRCodeService) GetByID(id int64) (*model.QRCode, error) {
	db := database.GetDB()

	var qr model.QRCode
	err := db.QueryRow(`
		SELECT id, name, platform, page, params, COALESCE(image_url,''), content, created_at, updated_at
		FROM qr_codes WHERE id = ?`, id).Scan(
		&qr.ID, &qr.Name, &qr.Platform, &qr.Page, &qr.Params,
		&qr.ImageURL, &qr.Content, &qr.CreatedAt, &qr.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &qr, nil
}

// List 获取二维码列表
func (s *QRCodeService) List(platform string) ([]model.QRCode, error) {
	db := database.GetDB()

	query := `SELECT id, name, platform, page, params, COALESCE(image_url,''), content, created_at, updated_at FROM qr_codes`
	var args []interface{}

	if platform != "" {
		query += ` WHERE platform = ?`
		args = append(args, platform)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []model.QRCode
	for rows.Next() {
		var qr model.QRCode
		if err := rows.Scan(&qr.ID, &qr.Name, &qr.Platform, &qr.Page, &qr.Params,
			&qr.ImageURL, &qr.Content, &qr.CreatedAt, &qr.UpdatedAt); err != nil {
			return nil, err
		}
		codes = append(codes, qr)
	}
	return codes, nil
}

// Delete 删除二维码
func (s *QRCodeService) Delete(id int64) error {
	db := database.GetDB()
	result, err := db.Exec("DELETE FROM qr_codes WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("二维码不存在")
	}
	return nil
}
