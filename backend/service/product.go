package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"itsyourturnring/database"
	"itsyourturnring/model"
)

type ProductService struct{}

func NewProductService() *ProductService {
	return &ProductService{}
}

// CreateProduct 创建商品
func (s *ProductService) CreateProduct(userID int64, req *model.ProductCreateRequest) (*model.Product, error) {
	db := database.GetDB()

	result, err := db.Exec(`
		INSERT INTO products (user_id, category_id, name, description, price, original_price,
			images, main_image, material, size, color, stock, is_featured, is_new, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'available')`,
		userID, req.CategoryID, req.Name, req.Description, req.Price, req.OriginalPrice,
		req.Images, req.MainImage, req.Material, req.Size, req.Color, req.Stock,
		req.IsFeatured, req.IsNew)
	if err != nil {
		return nil, err
	}

	productID, _ := result.LastInsertId()
	return s.GetProductByID(productID)
}

// UpdateProduct 更新商品
func (s *ProductService) UpdateProduct(productID int64, userID int64, req *model.ProductCreateRequest) (*model.Product, error) {
	db := database.GetDB()

	// 验证商品所属
	var ownerID int64
	err := db.QueryRow("SELECT user_id FROM products WHERE id = ?", productID).Scan(&ownerID)
	if err != nil {
		return nil, err
	}
	if ownerID != userID {
		return nil, errors.New("无权修改此商品")
	}

	// 获取旧数据用于日志
	oldProduct, _ := s.GetProductByID(productID)

	_, err = db.Exec(`
		UPDATE products SET category_id = ?, name = ?, description = ?, price = ?,
			original_price = ?, images = ?, main_image = ?, material = ?, size = ?,
			color = ?, stock = ?, is_featured = ?, is_new = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		req.CategoryID, req.Name, req.Description, req.Price, req.OriginalPrice,
		req.Images, req.MainImage, req.Material, req.Size, req.Color, req.Stock,
		req.IsFeatured, req.IsNew, productID)
	if err != nil {
		return nil, err
	}

	// 记录变更日志
	if oldProduct != nil && oldProduct.Price != req.Price {
		s.logProductChange(productID, "price", fmt.Sprintf("%.2f", oldProduct.Price), fmt.Sprintf("%.2f", req.Price), "价格变更", "")
	}
	if oldProduct != nil && oldProduct.Stock != req.Stock {
		s.logProductChange(productID, "stock", fmt.Sprintf("%d", oldProduct.Stock), fmt.Sprintf("%d", req.Stock), "库存变更", "")
	}

	return s.GetProductByID(productID)
}

// DeleteProduct 删除商品
func (s *ProductService) DeleteProduct(productID int64, userID int64) error {
	db := database.GetDB()

	result, err := db.Exec("DELETE FROM products WHERE id = ? AND user_id = ?", productID, userID)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("商品不存在或无权删除")
	}

	return nil
}

// GetProductByID 根据ID获取商品
func (s *ProductService) GetProductByID(productID int64) (*model.Product, error) {
	db := database.GetDB()

	var product model.Product
	var images, mainImage, material, size, color sql.NullString
	var originalPrice sql.NullFloat64

	err := db.QueryRow(`
		SELECT p.id, p.user_id, p.category_id, c.name as category_name, p.name, p.description,
			p.price, p.original_price, p.images, p.main_image, p.material, p.size, p.color,
			p.stock, p.sales, p.status, p.is_featured, p.is_new, p.sort_order,
			p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = ?`, productID).Scan(
		&product.ID, &product.UserID, &product.CategoryID, &product.CategoryName,
		&product.Name, &product.Description, &product.Price, &originalPrice,
		&images, &mainImage, &material, &size, &color,
		&product.Stock, &product.Sales, &product.Status, &product.IsFeatured,
		&product.IsNew, &product.SortOrder, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if originalPrice.Valid {
		product.OriginalPrice = originalPrice.Float64
	}
	if images.Valid {
		product.Images = images.String
	}
	if mainImage.Valid {
		product.MainImage = mainImage.String
	}
	if material.Valid {
		product.Material = material.String
	}
	if size.Valid {
		product.Size = size.String
	}
	if color.Valid {
		product.Color = color.String
	}

	// 获取规格
	product.Specs, _ = s.GetProductSpecs(productID)

	return &product, nil
}

// ListProducts 商品列表
func (s *ProductService) ListProducts(query *model.PageQuery, userID int64) (*model.PageResult, error) {
	db := database.GetDB()

	// 构建查询条件
	where := "WHERE p.user_id = ?"
	args := []interface{}{userID}

	if query.Category != "" {
		where += " AND c.code = ?"
		args = append(args, query.Category)
	}
	if query.Status != "" {
		where += " AND p.status = ?"
		args = append(args, query.Status)
	}
	if query.Keyword != "" {
		where += " AND (p.name LIKE ? OR p.description LIKE ?)"
		args = append(args, "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	// 查询总数
	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id %s", where)
	err := db.QueryRow(countSQL, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// 排序
	orderBy := "ORDER BY p.sort_order ASC, p.created_at DESC"
	if query.SortBy != "" {
		dir := "ASC"
		if query.SortDir == "desc" {
			dir = "DESC"
		}
		orderBy = fmt.Sprintf("ORDER BY p.%s %s", query.SortBy, dir)
	}

	// 分页
	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 查询数据
	dataSQL := fmt.Sprintf(`
		SELECT p.id, p.user_id, p.category_id, c.name as category_name, p.name, p.description,
			p.price, p.original_price, p.images, p.main_image, p.material, p.size, p.color,
			p.stock, p.sales, p.status, p.is_featured, p.is_new, p.sort_order,
			p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		%s %s LIMIT ? OFFSET ?`, where, orderBy)

	args = append(args, pageSize, offset)
	rows, err := db.Query(dataSQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var product model.Product
		var images, mainImage, material, size, color sql.NullString
		var originalPrice sql.NullFloat64

		err := rows.Scan(
			&product.ID, &product.UserID, &product.CategoryID, &product.CategoryName,
			&product.Name, &product.Description, &product.Price, &originalPrice,
			&images, &mainImage, &material, &size, &color,
			&product.Stock, &product.Sales, &product.Status, &product.IsFeatured,
			&product.IsNew, &product.SortOrder, &product.CreatedAt, &product.UpdatedAt)
		if err != nil {
			continue
		}

		if originalPrice.Valid {
			product.OriginalPrice = originalPrice.Float64
		}
		if images.Valid {
			product.Images = images.String
		}
		if mainImage.Valid {
			product.MainImage = mainImage.String
		}
		if material.Valid {
			product.Material = material.String
		}
		if size.Valid {
			product.Size = size.String
		}
		if color.Valid {
			product.Color = color.String
		}

		products = append(products, product)
	}

	return &model.PageResult{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     products,
	}, nil
}

// ListPublicProducts 公开商品列表(无需登录)
func (s *ProductService) ListPublicProducts(query *model.PageQuery) (*model.PageResult, error) {
	db := database.GetDB()

	where := "WHERE p.status = 'available'"
	args := []interface{}{}

	if query.Category != "" {
		where += " AND c.code = ?"
		args = append(args, query.Category)
	}
	if query.Keyword != "" {
		where += " AND (p.name LIKE ? OR p.description LIKE ?)"
		args = append(args, "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id %s", where)
	err := db.QueryRow(countSQL, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	orderBy := "ORDER BY p.is_featured DESC, p.is_new DESC, p.sales DESC, p.created_at DESC"
	if query.SortBy == "price" {
		dir := "ASC"
		if query.SortDir == "desc" {
			dir = "DESC"
		}
		orderBy = fmt.Sprintf("ORDER BY p.price %s", dir)
	} else if query.SortBy == "sales" {
		orderBy = "ORDER BY p.sales DESC"
	}

	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	dataSQL := fmt.Sprintf(`
		SELECT p.id, p.category_id, c.name as category_name, p.name, p.description,
			p.price, p.original_price, p.main_image, p.material, p.stock, p.sales,
			p.is_featured, p.is_new
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		%s %s LIMIT ? OFFSET ?`, where, orderBy)

	args = append(args, pageSize, offset)
	rows, err := db.Query(dataSQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var product model.Product
		var mainImage, material sql.NullString
		var originalPrice sql.NullFloat64

		err := rows.Scan(
			&product.ID, &product.CategoryID, &product.CategoryName, &product.Name,
			&product.Description, &product.Price, &originalPrice, &mainImage,
			&material, &product.Stock, &product.Sales, &product.IsFeatured, &product.IsNew)
		if err != nil {
			continue
		}

		if originalPrice.Valid {
			product.OriginalPrice = originalPrice.Float64
		}
		if mainImage.Valid {
			product.MainImage = mainImage.String
		}
		if material.Valid {
			product.Material = material.String
		}

		products = append(products, product)
	}

	return &model.PageResult{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     products,
	}, nil
}

// UpdateProductStatus 更新商品状态
func (s *ProductService) UpdateProductStatus(productID int64, userID int64, status string) error {
	db := database.GetDB()

	result, err := db.Exec(`
		UPDATE products SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ?`, status, productID, userID)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("商品不存在或无权修改")
	}

	return nil
}

// UpdateStock 更新库存
func (s *ProductService) UpdateStock(productID int64, change int, orderNo string) error {
	db := database.GetDB()

	// 获取当前库存
	var currentStock int
	err := db.QueryRow("SELECT stock FROM products WHERE id = ?", productID).Scan(&currentStock)
	if err != nil {
		return err
	}

	newStock := currentStock + change
	if newStock < 0 {
		return errors.New("库存不足")
	}

	_, err = db.Exec("UPDATE products SET stock = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		newStock, productID)
	if err != nil {
		return err
	}

	// 记录日志
	s.logProductChange(productID, "stock", fmt.Sprintf("%d", currentStock), fmt.Sprintf("%d", newStock), "订单扣减", orderNo)

	return nil
}

// GetProductSpecs 获取商品规格
func (s *ProductService) GetProductSpecs(productID int64) ([]model.ProductSpec, error) {
	db := database.GetDB()

	rows, err := db.Query(`
		SELECT id, product_id, name, value, price_adjustment, stock, created_at
		FROM product_specs WHERE product_id = ?`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var specs []model.ProductSpec
	for rows.Next() {
		var spec model.ProductSpec
		err := rows.Scan(&spec.ID, &spec.ProductID, &spec.Name, &spec.Value,
			&spec.PriceAdjustment, &spec.Stock, &spec.CreatedAt)
		if err != nil {
			continue
		}
		specs = append(specs, spec)
	}

	return specs, nil
}

// AddProductSpec 添加商品规格
func (s *ProductService) AddProductSpec(productID int64, name, value string, priceAdj float64, stock int) (*model.ProductSpec, error) {
	db := database.GetDB()

	result, err := db.Exec(`
		INSERT INTO product_specs (product_id, name, value, price_adjustment, stock)
		VALUES (?, ?, ?, ?, ?)`, productID, name, value, priceAdj, stock)
	if err != nil {
		return nil, err
	}

	specID, _ := result.LastInsertId()

	var spec model.ProductSpec
	err = db.QueryRow(`SELECT id, product_id, name, value, price_adjustment, stock, created_at
		FROM product_specs WHERE id = ?`, specID).Scan(
		&spec.ID, &spec.ProductID, &spec.Name, &spec.Value,
		&spec.PriceAdjustment, &spec.Stock, &spec.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &spec, nil
}

// DeleteProductSpec 删除商品规格
func (s *ProductService) DeleteProductSpec(specID int64) error {
	db := database.GetDB()
	_, err := db.Exec("DELETE FROM product_specs WHERE id = ?", specID)
	return err
}

// GetCategories 获取所有分类
func (s *ProductService) GetCategories() ([]model.Category, error) {
	db := database.GetDB()

	rows, err := db.Query(`
		SELECT id, name, code, icon, sort_order, status, created_at, updated_at
		FROM categories WHERE status = 'active' ORDER BY sort_order ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var cat model.Category
		err := rows.Scan(&cat.ID, &cat.Name, &cat.Code, &cat.Icon, &cat.SortOrder,
			&cat.Status, &cat.CreatedAt, &cat.UpdatedAt)
		if err != nil {
			continue
		}
		categories = append(categories, cat)
	}

	return categories, nil
}

