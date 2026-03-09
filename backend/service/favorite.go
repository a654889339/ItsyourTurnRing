package service

import (
	"database/sql"
	"errors"

	"itsyourturnring/database"
	"itsyourturnring/model"
)

type FavoriteService struct{}

func NewFavoriteService() *FavoriteService {
	return &FavoriteService{}
}

// AddFavorite 添加收藏
func (s *FavoriteService) AddFavorite(userID int64, productID int64) error {
	db := database.GetDB()

	// 检查商品是否存在
	var exists int
	err := db.QueryRow("SELECT 1 FROM products WHERE id = ?", productID).Scan(&exists)
	if err != nil {
		return errors.New("商品不存在")
	}

	_, err = db.Exec(`
		INSERT OR IGNORE INTO favorites (user_id, product_id)
		VALUES (?, ?)`, userID, productID)
	return err
}

// RemoveFavorite 取消收藏
func (s *FavoriteService) RemoveFavorite(userID int64, productID int64) error {
	db := database.GetDB()

	result, err := db.Exec("DELETE FROM favorites WHERE user_id = ? AND product_id = ?", userID, productID)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("收藏不存在")
	}

	return nil
}

// IsFavorite 检查是否收藏
func (s *FavoriteService) IsFavorite(userID int64, productID int64) (bool, error) {
	db := database.GetDB()

	var exists int
	err := db.QueryRow("SELECT 1 FROM favorites WHERE user_id = ? AND product_id = ?", userID, productID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// ListFavorites 收藏列表
func (s *FavoriteService) ListFavorites(userID int64, page, pageSize int) (*model.PageResult, error) {
	db := database.GetDB()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	// 查询总数
	var total int64
	err := db.QueryRow("SELECT COUNT(*) FROM favorites WHERE user_id = ?", userID).Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize

	rows, err := db.Query(`
		SELECT f.id, f.user_id, f.product_id, f.created_at,
			p.name, p.price, p.original_price, p.main_image, p.stock, p.status
		FROM favorites f
		JOIN products p ON f.product_id = p.id
		WHERE f.user_id = ?
		ORDER BY f.created_at DESC
		LIMIT ? OFFSET ?`, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favorites []model.Favorite
	for rows.Next() {
		var fav model.Favorite
		var product model.Product
		var originalPrice sql.NullFloat64
		var mainImage sql.NullString

		err := rows.Scan(
			&fav.ID, &fav.UserID, &fav.ProductID, &fav.CreatedAt,
			&product.Name, &product.Price, &originalPrice, &mainImage,
			&product.Stock, &product.Status)
		if err != nil {
			continue
		}

		if originalPrice.Valid {
			product.OriginalPrice = originalPrice.Float64
		}
		if mainImage.Valid {
			product.MainImage = mainImage.String
		}
		product.ID = fav.ProductID
		fav.Product = &product

		favorites = append(favorites, fav)
	}

	return &model.PageResult{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     favorites,
	}, nil
}

// GetFavoriteCount 获取收藏数量
func (s *FavoriteService) GetFavoriteCount(userID int64) (int, error) {
	db := database.GetDB()

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM favorites WHERE user_id = ?", userID).Scan(&count)
	return count, err
}

// BatchCheckFavorites 批量检查收藏状态
func (s *FavoriteService) BatchCheckFavorites(userID int64, productIDs []int64) (map[int64]bool, error) {
	db := database.GetDB()

	result := make(map[int64]bool)
	for _, id := range productIDs {
		result[id] = false
	}

	if len(productIDs) == 0 {
		return result, nil
	}

	// 构建IN查询
	placeholders := ""
	args := []interface{}{userID}
	for i, id := range productIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, id)
	}

	rows, err := db.Query(
		"SELECT product_id FROM favorites WHERE user_id = ? AND product_id IN ("+placeholders+")",
		args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var productID int64
		if rows.Scan(&productID) == nil {
			result[productID] = true
		}
	}

	return result, nil
}
