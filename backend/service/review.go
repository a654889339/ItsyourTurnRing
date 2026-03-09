package service

import (
	"database/sql"
	"errors"

	"itsyourturnring/database"
	"itsyourturnring/model"
)

type ReviewService struct{}

func NewReviewService() *ReviewService {
	return &ReviewService{}
}

// CreateReview 创建评价
func (s *ReviewService) CreateReview(userID int64, review *model.Review) (*model.Review, error) {
	db := database.GetDB()

	// 检查订单是否存在且已完成
	var orderStatus, payStatus string
	var orderUserID int64
	err := db.QueryRow("SELECT user_id, status, pay_status FROM orders WHERE id = ?", review.OrderID).Scan(
		&orderUserID, &orderStatus, &payStatus)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	if orderUserID != userID {
		return nil, errors.New("无权评价此订单")
	}
	if orderStatus != "received" && orderStatus != "completed" {
		return nil, errors.New("订单未完成，无法评价")
	}

	// 检查是否已评价
	var exists int
	err = db.QueryRow("SELECT 1 FROM reviews WHERE user_id = ? AND order_id = ? AND product_id = ?",
		userID, review.OrderID, review.ProductID).Scan(&exists)
	if err == nil {
		return nil, errors.New("已评价过此商品")
	}

	result, err := db.Exec(`
		INSERT INTO reviews (user_id, product_id, order_id, rating, content, images, is_anonymous)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, review.ProductID, review.OrderID, review.Rating, review.Content,
		review.Images, review.IsAnonymous)
	if err != nil {
		return nil, err
	}

	reviewID, _ := result.LastInsertId()
	return s.GetReviewByID(reviewID)
}

// GetReviewByID 根据ID获取评价
func (s *ReviewService) GetReviewByID(reviewID int64) (*model.Review, error) {
	db := database.GetDB()

	var review model.Review
	var images sql.NullString

	err := db.QueryRow(`
		SELECT r.id, r.user_id, r.product_id, r.order_id, r.rating, r.content,
			r.images, r.is_anonymous, r.created_at, u.username, u.avatar
		FROM reviews r
		LEFT JOIN users u ON r.user_id = u.id
		WHERE r.id = ?`, reviewID).Scan(
		&review.ID, &review.UserID, &review.ProductID, &review.OrderID,
		&review.Rating, &review.Content, &images, &review.IsAnonymous,
		&review.CreatedAt, &review.Username, &review.Avatar)
	if err != nil {
		return nil, err
	}

	if images.Valid {
		review.Images = images.String
	}

	// 匿名处理
	if review.IsAnonymous {
		review.Username = "匿名用户"
		review.Avatar = ""
	}

	return &review, nil
}

// ListProductReviews 商品评价列表
func (s *ReviewService) ListProductReviews(productID int64, page, pageSize int) (*model.PageResult, error) {
	db := database.GetDB()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var total int64
	err := db.QueryRow("SELECT COUNT(*) FROM reviews WHERE product_id = ?", productID).Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize

	rows, err := db.Query(`
		SELECT r.id, r.user_id, r.product_id, r.order_id, r.rating, r.content,
			r.images, r.is_anonymous, r.created_at, u.username, u.avatar
		FROM reviews r
		LEFT JOIN users u ON r.user_id = u.id
		WHERE r.product_id = ?
		ORDER BY r.created_at DESC
		LIMIT ? OFFSET ?`, productID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []model.Review
	for rows.Next() {
		var review model.Review
		var images, avatar sql.NullString

		err := rows.Scan(
			&review.ID, &review.UserID, &review.ProductID, &review.OrderID,
			&review.Rating, &review.Content, &images, &review.IsAnonymous,
			&review.CreatedAt, &review.Username, &avatar)
		if err != nil {
			continue
		}

		if images.Valid {
			review.Images = images.String
		}
		if avatar.Valid {
			review.Avatar = avatar.String
		}

		if review.IsAnonymous {
			review.Username = "匿名用户"
			review.Avatar = ""
		}

		reviews = append(reviews, review)
	}

	return &model.PageResult{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     reviews,
	}, nil
}

// GetProductReviewStats 获取商品评价统计
func (s *ReviewService) GetProductReviewStats(productID int64) (map[string]interface{}, error) {
	db := database.GetDB()

	var total int
	var avgRating float64
	err := db.QueryRow(`
		SELECT COUNT(*), COALESCE(AVG(rating), 0)
		FROM reviews WHERE product_id = ?`, productID).Scan(&total, &avgRating)
	if err != nil {
		return nil, err
	}

	// 各星级数量
	rows, err := db.Query(`
		SELECT rating, COUNT(*) FROM reviews
		WHERE product_id = ?
		GROUP BY rating`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ratingCounts := make(map[int]int)
	for i := 1; i <= 5; i++ {
		ratingCounts[i] = 0
	}
	for rows.Next() {
		var rating, count int
		if rows.Scan(&rating, &count) == nil {
			ratingCounts[rating] = count
		}
	}

	return map[string]interface{}{
		"total":        total,
		"avg_rating":   avgRating,
		"rating_counts": ratingCounts,
	}, nil
}

// DeleteReview 删除评价 (管理员)
func (s *ReviewService) DeleteReview(reviewID int64) error {
	db := database.GetDB()

	result, err := db.Exec("DELETE FROM reviews WHERE id = ?", reviewID)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("评价不存在")
	}

	return nil
}

// ListUserReviews 用户评价列表
func (s *ReviewService) ListUserReviews(userID int64, page, pageSize int) (*model.PageResult, error) {
	db := database.GetDB()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var total int64
	err := db.QueryRow("SELECT COUNT(*) FROM reviews WHERE user_id = ?", userID).Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize

	rows, err := db.Query(`
		SELECT r.id, r.user_id, r.product_id, r.order_id, r.rating, r.content,
			r.images, r.is_anonymous, r.created_at, p.name, p.main_image
		FROM reviews r
		LEFT JOIN products p ON r.product_id = p.id
		WHERE r.user_id = ?
		ORDER BY r.created_at DESC
		LIMIT ? OFFSET ?`, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []map[string]interface{}
	for rows.Next() {
		var review model.Review
		var images, productName, productImage sql.NullString

		err := rows.Scan(
			&review.ID, &review.UserID, &review.ProductID, &review.OrderID,
			&review.Rating, &review.Content, &images, &review.IsAnonymous,
			&review.CreatedAt, &productName, &productImage)
		if err != nil {
			continue
		}

		if images.Valid {
			review.Images = images.String
		}

		reviewMap := map[string]interface{}{
			"id":           review.ID,
			"product_id":   review.ProductID,
			"order_id":     review.OrderID,
			"rating":       review.Rating,
			"content":      review.Content,
			"images":       review.Images,
			"is_anonymous": review.IsAnonymous,
			"created_at":   review.CreatedAt,
		}
		if productName.Valid {
			reviewMap["product_name"] = productName.String
		}
		if productImage.Valid {
			reviewMap["product_image"] = productImage.String
		}

		reviews = append(reviews, reviewMap)
	}

	return &model.PageResult{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     reviews,
	}, nil
}