// GetProductChangeLogs 获取商品变更日志
func (s *ProductService) GetProductChangeLogs(productID int64) ([]model.ProductChangeLog, error) {
	db := database.GetDB()

	rows, err := db.Query(`
		SELECT id, product_id, change_type, old_value, new_value, remark, order_no, created_at
		FROM product_change_logs WHERE product_id = ? ORDER BY created_at DESC`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.ProductChangeLog
	for rows.Next() {
		var log model.ProductChangeLog
		var orderNo sql.NullString
		err := rows.Scan(&log.ID, &log.ProductID, &log.ChangeType, &log.OldValue,
			&log.NewValue, &log.Remark, &orderNo, &log.CreatedAt)
		if err != nil {
			continue
		}
		if orderNo.Valid {
			log.OrderNo = orderNo.String
		}
		logs = append(logs, log)
	}

	return logs, nil
}

func (s *ProductService) logProductChange(productID int64, changeType, oldValue, newValue, remark, orderNo string) {
	db := database.GetDB()
	_, _ = db.Exec(`
		INSERT INTO product_change_logs (product_id, change_type, old_value, new_value, remark, order_no)
		VALUES (?, ?, ?, ?, ?, ?)`, productID, changeType, oldValue, newValue, remark, orderNo)
}

// GetFeaturedProducts 获取推荐商品
func (s *ProductService) GetFeaturedProducts(limit int) ([]model.Product, error) {
	db := database.GetDB()

	rows, err := db.Query(`
		SELECT p.id, p.category_id, c.name as category_name, p.name, p.price,
			p.original_price, p.main_image, p.sales, p.is_featured, p.is_new
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.status = 'available' AND p.is_featured = TRUE
		ORDER BY p.sales DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var product model.Product
		var mainImage sql.NullString
		var originalPrice sql.NullFloat64

		err := rows.Scan(&product.ID, &product.CategoryID, &product.CategoryName,
			&product.Name, &product.Price, &originalPrice, &mainImage,
			&product.Sales, &product.IsFeatured, &product.IsNew)
		if err != nil {
			continue
		}

		if originalPrice.Valid {
			product.OriginalPrice = originalPrice.Float64
		}
		if mainImage.Valid {
			product.MainImage = mainImage.String
		}

		products = append(products, product)
	}

	return products, nil
}

// GetNewProducts 获取新品
func (s *ProductService) GetNewProducts(limit int) ([]model.Product, error) {
	db := database.GetDB()

	rows, err := db.Query(`
		SELECT p.id, p.category_id, c.name as category_name, p.name, p.price,
			p.original_price, p.main_image, p.sales, p.is_featured, p.is_new
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.status = 'available' AND p.is_new = TRUE
		ORDER BY p.created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var product model.Product
		var mainImage sql.NullString
		var originalPrice sql.NullFloat64

		err := rows.Scan(&product.ID, &product.CategoryID, &product.CategoryName,
			&product.Name, &product.Price, &originalPrice, &mainImage,
			&product.Sales, &product.IsFeatured, &product.IsNew)
		if err != nil {
			continue
		}

		if originalPrice.Valid {
			product.OriginalPrice = originalPrice.Float64
		}
		if mainImage.Valid {
			product.MainImage = mainImage.String
		}

		products = append(products, product)
	}

	return products, nil
}

// BatchUpdateProducts 批量更新商品
func (s *ProductService) BatchUpdateProducts(productIDs []int64, userID int64, updates map[string]interface{}) error {
	db := database.GetDB()

	for _, productID := range productIDs {
		// 验证所属
		var ownerID int64
		err := db.QueryRow("SELECT user_id FROM products WHERE id = ?", productID).Scan(&ownerID)
		if err != nil || ownerID != userID {
			continue
		}

		if status, ok := updates["status"].(string); ok {
			_, _ = db.Exec("UPDATE products SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
				status, productID)
		}
		if isFeatured, ok := updates["is_featured"].(bool); ok {
			_, _ = db.Exec("UPDATE products SET is_featured = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
				isFeatured, productID)
		}
	}

	return nil
}

// SearchProducts 搜索商品
func (s *ProductService) SearchProducts(keyword string, limit int) ([]model.Product, error) {
	db := database.GetDB()

	rows, err := db.Query(`
		SELECT p.id, p.category_id, c.name as category_name, p.name, p.price,
			p.original_price, p.main_image, p.sales
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.status = 'available' AND (p.name LIKE ? OR p.description LIKE ?)
		ORDER BY p.sales DESC LIMIT ?`, "%"+keyword+"%", "%"+keyword+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var product model.Product
		var mainImage sql.NullString
		var originalPrice sql.NullFloat64

		err := rows.Scan(&product.ID, &product.CategoryID, &product.CategoryName,
			&product.Name, &product.Price, &originalPrice, &mainImage, &product.Sales)
		if err != nil {
			continue
		}

		if originalPrice.Valid {
			product.OriginalPrice = originalPrice.Float64
		}
		if mainImage.Valid {
			product.MainImage = mainImage.String
		}

		products = append(products, product)
	}

	return products, nil
}

// ExportProducts 导出商品数据
func (s *ProductService) ExportProducts(userID int64) ([]byte, error) {
	query := &model.PageQuery{
		Page:     1,
		PageSize: 10000,
	}

	result, err := s.ListProducts(query, userID)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(result.Data, "", "  ")
}
