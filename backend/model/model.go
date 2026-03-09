package model

import (
	"database/sql"
	"time"
)

// User 用户模型
type User struct {
	ID            int64          `json:"id"`
	Username      string         `json:"username"`
	Password      string         `json:"-"`
	Nickname      string         `json:"nickname"`
	Email         sql.NullString `json:"email"`
	Phone         sql.NullString `json:"phone"`
	Avatar        sql.NullString `json:"avatar"`
	Role          string         `json:"role"`
	WechatOpenID  string         `json:"wechat_openid,omitempty"`
	AlipayUserID  string         `json:"alipay_userid,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// Category 商品分类模型
type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Icon      string    `json:"icon"`
	SortOrder int       `json:"sort_order"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Product 商品模型 (手链、项链、配饰等)
type Product struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	CategoryID    int64     `json:"category_id"`
	CategoryName  string    `json:"category_name,omitempty"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Price         float64   `json:"price"`
	OriginalPrice float64   `json:"original_price"`
	Images        string    `json:"images"`          // JSON数组
	MainImage     string    `json:"main_image"`
	Video         string    `json:"video"`           // 视频URL
	Material      string    `json:"material"`        // 材质
	Size          string    `json:"size"`            // 尺寸
	Color         string    `json:"color"`           // 颜色
	Stock         int       `json:"stock"`
	Sales         int       `json:"sales"`
	Status        string    `json:"status"`          // available/sold_out/disabled
	IsFeatured    bool      `json:"is_featured"`     // 推荐
	IsNew         bool      `json:"is_new"`          // 新品
	SortOrder     int       `json:"sort_order"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Specs         []ProductSpec `json:"specs,omitempty"`
}

// ProductSpec 商品规格
type ProductSpec struct {
	ID              int64     `json:"id"`
	ProductID       int64     `json:"product_id"`
	Name            string    `json:"name"`
	Value           string    `json:"value"`
	PriceAdjustment float64   `json:"price_adjustment"`
	Stock           int       `json:"stock"`
	CreatedAt       time.Time `json:"created_at"`
}

// CartItem 购物车项
type CartItem struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	ProductID    int64     `json:"product_id"`
	SpecID       *int64    `json:"spec_id"`
	Quantity     int       `json:"quantity"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Product      *Product  `json:"product,omitempty"`
	Spec         *ProductSpec `json:"spec,omitempty"`
}

// Address 收货地址
type Address struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Province  string    `json:"province"`
	City      string    `json:"city"`
	District  string    `json:"district"`
	Detail    string    `json:"detail"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Order 订单模型
type Order struct {
	ID              int64       `json:"id"`
	UserID          int64       `json:"user_id"`
	OrderNo         string      `json:"order_no"`
	TotalPrice      float64     `json:"total_price"`
	PayPrice        float64     `json:"pay_price"`
	Freight         float64     `json:"freight"`
	Status          string      `json:"status"`       // pending/paid/shipped/received/completed/cancelled
	PayStatus       string      `json:"pay_status"`   // unpaid/paid/refunded
	PayMethod       string      `json:"pay_method"`   // wechat/alipay
	PayTime         *time.Time  `json:"pay_time"`
	ShipTime        *time.Time  `json:"ship_time"`
	ReceiveTime     *time.Time  `json:"receive_time"`
	AddressName     string      `json:"address_name"`
	AddressPhone    string      `json:"address_phone"`
	AddressProvince string      `json:"address_province"`
	AddressCity     string      `json:"address_city"`
	AddressDistrict string      `json:"address_district"`
	AddressDetail   string      `json:"address_detail"`
	ExpressCompany  string      `json:"express_company"`
	ExpressNo       string      `json:"express_no"`
	Remark          string      `json:"remark"`
	OrderSource     string      `json:"order_source"` // web/wechat_mp/alipay_mp
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	Items           []OrderItem `json:"items,omitempty"`
}

// OrderItem 订单项
type OrderItem struct {
	ID           int64     `json:"id"`
	OrderID      int64     `json:"order_id"`
	ProductID    int64     `json:"product_id"`
	SpecID       *int64    `json:"spec_id"`
	ProductName  string    `json:"product_name"`
	ProductImage string    `json:"product_image"`
	SpecName     string    `json:"spec_name"`
	Price        float64   `json:"price"`
	Quantity     int       `json:"quantity"`
	CreatedAt    time.Time `json:"created_at"`
}

// Favorite 收藏
type Favorite struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	ProductID int64     `json:"product_id"`
	CreatedAt time.Time `json:"created_at"`
	Product   *Product  `json:"product,omitempty"`
}

// Review 评价
type Review struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	ProductID   int64     `json:"product_id"`
	OrderID     int64     `json:"order_id"`
	Rating      int       `json:"rating"`
	Content     string    `json:"content"`
	Images      string    `json:"images"` // JSON数组
	IsAnonymous bool      `json:"is_anonymous"`
	CreatedAt   time.Time `json:"created_at"`
	Username    string    `json:"username,omitempty"`
	Avatar      string    `json:"avatar,omitempty"`
}

