package service

import (
	"fmt"
	"time"

	"itsyourturnring/database"
	"itsyourturnring/model"
)

type ReportService struct{}

func NewReportService() *ReportService {
	return &ReportService{}
}

// GetDashboardStats 获取仪表盘统计
func (s *ReportService) GetDashboardStats(userID int64) (*model.DashboardStats, error) {
	db := database.GetDB()

	stats := &model.DashboardStats{}

	// 今日订单数和金额
	today := time.Now().Format("2006-01-02")
	db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(pay_price), 0)
		FROM orders WHERE DATE(created_at) = ? AND pay_status = 'paid'`, today).Scan(
		&stats.TodayOrders, &stats.TodayAmount)

	// 商品总数
	db.QueryRow("SELECT COUNT(*) FROM products WHERE user_id = ?", userID).Scan(&stats.TotalProducts)

	// 用户总数 (管理员视角)
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)

	// 待处理订单数
	db.QueryRow(`SELECT COUNT(*) FROM orders WHERE status IN ('pending', 'paid')`).Scan(&stats.PendingOrders)

	// 低库存商品数
	db.QueryRow("SELECT COUNT(*) FROM products WHERE user_id = ? AND stock < 10 AND status = 'available'", userID).Scan(&stats.LowStockProducts)

	return stats, nil
}

// GetSalesReport 获取销售报表
func (s *ReportService) GetSalesReport(startDate, endDate string) ([]model.SalesReportData, error) {
	db := database.GetDB()

	rows, err := db.Query(`
		SELECT DATE(created_at) as date, COUNT(*) as orders, COALESCE(SUM(pay_price), 0) as amount
		FROM orders
		WHERE DATE(created_at) BETWEEN ? AND ? AND pay_status = 'paid'
		GROUP BY DATE(created_at)
		ORDER BY date ASC`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []model.SalesReportData
	for rows.Next() {
		var report model.SalesReportData
		if rows.Scan(&report.Date, &report.Orders, &report.Amount) == nil {
			reports = append(reports, report)
		}
	}

	return reports, nil
}

// GetProductSalesRank 获取商品销量排行
func (s *ReportService) GetProductSalesRank(startDate, endDate string, limit int) ([]model.ProductSalesRank, error) {
	db := database.GetDB()

	if limit <= 0 {
		limit = 10
	}

	rows, err := db.Query(`
		SELECT oi.product_id, oi.product_name, SUM(oi.quantity) as sales, SUM(oi.price * oi.quantity) as amount
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		WHERE DATE(o.created_at) BETWEEN ? AND ? AND o.pay_status = 'paid'
		GROUP BY oi.product_id, oi.product_name
		ORDER BY sales DESC
		LIMIT ?`, startDate, endDate, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ranks []model.ProductSalesRank
	for rows.Next() {
		var rank model.ProductSalesRank
		if rows.Scan(&rank.ProductID, &rank.ProductName, &rank.Sales, &rank.Amount) == nil {
			ranks = append(ranks, rank)
		}
	}

	return ranks, nil
}

// GetCategorySalesReport 获取分类销售报表
func (s *ReportService) GetCategorySalesReport(startDate, endDate string) ([]map[string]interface{}, error) {
	db := database.GetDB()

	rows, err := db.Query(`
		SELECT c.name as category, COUNT(DISTINCT o.id) as orders,
			SUM(oi.quantity) as products, SUM(oi.price * oi.quantity) as amount
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		JOIN products p ON oi.product_id = p.id
		JOIN categories c ON p.category_id = c.id
		WHERE DATE(o.created_at) BETWEEN ? AND ? AND o.pay_status = 'paid'
		GROUP BY c.id, c.name
		ORDER BY amount DESC`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []map[string]interface{}
	for rows.Next() {
		var category string
		var orders, products int
		var amount float64
		if rows.Scan(&category, &orders, &products, &amount) == nil {
			reports = append(reports, map[string]interface{}{
				"category": category,
				"orders":   orders,
				"products": products,
				"amount":   amount,
			})
		}
	}

	return reports, nil
}

// GetSalesTrend 获取销售趋势
func (s *ReportService) GetSalesTrend(period string, userID int64) ([]model.SalesReportData, error) {
	db := database.GetDB()

	var dateFormat, startDate string
	now := time.Now()

	switch period {
	case "week":
		dateFormat = "%Y-%m-%d"
		startDate = now.AddDate(0, 0, -7).Format("2006-01-02")
	case "month":
		dateFormat = "%Y-%m-%d"
		startDate = now.AddDate(0, -1, 0).Format("2006-01-02")
	case "year":
		dateFormat = "%Y-%m"
		startDate = now.AddDate(-1, 0, 0).Format("2006-01-02")
	default:
		dateFormat = "%Y-%m-%d"
		startDate = now.AddDate(0, 0, -7).Format("2006-01-02")
	}

	query := fmt.Sprintf(`
		SELECT strftime('%s', created_at) as date, COUNT(*) as orders, COALESCE(SUM(pay_price), 0) as amount
		FROM orders
		WHERE created_at >= ? AND pay_status = 'paid'
		GROUP BY strftime('%s', created_at)
		ORDER BY date ASC`, dateFormat, dateFormat)

	rows, err := db.Query(query, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []model.SalesReportData
	for rows.Next() {
		var report model.SalesReportData
		if rows.Scan(&report.Date, &report.Orders, &report.Amount) == nil {
			reports = append(reports, report)
		}
	}

	return reports, nil
}

// GetOrderStatusStats 获取订单状态统计
func (s *ReportService) GetOrderStatusStats() (map[string]int, error) {
	db := database.GetDB()

	rows, err := db.Query(`
		SELECT status, COUNT(*) as count
		FROM orders
		GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if rows.Scan(&status, &count) == nil {
			stats[status] = count
		}
	}

	return stats, nil
}

// GetProductStockAlert 获取库存预警商品
func (s *ReportService) GetProductStockAlert(userID int64, threshold int) ([]model.Product, error) {
	db := database.GetDB()

	if threshold <= 0 {
		threshold = 10
	}

	rows, err := db.Query(`
		SELECT p.id, p.name, p.main_image, p.stock, p.status, c.name as category_name
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.user_id = ? AND p.stock < ? AND p.status = 'available'
		ORDER BY p.stock ASC`, userID, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var product model.Product
		if rows.Scan(&product.ID, &product.Name, &product.MainImage, &product.Stock,
			&product.Status, &product.CategoryName) == nil {
			products = append(products, product)
		}
	}

	return products, nil
}

// GetRevenueByPayMethod 获取支付方式统计
func (s *ReportService) GetRevenueByPayMethod(startDate, endDate string) ([]map[string]interface{}, error) {
	db := database.GetDB()

	rows, err := db.Query(`
		SELECT COALESCE(pay_method, 'unknown') as method, COUNT(*) as orders, SUM(pay_price) as amount
		FROM orders
		WHERE DATE(created_at) BETWEEN ? AND ? AND pay_status = 'paid'
		GROUP BY pay_method`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []map[string]interface{}
	for rows.Next() {
		var method string
		var orders int
		var amount float64
		if rows.Scan(&method, &orders, &amount) == nil {
			stats = append(stats, map[string]interface{}{
				"method": method,
				"orders": orders,
				"amount": amount,
			})
		}
	}

	return stats, nil
}
