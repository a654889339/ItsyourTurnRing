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

// resolveScenePageParams 根据 scene 自动推导 page 和 params
func (s *QRCodeService) resolveScenePageParams(req *model.QRCodeCreateRequest) {
	cfg := config.GetConfig()
	pages := cfg.WechatMP.Pages

	switch req.Scene {
	case "product_view":
		req.Page = pages.Product
		if req.ProductID > 0 {
			req.Params = fmt.Sprintf("id=%d", req.ProductID)
		}
	case "product_buy":
		req.Page = pages.Product
		if req.ProductID > 0 {
			req.Params = fmt.Sprintf("id=%d&action=buy", req.ProductID)
		}
	case "order_status":
		req.Page = pages.Order
		if req.OrderNo != "" {
			req.Params = fmt.Sprintf("orderNo=%s", req.OrderNo)
		}
	case "home":
		req.Page = pages.Home
	case "custom":
		// page 和 params 由请求体直接提供
	default:
		req.Scene = "custom"
	}

	if req.Page == "" {
		req.Page = "pages/index/index"
	}
}

func (s *QRCodeService) Create(req *model.QRCodeCreateRequest) (*model.QRCode, error) {
	db := database.GetDB()

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return nil, fmt.Errorf("名称不能为空")
	}
	if req.Platform != "wechat" && req.Platform != "alipay" {
		return nil, fmt.Errorf("平台必须是 wechat 或 alipay")
	}
	if req.Scene == "" {
		req.Scene = "custom"
	}

	s.resolveScenePageParams(req)

	content := s.BuildSchemeURL(req.Platform, req.Page, req.Params)

	result, err := db.Exec(`
		INSERT INTO qr_codes (name, scene, platform, page, params, content)
		VALUES (?, ?, ?, ?, ?, ?)`,
		req.Name, req.Scene, req.Platform, req.Page, req.Params, content)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return s.GetByID(id)
}

func (s *QRCodeService) GetByID(id int64) (*model.QRCode, error) {
	db := database.GetDB()

	var qr model.QRCode
	err := db.QueryRow(`
		SELECT id, name, COALESCE(scene,'custom'), platform, page, params,
		       COALESCE(image_url,''), content, created_at, updated_at
		FROM qr_codes WHERE id = ?`, id).Scan(
		&qr.ID, &qr.Name, &qr.Scene, &qr.Platform, &qr.Page, &qr.Params,
		&qr.ImageURL, &qr.Content, &qr.CreatedAt, &qr.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &qr, nil
}

func (s *QRCodeService) List(platform string) ([]model.QRCode, error) {
	db := database.GetDB()

	query := `SELECT id, name, COALESCE(scene,'custom'), platform, page, params,
	                 COALESCE(image_url,''), content, created_at, updated_at
	          FROM qr_codes`
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
		if err := rows.Scan(&qr.ID, &qr.Name, &qr.Scene, &qr.Platform, &qr.Page, &qr.Params,
			&qr.ImageURL, &qr.Content, &qr.CreatedAt, &qr.UpdatedAt); err != nil {
			return nil, err
		}
		codes = append(codes, qr)
	}
	return codes, nil
}

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
