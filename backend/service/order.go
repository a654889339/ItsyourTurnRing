package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"itsyourturnring/database"
	"itsyourturnring/model"
)

type OrderService struct {
	productService *ProductService
}

func NewOrderService() *OrderService {
	return &OrderService{
		productService: NewProductService(),
	}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(userID int64, req *model.OrderCreateRequest) (*model.Order, error) {
	db := database.GetDB()

	// 获取收货地址
	var address model.Address
	err := db.QueryRow(`
		SELECT id, name, phone, province, city, district, detail
		FROM addresses WHERE id = ? AND user_id = ?`, req.AddressID, userID).Scan(
		&address.ID, &address.Name, &address.Phone, &address.Province,
		&address.City, &address.District, &address.Detail)
	if err != nil {
		return nil, errors.New("收货地址不存在")
	}

	// 获取购物车商品
	var cartItems []model.CartItem
	var totalPrice float64

	for _, cartID := range req.CartIDs {
		var item model.CartItem
		var product model.Product
		var specID sql.NullInt64
		var specName sql.NullString
		var priceAdj sql.NullFloat64

		err := db.QueryRow(`
			SELECT c.id, c.product_id, c.spec_id, c.quantity,
				p.name, p.price, p.main_image, p.stock,
				ps.name, ps.price_adjustment
			FROM cart_items c
			JOIN products p ON c.product_id = p.id
			LEFT JOIN product_specs ps ON c.spec_id = ps.id
			WHERE c.id = ? AND c.user_id = ?`, cartID, userID).Scan(
			&item.ID, &item.ProductID, &specID, &item.Quantity,
			&product.Name, &product.Price, &product.MainImage, &product.Stock,
			&specName, &priceAdj)
		if err != nil {
			continue
		}

		if specID.Valid {
			specIDVal := specID.Int64
			item.SpecID = &specIDVal
		}

		// 检查库存
		if product.Stock < item.Quantity {
			return nil, fmt.Errorf("商品 %s 库存不足", product.Name)
		}

		item.Product = &product
		if specName.Valid {
			item.Spec = &model.ProductSpec{
				Name:            specName.String,
				PriceAdjustment: priceAdj.Float64,
			}
		}

		price := product.Price
		if priceAdj.Valid {
			price += priceAdj.Float64
		}
		totalPrice += price * float64(item.Quantity)

		cartItems = append(cartItems, item)
	}

	if len(cartItems) == 0 {
		return nil, errors.New("购物车为空")
	}

	// 生成订单号
	orderNo := fmt.Sprintf("RG%s%04d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)

	// 开启事务
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 创建订单
	result, err := tx.Exec(`
		INSERT INTO orders (user_id, order_no, total_price, pay_price, status, pay_status,
			address_name, address_phone, address_province, address_city, address_district,
			address_detail, remark, order_source)
		VALUES (?, ?, ?, ?, 'pending', 'unpaid', ?, ?, ?, ?, ?, ?, ?, 'web')`,
		userID, orderNo, totalPrice, totalPrice,
		address.Name, address.Phone, address.Province, address.City, address.District,
		address.Detail, req.Remark)
	if err != nil {
		return nil, err
	}

	orderID, _ := result.LastInsertId()

	// 创建订单项并扣减库存
	for _, item := range cartItems {
		price := item.Product.Price
		specName := ""
		if item.Spec != nil {
			price += item.Spec.PriceAdjustment
			specName = item.Spec.Name
		}

		_, err = tx.Exec(`
			INSERT INTO order_items (order_id, product_id, spec_id, product_name, product_image, spec_name, price, quantity)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			orderID, item.ProductID, item.SpecID, item.Product.Name, item.Product.MainImage,
			specName, price, item.Quantity)
		if err != nil {
			return nil, err
		}

		// 扣减库存
		_, err = tx.Exec("UPDATE products SET stock = stock - ? WHERE id = ?", item.Quantity, item.ProductID)
		if err != nil {
			return nil, err
		}

		// 删除购物车
		_, _ = tx.Exec("DELETE FROM cart_items WHERE id = ?", item.ID)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return s.GetOrderByID(orderID, userID)
}

// GetOrderByID 根据ID获取订单
func (s *OrderService) GetOrderByID(orderID int64, userID int64) (*model.Order, error) {
	db := database.GetDB()

	var order model.Order
	var payMethod, expressCompany, expressNo, remark sql.NullString
	var payTime, shipTime, receiveTime sql.NullTime

	err := db.QueryRow(`
		SELECT id, user_id, order_no, total_price, pay_price, freight, status, pay_status,
			pay_method, pay_time, ship_time, receive_time,
			address_name, address_phone, address_province, address_city, address_district,
			address_detail, express_company, express_no, remark, order_source,
			created_at, updated_at
		FROM orders WHERE id = ? AND user_id = ?`, orderID, userID).Scan(
		&order.ID, &order.UserID, &order.OrderNo, &order.TotalPrice, &order.PayPrice,
		&order.Freight, &order.Status, &order.PayStatus, &payMethod, &payTime,
		&shipTime, &receiveTime, &order.AddressName, &order.AddressPhone,
		&order.AddressProvince, &order.AddressCity, &order.AddressDistrict,
		&order.AddressDetail, &expressCompany, &expressNo, &remark,
		&order.OrderSource, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if payMethod.Valid {
		order.PayMethod = payMethod.String
	}
	if expressCompany.Valid {
		order.ExpressCompany = expressCompany.String
	}
	if expressNo.Valid {
		order.ExpressNo = expressNo.String
	}
	if remark.Valid {
		order.Remark = remark.String
	}
	if payTime.Valid {
		order.PayTime = &payTime.Time
	}
	if shipTime.Valid {
		order.ShipTime = &shipTime.Time
	}
	if receiveTime.Valid {
		order.ReceiveTime = &receiveTime.Time
	}

	// 获取订单项
	order.Items, _ = s.GetOrderItems(orderID)

	return &order, nil
}

// GetOrderByOrderNo 根据订单号获取订单
func (s *OrderService) GetOrderByOrderNo(orderNo string) (*model.Order, error) {
	db := database.GetDB()

	var orderID, userID int64
	err := db.QueryRow("SELECT id, user_id FROM orders WHERE order_no = ?", orderNo).Scan(&orderID, &userID)
	if err != nil {
		return nil, err
	}

	return s.GetOrderByID(orderID, userID)
}

// GetOrderItems 获取订单项
func (s *OrderService) GetOrderItems(orderID int64) ([]model.OrderItem, error) {
	db := database.GetDB()

	rows, err := db.Query(`
		SELECT id, order_id, product_id, spec_id, product_name, product_image,
			spec_name, price, quantity, created_at
		FROM order_items WHERE order_id = ?`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.OrderItem
	for rows.Next() {
		var item model.OrderItem
		var specID sql.NullInt64
		var productImage, specName sql.NullString

		err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &specID,
			&item.ProductName, &productImage, &specName, &item.Price,
			&item.Quantity, &item.CreatedAt)
		if err != nil {
			continue
		}

		if specID.Valid {
			specIDVal := specID.Int64
			item.SpecID = &specIDVal
		}
		if productImage.Valid {
			item.ProductImage = productImage.String
		}
		if specName.Valid {
			item.SpecName = specName.String
		}

		items = append(items, item)
	}

	return items, nil
}

// ListOrders 订单列表
func (s *OrderService) ListOrders(query *model.PageQuery, userID int64) (*model.PageResult, error) {
	db := database.GetDB()

	where := "WHERE user_id = ?"
	args := []interface{}{userID}

	if query.Status != "" {
		where += " AND status = ?"
		args = append(args, query.Status)
	}
	if query.Keyword != "" {
		where += " AND order_no LIKE ?"
		args = append(args, "%"+query.Keyword+"%")
	}

	var total int64
	err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM orders %s", where), args...).Scan(&total)
	if err != nil {
		return nil, err
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
		SELECT id, user_id, order_no, total_price, pay_price, freight, status, pay_status,
			pay_method, pay_time, ship_time, receive_time,
			address_name, address_phone, address_province, address_city, address_district,
			address_detail, express_company, express_no, remark, order_source,
			created_at, updated_at
		FROM orders %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, where)

	args = append(args, pageSize, offset)
	rows, err := db.Query(dataSQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		var payMethod, expressCompany, expressNo, remark sql.NullString
		var payTime, shipTime, receiveTime sql.NullTime

		err := rows.Scan(
			&order.ID, &order.UserID, &order.OrderNo, &order.TotalPrice, &order.PayPrice,
			&order.Freight, &order.Status, &order.PayStatus, &payMethod, &payTime,
			&shipTime, &receiveTime, &order.AddressName, &order.AddressPhone,
			&order.AddressProvince, &order.AddressCity, &order.AddressDistrict,
			&order.AddressDetail, &expressCompany, &expressNo, &remark,
			&order.OrderSource, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			continue
		}

		if payMethod.Valid {
			order.PayMethod = payMethod.String
		}
		if expressCompany.Valid {
			order.ExpressCompany = expressCompany.String
		}
		if expressNo.Valid {
			order.ExpressNo = expressNo.String
		}
		if remark.Valid {
			order.Remark = remark.String
		}
		if payTime.Valid {
			order.PayTime = &payTime.Time
		}
		if shipTime.Valid {
			order.ShipTime = &shipTime.Time
		}
		if receiveTime.Valid {
			order.ReceiveTime = &receiveTime.Time
		}

		// 获取订单项
		order.Items, _ = s.GetOrderItems(order.ID)

		orders = append(orders, order)
	}

	return &model.PageResult{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     orders,
	}, nil
}

// UpdateOrderStatus 更新订单状态
func (s *OrderService) UpdateOrderStatus(orderID int64, userID int64, req *model.OrderUpdateStatusRequest) error {
	db := database.GetDB()

	// 验证订单所属
	var ownerID int64
	err := db.QueryRow("SELECT user_id FROM orders WHERE id = ?", orderID).Scan(&ownerID)
	if err != nil {
		return errors.New("订单不存在")
	}
	if ownerID != userID {
		return errors.New("无权修改此订单")
	}

	updateSQL := "UPDATE orders SET status = ?, updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{req.Status}

	if req.Status == "shipped" {
		updateSQL += ", ship_time = CURRENT_TIMESTAMP, express_company = ?, express_no = ?"
		args = append(args, req.ExpressCompany, req.ExpressNo)
	} else if req.Status == "received" || req.Status == "completed" {
		updateSQL += ", receive_time = CURRENT_TIMESTAMP"
	}

	updateSQL += " WHERE id = ?"
	args = append(args, orderID)

	_, err = db.Exec(updateSQL, args...)
	return err
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(orderID int64, userID int64) error {
	db := database.GetDB()

	// 获取订单状态
	var status string
	err := db.QueryRow("SELECT status FROM orders WHERE id = ? AND user_id = ?", orderID, userID).Scan(&status)
	if err != nil {
		return errors.New("订单不存在")
	}

	if status != "pending" {
		return errors.New("只能取消待付款订单")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 恢复库存
	rows, err := tx.Query("SELECT product_id, quantity FROM order_items WHERE order_id = ?", orderID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var productID int64
		var quantity int
		if err := rows.Scan(&productID, &quantity); err != nil {
			continue
		}
		_, _ = tx.Exec("UPDATE products SET stock = stock + ? WHERE id = ?", quantity, productID)
	}

	// 更新订单状态
	_, err = tx.Exec("UPDATE orders SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP WHERE id = ?", orderID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// PayOrder 支付订单(模拟)
func (s *OrderService) PayOrder(orderID int64, userID int64, payMethod string) error {
	db := database.GetDB()

	result, err := db.Exec(`
		UPDATE orders SET status = 'paid', pay_status = 'paid', pay_method = ?,
			pay_time = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ? AND status = 'pending'`,
		payMethod, orderID, userID)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("订单不存在或状态不正确")
	}

	// 更新销量
	rows, err := db.Query("SELECT product_id, quantity FROM order_items WHERE order_id = ?", orderID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var productID int64
			var quantity int
			if rows.Scan(&productID, &quantity) == nil {
				_, _ = db.Exec("UPDATE products SET sales = sales + ? WHERE id = ?", quantity, productID)
			}
		}
	}

	return nil
}

// ConfirmReceive 确认收货
func (s *OrderService) ConfirmReceive(orderID int64, userID int64) error {
	db := database.GetDB()

	result, err := db.Exec(`
		UPDATE orders SET status = 'received', receive_time = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ? AND status = 'shipped'`,
		orderID, userID)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("订单不存在或状态不正确")
	}

	return nil
}

// AdminListOrders 管理员订单列表
func (s *OrderService) AdminListOrders(query *model.PageQuery, adminID int64) (*model.PageResult, error) {
	db := database.GetDB()

	// 管理员可以查看所有订单
	where := "WHERE 1=1"
	args := []interface{}{}

	if query.Status != "" {
		where += " AND status = ?"
		args = append(args, query.Status)
	}
	if query.Keyword != "" {
		where += " AND (order_no LIKE ? OR address_name LIKE ? OR address_phone LIKE ?)"
		args = append(args, "%"+query.Keyword+"%", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	var total int64
	err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM orders %s", where), args...).Scan(&total)
	if err != nil {
		return nil, err
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
		SELECT id, user_id, order_no, total_price, pay_price, freight, status, pay_status,
			pay_method, pay_time, ship_time, receive_time,
			address_name, address_phone, address_province, address_city, address_district,
			address_detail, express_company, express_no, remark, order_source,
			created_at, updated_at
		FROM orders %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, where)

	args = append(args, pageSize, offset)
	rows, err := db.Query(dataSQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		var payMethod, expressCompany, expressNo, remark sql.NullString
		var payTime, shipTime, receiveTime sql.NullTime

		err := rows.Scan(
			&order.ID, &order.UserID, &order.OrderNo, &order.TotalPrice, &order.PayPrice,
			&order.Freight, &order.Status, &order.PayStatus, &payMethod, &payTime,
			&shipTime, &receiveTime, &order.AddressName, &order.AddressPhone,
			&order.AddressProvince, &order.AddressCity, &order.AddressDistrict,
			&order.AddressDetail, &expressCompany, &expressNo, &remark,
			&order.OrderSource, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			continue
		}

		if payMethod.Valid {
			order.PayMethod = payMethod.String
		}
		if expressCompany.Valid {
			order.ExpressCompany = expressCompany.String
		}
		if expressNo.Valid {
			order.ExpressNo = expressNo.String
		}
		if remark.Valid {
			order.Remark = remark.String
		}
		if payTime.Valid {
			order.PayTime = &payTime.Time
		}
		if shipTime.Valid {
			order.ShipTime = &shipTime.Time
		}
		if receiveTime.Valid {
			order.ReceiveTime = &receiveTime.Time
		}

		order.Items, _ = s.GetOrderItems(order.ID)
		orders = append(orders, order)
	}

	return &model.PageResult{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     orders,
	}, nil
}

// AdminUpdateOrderStatus 管理员更新订单状态（带日志）
func (s *OrderService) AdminUpdateOrderStatus(orderID int64, req *model.OrderUpdateStatusRequest) error {
	db := database.GetDB()

	var oldStatus string
	if err := db.QueryRow("SELECT status FROM orders WHERE id = ?", orderID).Scan(&oldStatus); err != nil {
		return errors.New("订单不存在")
	}

	updateSQL := "UPDATE orders SET status = ?, updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{req.Status}

	if req.Status == "shipped" {
		updateSQL += ", ship_time = CURRENT_TIMESTAMP, express_company = ?, express_no = ?"
		args = append(args, req.ExpressCompany, req.ExpressNo)
	} else if req.Status == "paid" {
		updateSQL += ", pay_status = 'paid', pay_time = CURRENT_TIMESTAMP"
	} else if req.Status == "received" || req.Status == "completed" {
		updateSQL += ", receive_time = CURRENT_TIMESTAMP"
	}

	updateSQL += " WHERE id = ?"
	args = append(args, orderID)

	if _, err := db.Exec(updateSQL, args...); err != nil {
		return err
	}

	s.logChange(orderID, "status", oldStatus, req.Status, "admin")
	return nil
}

// AdminUpdatePrice 管理员修改订单总价（带日志）
func (s *OrderService) AdminUpdatePrice(orderID int64, newPrice float64) error {
	db := database.GetDB()

	var oldPrice float64
	if err := db.QueryRow("SELECT pay_price FROM orders WHERE id = ?", orderID).Scan(&oldPrice); err != nil {
		return errors.New("订单不存在")
	}

	_, err := db.Exec("UPDATE orders SET pay_price = ?, total_price = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		newPrice, newPrice, orderID)
	if err != nil {
		return err
	}

	s.logChange(orderID, "price",
		fmt.Sprintf("%.2f", oldPrice),
		fmt.Sprintf("%.2f", newPrice),
		"admin")
	return nil
}

// AdminAppendRemark 管理员追加订单备注（带日志）
func (s *OrderService) AdminAppendRemark(orderID int64, newRemark string) error {
	db := database.GetDB()

	var oldRemark sql.NullString
	if err := db.QueryRow("SELECT remark FROM orders WHERE id = ?", orderID).Scan(&oldRemark); err != nil {
		return errors.New("订单不存在")
	}

	old := ""
	if oldRemark.Valid {
		old = oldRemark.String
	}

	_, err := db.Exec("UPDATE orders SET remark = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		newRemark, orderID)
	if err != nil {
		return err
	}

	s.logChange(orderID, "remark", old, newRemark, "admin")
	return nil
}

// GetOrderChangeLogs 获取订单变更日志
func (s *OrderService) GetOrderChangeLogs(orderID int64) ([]model.OrderChangeLog, error) {
	db := database.GetDB()

	rows, err := db.Query(`SELECT id, order_id, change_type, COALESCE(old_value,''), COALESCE(new_value,''), COALESCE(operator,''), created_at
		FROM order_change_logs WHERE order_id = ? ORDER BY created_at DESC`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.OrderChangeLog
	for rows.Next() {
		var l model.OrderChangeLog
		if err := rows.Scan(&l.ID, &l.OrderID, &l.ChangeType, &l.OldValue, &l.NewValue, &l.Operator, &l.CreatedAt); err != nil {
			continue
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (s *OrderService) logChange(orderID int64, changeType, oldVal, newVal, operator string) {
	db := database.GetDB()
	db.Exec(`INSERT INTO order_change_logs (order_id, change_type, old_value, new_value, operator) VALUES (?, ?, ?, ?, ?)`,
		orderID, changeType, oldVal, newVal, operator)
}

// CreateAdminOrder 管理员快速下单（不经购物车）
func (s *OrderService) CreateAdminOrder(userID int64, req *model.AdminOrderRequest) (*model.Order, error) {
	db := database.GetDB()

	if len(req.Items) == 0 {
		return nil, errors.New("请选择商品")
	}
	if req.AddressName == "" || req.AddressPhone == "" || req.AddressDetail == "" {
		return nil, errors.New("请填写收货信息")
	}

	orderNo := fmt.Sprintf("A%s%04d", time.Now().Format("20060102150405"), time.Now().Nanosecond()/100000)

	var totalPrice float64
	var orderItems []model.OrderItem

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			continue
		}
		var product model.Product
		var mainImage sql.NullString
		err := db.QueryRow(`SELECT id, name, price, main_image, stock FROM products WHERE id = ?`, item.ProductID).
			Scan(&product.ID, &product.Name, &product.Price, &mainImage, &product.Stock)
		if err != nil {
			return nil, fmt.Errorf("商品(ID:%d)不存在", item.ProductID)
		}
		if mainImage.Valid {
			product.MainImage = mainImage.String
		}
		if product.Stock < item.Quantity {
			return nil, fmt.Errorf("商品「%s」库存不足", product.Name)
		}

		price := product.Price
		specName := ""
		if item.SpecID != nil {
			var adj sql.NullFloat64
			var sn sql.NullString
			db.QueryRow("SELECT name, price_adjustment FROM product_specs WHERE id = ? AND product_id = ?",
				*item.SpecID, item.ProductID).Scan(&sn, &adj)
			if sn.Valid {
				specName = sn.String
			}
			if adj.Valid {
				price += adj.Float64
			}
		}

		itemTotal := price * float64(item.Quantity)
		totalPrice += itemTotal

		orderItems = append(orderItems, model.OrderItem{
			ProductID:    item.ProductID,
			SpecID:       item.SpecID,
			ProductName:  product.Name,
			ProductImage: product.MainImage,
			SpecName:     specName,
			Price:        price,
			Quantity:     item.Quantity,
		})
	}

	result, err := db.Exec(`
		INSERT INTO orders (user_id, order_no, total_price, pay_price, freight, status, pay_status,
			address_name, address_phone, address_province, address_city, address_district, address_detail,
			remark, order_source)
		VALUES (?, ?, ?, ?, 0, 'pending', 'unpaid', ?, ?, ?, ?, ?, ?, ?, 'web')`,
		userID, orderNo, totalPrice, totalPrice,
		req.AddressName, req.AddressPhone, req.AddressProvince, req.AddressCity,
		req.AddressDistrict, req.AddressDetail, req.Remark)
	if err != nil {
		return nil, err
	}

	orderID, _ := result.LastInsertId()

	for _, oi := range orderItems {
		db.Exec(`INSERT INTO order_items (order_id, product_id, spec_id, product_name, product_image, spec_name, price, quantity)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			orderID, oi.ProductID, oi.SpecID, oi.ProductName, oi.ProductImage, oi.SpecName, oi.Price, oi.Quantity)

		s.productService.UpdateStock(oi.ProductID, -oi.Quantity, orderNo)
	}

	return s.GetOrderByID(orderID, userID)
}
