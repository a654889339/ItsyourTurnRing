package service

import (
	"errors"

	"itsyourturnring/database"
	"itsyourturnring/model"
)

type BannerService struct{}

func NewBannerService() *BannerService {
	return &BannerService{}
}

// CreateBanner 创建轮播图
func (s *BannerService) CreateBanner(banner *model.Banner) (*model.Banner, error) {
	db := database.GetDB()

	result, err := db.Exec(`
		INSERT INTO banners (title, image, link, sort_order, status)
		VALUES (?, ?, ?, ?, ?)`,
		banner.Title, banner.Image, banner.Link, banner.SortOrder, banner.Status)
	if err != nil {
		return nil, err
	}

	bannerID, _ := result.LastInsertId()
	return s.GetBannerByID(bannerID)
}

// UpdateBanner 更新轮播图
func (s *BannerService) UpdateBanner(bannerID int64, banner *model.Banner) (*model.Banner, error) {
	db := database.GetDB()

	_, err := db.Exec(`
		UPDATE banners SET title = ?, image = ?, link = ?, sort_order = ?,
			status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		banner.Title, banner.Image, banner.Link, banner.SortOrder, banner.Status, bannerID)
	if err != nil {
		return nil, err
	}

	return s.GetBannerByID(bannerID)
}

// DeleteBanner 删除轮播图
func (s *BannerService) DeleteBanner(bannerID int64) error {
	db := database.GetDB()

	result, err := db.Exec("DELETE FROM banners WHERE id = ?", bannerID)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("轮播图不存在")
	}

	return nil
}

// GetBannerByID 根据ID获取轮播图
func (s *BannerService) GetBannerByID(bannerID int64) (*model.Banner, error) {
	db := database.GetDB()

	var banner model.Banner
	err := db.QueryRow(`
		SELECT id, title, image, link, sort_order, status, created_at, updated_at
		FROM banners WHERE id = ?`, bannerID).Scan(
		&banner.ID, &banner.Title, &banner.Image, &banner.Link,
		&banner.SortOrder, &banner.Status, &banner.CreatedAt, &banner.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &banner, nil
}

// ListBanners 轮播图列表
func (s *BannerService) ListBanners(status string) ([]model.Banner, error) {
	db := database.GetDB()

	query := "SELECT id, title, image, link, sort_order, status, created_at, updated_at FROM banners"
	args := []interface{}{}

	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}

	query += " ORDER BY sort_order ASC, created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var banners []model.Banner
	for rows.Next() {
		var banner model.Banner
		err := rows.Scan(
			&banner.ID, &banner.Title, &banner.Image, &banner.Link,
			&banner.SortOrder, &banner.Status, &banner.CreatedAt, &banner.UpdatedAt)
		if err != nil {
			continue
		}
		banners = append(banners, banner)
	}

	return banners, nil
}

// GetActiveBanners 获取激活的轮播图
func (s *BannerService) GetActiveBanners() ([]model.Banner, error) {
	return s.ListBanners("active")
}

// UpdateBannerStatus 更新轮播图状态
func (s *BannerService) UpdateBannerStatus(bannerID int64, status string) error {
	db := database.GetDB()

	result, err := db.Exec(`
		UPDATE banners SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, status, bannerID)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("轮播图不存在")
	}

	return nil
}
