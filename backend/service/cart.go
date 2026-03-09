package service

import (
	"database/sql"
	"errors"
	"fmt"

	"itsyourturnring/database"
	"itsyourturnring/model"
)

type CartService struct{}

func NewCartService() *CartService {
	return &CartService{}
}

// AddToCart 添加到购物车
func (s *CartService) AddToCart(userID int64, req *model.CartAddRequest) (*model.CartItem, error) {
	db := database.GetDB()

	// 检查商品是否存在且有库存
	var stock int
	var status string
	err := db.QueryRow("SELECT stock, status FROM products WHERE id = ?", req.ProductID).Scan(&stock, &status)
	if err != nil {
		return nil, errors.New("商品不存在")
	}
	if status != "available" {
		return nil, errors.New("商品已下架")
	}
	if stock < req.Quantity {
		return nil, errors.New("库存不足")
	}

	// 检查购物车是否已有该商品
	var existingID int64
	var existingQty int
	query := "SELECT id, quantity FROM cart_items WHERE user_id = ? AND product_id = ?"
	args := []interface{}{userID, req.ProductID}

	if req.SpecID != nil {
		query += " AND spec_id = ?"
		args = append(args, *req.SpecID)
	} else {
		query += " AND spec_id IS NULL"
	}

	err = db.QueryRow(query, args...).Scan(&existingID, &existingQty)
	if err == nil {
		// 更新数量
		newQty := existingQty + req.Quantity
		if newQty > stock {
			return nil, errors.New("超出库存数量")
		}
		_, err = db.Exec("UPDATE cart_items SET quantity = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			newQty, existingID)
		if err != nil {
			return nil, err
		}
		return s.GetCartItem(existingID)
	}

	// 新增购物车项
	result, err := db.Exec(`
		INSERT INTO cart_items (user_id, product_id, spec_id, quantity)
		VALUES (?, ?, ?, ?)`, userID, req.ProductID, req.SpecID, req.Quantity)
	if err != nil {
		return nil, err
	}

	cartID, _ := result.LastInsertId()
	return s.GetCartItem(cartID)
}

// GetCartItem 获取单个购物车项
func (s *CartService) GetCartItem(cartID int64) (*model.CartItem, error) {
	db := database.GetDB()

	var item model.CartItem
	var specID sql.NullInt64

	err := db.QueryRow(`
		SELECT id, user_id, product_id, spec_id, quantity, created_at, updated_at
		FROM cart_items WHERE id = ?`, cartID).Scan(
		&item.ID, &item.UserID, &item.ProductID, &specID, &item.Quantity,
		&item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if specID.Valid {
		specIDVal := specID.Int64
		item.SpecID = &specIDVal
	}

	// 获取商品信息
	productService := NewProductService()
	item.Product, _ = productService.GetProductByID(item.ProductID)

	// 获取规格信息
	if item.SpecID != nil {
		var spec model.ProductSpec
		err = db.QueryRow(`
			SELECT id, product_id, name, value, price_adjustment, stock
			FROM product_specs WHERE id = ?`, *item.SpecID).Scan(
			&spec.ID, &spec.ProductID, &spec.Name, &spec.Value,
			&spec.PriceAdjustment, &spec.Stock)
		if err == nil {
			item.Spec = &spec
		}
	}

	return &item, nil
}

// ListCartItems 购物车列表
func (s *CartService) ListCartItems(userID int64) ([]model.CartItem, error) {
	db := database.GetDB()

	rows, err := db.Query(`
		SELECT c.id, c.user_id, c.product_id, c.spec_id, c.quantity, c.created_at, c.updated_at,
			p.id, p.name, p.price, p.original_price, p.main_image, p.stock, p.status,
			ps.id, ps.name, ps.value, ps.price_adjustment, ps.stock
		FROM cart_items c
		JOIN products p ON c.product_id = p.id
		LEFT JOIN product_specs ps ON c.spec_id = ps.id
		WHERE c.user_id = ?
		ORDER BY c.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.CartItem
	for rows.Next() {
		var item model.CartItem
		var product model.Product
		var specID, psID sql.NullInt64
		var originalPrice, psPrice sql.NullFloat64
		var mainImage, psName, psValue sql.NullString
		var psStock sql.NullInt64

		err := rows.Scan(
			&item.ID, &item.UserID, &item.ProductID, &specID, &item.Quantity,
			&item.CreatedAt, &item.UpdatedAt,
			&product.ID, &product.Name, &product.Price, &originalPrice, &mainImage,
			&product.Stock, &product.Status,
			&psID, &psName, &psValue, &psPrice, &psStock)
		if err != nil {
			continue
		}

		if specID.Valid {
			specIDVal := specID.Int64
			item.SpecID = &specIDVal
		}
		if originalPrice.Valid {
			product.OriginalPrice = originalPrice.Float64
		}
		if mainImage.Valid {
			product.MainImage = mainImage.String
		}
		item.Product = &product

		if psID.Valid {
			spec := &model.ProductSpec{
				ID:        psID.Int64,
				ProductID: product.ID,
			}
			if psName.Valid {
				spec.Name = psName.String
			}
			if psValue.Valid {
				spec.Value = psValue.String
			}
			if psPrice.Valid {
				spec.PriceAdjustment = psPrice.Float64
			}
			if psStock.Valid {
				spec.Stock = int(psStock.Int64)
			}
			item.Spec = spec
		}

		items = append(items, item)
	}

	return items, nil
}

// UpdateCartQuantity 更新购物车数量
func (s *CartService) UpdateCartQuantity(cartID int64, userID int64, quantity int) error {
	db := database.GetDB()

	if quantity <= 0 {
		return s.RemoveFromCart(cartID, userID)
	}

	// 检查库存
	var productID int64
	var specID sql.NullInt64
	err := db.QueryRow("SELECT product_id, spec_id FROM cart_items WHERE id = ? AND user_id = ?",
		cartID, userID).Scan(&productID, &specID)
	if err != nil {
		return errors.New("购物车项不存在")
	}

	var stock int
	if specID.Valid {
		err = db.QueryRow("SELECT stock FROM product_specs WHERE id = ?", specID.Int64).Scan(&stock)
	} else {
		err = db.QueryRow("SELECT stock FROM products WHERE id = ?", productID).Scan(&stock)
	}
	if err != nil {
		return err
	}

	if quantity > stock {
		return errors.New("超出库存数量")
	}

	_, err = db.Exec("UPDATE cart_items SET quantity = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?",
		quantity, cartID, userID)
	return err
}

// RemoveFromCart 从购物车移除
func (s *CartService) RemoveFromCart(cartID int64, userID int64) error {
	db := database.GetDB()

	result, err := db.Exec("DELETE FROM cart_items WHERE id = ? AND user_id = ?", cartID, userID)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("购物车项不存在")
	}

	return nil
}

// ClearCart 清空购物车
func (s *CartService) ClearCart(userID int64) error {
	db := database.GetDB()
	_, err := db.Exec("DELETE FROM cart_items WHERE user_id = ?", userID)
	return err
}

// GetCartCount 获取购物车数量
func (s *CartService) GetCartCount(userID int64) (int, error) {
	db := database.GetDB()

	var count int
	err := db.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE user_id = ?", userID).Scan(&count)
	return count, err
}

// GetCartTotal 获取购物车总价
func (s *CartService) GetCartTotal(userID int64) (float64, error) {
	db := database.GetDB()

	rows, err := db.Query(`
		SELECT p.price, COALESCE(ps.price_adjustment, 0), c.quantity
		FROM cart_items c
		JOIN products p ON c.product_id = p.id
		LEFT JOIN product_specs ps ON c.spec_id = ps.id
		WHERE c.user_id = ? AND p.status = 'available'`, userID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var total float64
	for rows.Next() {
		var price, priceAdj float64
		var quantity int
		if rows.Scan(&price, &priceAdj, &quantity) == nil {
			total += (price + priceAdj) * float64(quantity)
		}
	}

	return total, nil
}

// CheckCartItems 检查购物车商品状态
func (s *CartService) CheckCartItems(userID int64) ([]map[string]interface{}, error) {
	db := database.GetDB()

	rows, err := db.Query(`
		SELECT c.id, c.quantity, p.name, p.stock, p.status
		FROM cart_items c
		JOIN products p ON c.product_id = p.id
		WHERE c.user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []map[string]interface{}
	for rows.Next() {
		var cartID int64
		var quantity, stock int
		var name, status string
		if rows.Scan(&cartID, &quantity, &name, &stock, &status) != nil {
			continue
		}

		if status != "available" {
			issues = append(issues, map[string]interface{}{
				"cart_id": cartID,
				"name":    name,
				"issue":   "商品已下架",
			})
		} else if stock < quantity {
			issues = append(issues, map[string]interface{}{
				"cart_id": cartID,
				"name":    name,
				"issue":   fmt.Sprintf("库存不足，当前库存%d", stock),
			})
		}
	}

	return issues, nil
}
