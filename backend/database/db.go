package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"itsyourturnring/config"
)

var DB *sql.DB

func InitDB() error {
	cfg := config.GetConfig()

	var err error
	if cfg.Database.Driver == "mysql" {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Database.MySQL.User,
			cfg.Database.MySQL.Password,
			cfg.Database.MySQL.Host,
			cfg.Database.MySQL.Port,
			cfg.Database.MySQL.Database,
		)
		DB, err = sql.Open("mysql", dsn)
	} else {
		// 确保数据目录存在
		dbPath := cfg.Database.SQLitePath
		if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
			return err
		}
		DB, err = sql.Open("sqlite3", dbPath)
	}

	if err != nil {
		return err
	}

	if err := DB.Ping(); err != nil {
		return err
	}

	return createTables()
}

func createTables() error {
	tables := []string{
		// 用户表
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username VARCHAR(50) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			email VARCHAR(100) UNIQUE,
			phone VARCHAR(20),
			avatar VARCHAR(500),
			role VARCHAR(20) DEFAULT 'user',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// 商品分类表
		`CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(50) NOT NULL,
			code VARCHAR(50) UNIQUE NOT NULL,
			icon VARCHAR(500),
			sort_order INTEGER DEFAULT 0,
			status VARCHAR(20) DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// 商品表 (手链、项链、配饰等)
		`CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			category_id INTEGER NOT NULL,
			name VARCHAR(200) NOT NULL,
			description TEXT,
			price DECIMAL(10, 2) NOT NULL,
			original_price DECIMAL(10, 2),
			images TEXT,
			main_image VARCHAR(500),
			material VARCHAR(100),
			size VARCHAR(50),
			color VARCHAR(50),
			stock INTEGER DEFAULT 0,
			sales INTEGER DEFAULT 0,
			status VARCHAR(20) DEFAULT 'available',
			is_featured BOOLEAN DEFAULT FALSE,
			is_new BOOLEAN DEFAULT FALSE,
			sort_order INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (category_id) REFERENCES categories(id)
		)`,

		// 商品规格表
		`CREATE TABLE IF NOT EXISTS product_specs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			product_id INTEGER NOT NULL,
			name VARCHAR(100) NOT NULL,
			value VARCHAR(200) NOT NULL,
			price_adjustment DECIMAL(10, 2) DEFAULT 0,
			stock INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
		)`,

		// 购物车表
		`CREATE TABLE IF NOT EXISTS cart_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			spec_id INTEGER,
			quantity INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (product_id) REFERENCES products(id),
			FOREIGN KEY (spec_id) REFERENCES product_specs(id)
		)`,

		// 收货地址表
		`CREATE TABLE IF NOT EXISTS addresses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name VARCHAR(50) NOT NULL,
			phone VARCHAR(20) NOT NULL,
			province VARCHAR(50) NOT NULL,
			city VARCHAR(50) NOT NULL,
			district VARCHAR(50) NOT NULL,
			detail VARCHAR(200) NOT NULL,
			is_default BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,

		// 订单表
		`CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			order_no VARCHAR(50) UNIQUE NOT NULL,
			total_price DECIMAL(10, 2) NOT NULL,
			pay_price DECIMAL(10, 2) NOT NULL,
			freight DECIMAL(10, 2) DEFAULT 0,
			status VARCHAR(20) DEFAULT 'pending',
			pay_status VARCHAR(20) DEFAULT 'unpaid',
			pay_method VARCHAR(20),
			pay_time DATETIME,
			ship_time DATETIME,
			receive_time DATETIME,
			address_name VARCHAR(50),
			address_phone VARCHAR(20),
			address_province VARCHAR(50),
			address_city VARCHAR(50),
			address_district VARCHAR(50),
			address_detail VARCHAR(200),
			express_company VARCHAR(50),
			express_no VARCHAR(50),
			remark TEXT,
			order_source VARCHAR(20) DEFAULT 'web',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,

		// 订单项表
		`CREATE TABLE IF NOT EXISTS order_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			spec_id INTEGER,
			product_name VARCHAR(200) NOT NULL,
			product_image VARCHAR(500),
			spec_name VARCHAR(100),
			price DECIMAL(10, 2) NOT NULL,
			quantity INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
			FOREIGN KEY (product_id) REFERENCES products(id)
		)`,

		// 商品收藏表
		`CREATE TABLE IF NOT EXISTS favorites (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, product_id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (product_id) REFERENCES products(id)
		)`,

		// 商品评价表
		`CREATE TABLE IF NOT EXISTS reviews (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			order_id INTEGER NOT NULL,
			rating INTEGER NOT NULL,
			content TEXT,
			images TEXT,
			is_anonymous BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (product_id) REFERENCES products(id),
			FOREIGN KEY (order_id) REFERENCES orders(id)
		)`,

		// 轮播图表
		`CREATE TABLE IF NOT EXISTS banners (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title VARCHAR(100),
			image VARCHAR(500) NOT NULL,
			link VARCHAR(500),
			sort_order INTEGER DEFAULT 0,
			status VARCHAR(20) DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// 验证码表
		`CREATE TABLE IF NOT EXISTS verification_codes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			target VARCHAR(100) NOT NULL,
			code VARCHAR(10) NOT NULL,
			type VARCHAR(20) NOT NULL,
			expires_at DATETIME NOT NULL,
			used BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// 商品变更日志
		`CREATE TABLE IF NOT EXISTS product_change_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			product_id INTEGER NOT NULL,
			change_type VARCHAR(20) NOT NULL,
			old_value VARCHAR(200),
			new_value VARCHAR(200),
			remark TEXT,
			order_no VARCHAR(50),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (product_id) REFERENCES products(id)
		)`,

		// 销售报表快照
		`CREATE TABLE IF NOT EXISTS sales_reports (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			report_date DATE NOT NULL,
			total_orders INTEGER DEFAULT 0,
			total_amount DECIMAL(12, 2) DEFAULT 0,
			total_products INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(report_date)
		)`,
	}

	for _, table := range tables {
		if _, err := DB.Exec(table); err != nil {
			return fmt.Errorf("create table failed: %v", err)
		}
	}

	// 初始化默认分类
	initDefaultCategories()

	return nil
}

func initDefaultCategories() {
	categories := []struct {
		Name string
		Code string
		Icon string
	}{
		{"手链", "bracelet", "bracelet.png"},
		{"项链", "necklace", "necklace.png"},
		{"配饰", "accessory", "accessory.png"},
		{"戒指", "ring", "ring.png"},
		{"耳饰", "earring", "earring.png"},
	}

	for i, cat := range categories {
		_, _ = DB.Exec(`INSERT OR IGNORE INTO categories (name, code, icon, sort_order) VALUES (?, ?, ?, ?)`,
			cat.Name, cat.Code, cat.Icon, i)
	}
}

func GetDB() *sql.DB {
	return DB
}