// Banner 轮播图
type Banner struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Image     string    `json:"image"`
	Link      string    `json:"link"`
	SortOrder int       `json:"sort_order"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VerificationCode 验证码
type VerificationCode struct {
	ID        int64     `json:"id"`
	Target    string    `json:"target"` // 邮箱或手机号
	Code      string    `json:"code"`
	Type      string    `json:"type"`   // register/login/reset
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

// QRCode 小程序二维码
type QRCode struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Scene     string    `json:"scene"`     // product_view / product_buy / order_status / home / custom
	Platform  string    `json:"platform"`  // wechat / alipay
	Page      string    `json:"page"`
	Params    string    `json:"params"`
	ImageURL  string    `json:"image_url"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// QRCodeCreateRequest 创建二维码请求
type QRCodeCreateRequest struct {
	Name      string `json:"name"`
	Scene     string `json:"scene"`
	Platform  string `json:"platform"`
	Page      string `json:"page"`
	Params    string `json:"params"`
	ProductID int64  `json:"product_id"`
	OrderNo   string `json:"order_no"`
}

// ProductChangeLog 商品变更日志
type ProductChangeLog struct {
	ID         int64     `json:"id"`
	ProductID  int64     `json:"product_id"`
	ChangeType string    `json:"change_type"` // stock/price/status
	OldValue   string    `json:"old_value"`
	NewValue   string    `json:"new_value"`
	Remark     string    `json:"remark"`
	OrderNo    string    `json:"order_no"`
	CreatedAt  time.Time `json:"created_at"`
}

// OrderChangeLog 订单变更日志
type OrderChangeLog struct {
	ID         int64     `json:"id"`
	OrderID    int64     `json:"order_id"`
	ChangeType string    `json:"change_type"` // status/price/remark
	OldValue   string    `json:"old_value"`
	NewValue   string    `json:"new_value"`
	Operator   string    `json:"operator"`
	CreatedAt  time.Time `json:"created_at"`
}

// SalesReport 销售报表
type SalesReport struct {
	ID            int64     `json:"id"`
	ReportDate    time.Time `json:"report_date"`
	TotalOrders   int       `json:"total_orders"`
	TotalAmount   float64   `json:"total_amount"`
	TotalProducts int       `json:"total_products"`
	CreatedAt     time.Time `json:"created_at"`
}

// API请求/响应结构

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

// ProductCreateRequest 创建商品请求
type ProductCreateRequest struct {
	CategoryID    int64   `json:"category_id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Price         float64 `json:"price"`
	OriginalPrice float64 `json:"original_price"`
	Images        string  `json:"images"`
	MainImage     string  `json:"main_image"`
	Video         string  `json:"video"`
	Material      string  `json:"material"`
	Size          string  `json:"size"`
	Color         string  `json:"color"`
	Stock         int     `json:"stock"`
	IsFeatured    bool    `json:"is_featured"`
	IsNew         bool    `json:"is_new"`
}

// CartAddRequest 添加购物车请求
type CartAddRequest struct {
	ProductID int64  `json:"product_id"`
	SpecID    *int64 `json:"spec_id"`
	Quantity  int    `json:"quantity"`
}

// OrderCreateRequest 创建订单请求
type OrderCreateRequest struct {
	AddressID int64  `json:"address_id"`
	CartIDs   []int64 `json:"cart_ids"`
	Remark    string `json:"remark"`
}

// AdminOrderItem 管理员下单商品项
type AdminOrderItem struct {
	ProductID int64  `json:"product_id"`
	SpecID    *int64 `json:"spec_id"`
	Quantity  int    `json:"quantity"`
}

// AdminOrderRequest 管理员快速下单请求
type AdminOrderRequest struct {
	Items           []AdminOrderItem `json:"items"`
	AddressName     string           `json:"address_name"`
	AddressPhone    string           `json:"address_phone"`
	AddressProvince string           `json:"address_province"`
	AddressCity     string           `json:"address_city"`
	AddressDistrict string           `json:"address_district"`
	AddressDetail   string           `json:"address_detail"`
	Remark          string           `json:"remark"`
}

// OrderUpdateStatusRequest 更新订单状态请求
type OrderUpdateStatusRequest struct {
	Status         string `json:"status"`
	ExpressCompany string `json:"express_company"`
	ExpressNo      string `json:"express_no"`
}

// PageQuery 分页查询
type PageQuery struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Keyword  string `json:"keyword"`
	Category string `json:"category"`
	Status   string `json:"status"`
	SortBy   string `json:"sort_by"`
	SortDir  string `json:"sort_dir"`
}

// PageResult 分页结果
type PageResult struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Data     interface{} `json:"data"`
}

// APIResponse 通用API响应
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// TokenResponse 登录成功响应
type TokenResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// DashboardStats 仪表盘统计
type DashboardStats struct {
	TodayOrders     int     `json:"today_orders"`
	TodayAmount     float64 `json:"today_amount"`
	TotalProducts   int     `json:"total_products"`
	TotalUsers      int     `json:"total_users"`
	PendingOrders   int     `json:"pending_orders"`
	LowStockProducts int    `json:"low_stock_products"`
}

// SalesReportData 销售报表数据
type SalesReportData struct {
	Date   string  `json:"date"`
	Orders int     `json:"orders"`
	Amount float64 `json:"amount"`
}

// ProductSalesRank 商品销量排行
type ProductSalesRank struct {
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	Sales       int     `json:"sales"`
	Amount      float64 `json:"amount"`
}
