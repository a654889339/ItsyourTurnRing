package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"itsyourturnring/config"
	"itsyourturnring/database"
	"itsyourturnring/model"
	"itsyourturnring/service"
)

var (
	authService     *service.AuthService
	productService  *service.ProductService
	orderService    *service.OrderService
	cartService     *service.CartService
	addressService  *service.AddressService
	uploadService   *service.UploadService
	reportService   *service.ReportService
	bannerService   *service.BannerService
	favoriteService *service.FavoriteService
	reviewService   *service.ReviewService
	qrcodeService   *service.QRCodeService
)

func main() {
	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// 初始化服务
	initServices()

	// 设置路由
	mux := http.NewServeMux()
	setupRoutes(mux)

	// 启动服务器
	cfg := config.GetConfig()
	addr := fmt.Sprintf(":%d", cfg.Server.BackendPort)
	log.Printf("Server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, corsMiddleware(mux)))
}

func initServices() {
	authService = service.NewAuthService()
	productService = service.NewProductService()
	orderService = service.NewOrderService()
	cartService = service.NewCartService()
	addressService = service.NewAddressService()
	uploadService = service.NewUploadService()
	reportService = service.NewReportService()
	bannerService = service.NewBannerService()
	favoriteService = service.NewFavoriteService()
	reviewService = service.NewReviewService()
	qrcodeService = service.NewQRCodeService()
}

func setupRoutes(mux *http.ServeMux) {
	// 健康检查
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/api/v1/health", handleHealth)

	// 认证路由 (无需登录)
	mux.HandleFunc("/api/v1/auth/register", handleRegister)
	mux.HandleFunc("/api/v1/auth/login", handleLogin)
	mux.HandleFunc("/api/v1/auth/me", authMiddleware(handleGetCurrentUser))

	// 公开API (无需登录)
	mux.HandleFunc("/api/v1/public/products", handlePublicProducts)
	mux.HandleFunc("/api/v1/public/products/", handlePublicProductDetail)
	mux.HandleFunc("/api/v1/public/categories", handlePublicCategories)
	mux.HandleFunc("/api/v1/public/banners", handlePublicBanners)
	mux.HandleFunc("/api/v1/public/home", handlePublicHome)

	// 商品管理 (需要登录)
	mux.HandleFunc("/api/v1/products", authMiddleware(handleProducts))
	mux.HandleFunc("/api/v1/products/", authMiddleware(handleProductByID))
	mux.HandleFunc("/api/v1/categories", authMiddleware(handleCategories))

	// 购物车
	mux.HandleFunc("/api/v1/cart", authMiddleware(handleCart))
	mux.HandleFunc("/api/v1/cart/", authMiddleware(handleCartItem))

	// 收货地址
	mux.HandleFunc("/api/v1/addresses", authMiddleware(handleAddresses))
	mux.HandleFunc("/api/v1/addresses/", authMiddleware(handleAddressByID))

	// 订单
	mux.HandleFunc("/api/v1/orders", authMiddleware(handleOrders))
	mux.HandleFunc("/api/v1/orders/admin", authMiddleware(handleAdminOrder))
	mux.HandleFunc("/api/v1/orders/", authMiddleware(handleOrderByID))

	// 收藏
	mux.HandleFunc("/api/v1/favorites", authMiddleware(handleFavorites))
	mux.HandleFunc("/api/v1/favorites/", authMiddleware(handleFavoriteByID))

	// 评价
	mux.HandleFunc("/api/v1/reviews", authMiddleware(handleReviews))
	mux.HandleFunc("/api/v1/reviews/product/", handleProductReviews)

	// 轮播图管理
	mux.HandleFunc("/api/v1/banners", authMiddleware(handleBanners))
	mux.HandleFunc("/api/v1/banners/", authMiddleware(handleBannerByID))

	// 报表
	mux.HandleFunc("/api/v1/reports/dashboard", authMiddleware(handleDashboard))
	mux.HandleFunc("/api/v1/reports/sales", authMiddleware(handleSalesReport))
	mux.HandleFunc("/api/v1/reports/products", authMiddleware(handleProductSalesRank))

	// 二维码管理
	mux.HandleFunc("/api/v1/qrcodes", authMiddleware(handleQRCodes))
	mux.HandleFunc("/api/v1/qrcodes/", authMiddleware(handleQRCodeByID))
	mux.HandleFunc("/api/v1/qrcodes/pages", authMiddleware(handleQRCodePages))

	// 上传
	mux.HandleFunc("/api/v1/upload/image", authMiddleware(handleUploadImage))
	mux.HandleFunc("/api/v1/upload/file", authMiddleware(handleUploadFile))

	// 静态文件
	fs := http.FileServer(http.Dir("./uploads"))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", fs))
}

// CORS中间件
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// 认证中间件
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			jsonError(w, "未登录", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			jsonError(w, "登录已过期", http.StatusUnauthorized)
			return
		}

		// 将用户ID放入请求上下文
		userID := int64((*claims)["user_id"].(float64))
		r.Header.Set("X-User-ID", strconv.FormatInt(userID, 10))

		next(w, r)
	}
}

func getUserID(r *http.Request) int64 {
	userID, _ := strconv.ParseInt(r.Header.Get("X-User-ID"), 10, 64)
	return userID
}

// JSON响应辅助函数
func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(model.APIResponse{
		Code:    code,
		Message: message,
	})
}

// ============ 处理器 ============

func handleHealth(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, map[string]string{"status": "ok"})
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result, err := authService.Register(&req)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonResponse(w, result)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result, err := authService.Login(&req)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonResponse(w, result)
}

func handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	user, err := authService.GetUserByID(userID)
	if err != nil {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}

	jsonResponse(w, user)
}

// 公开API处理器
func handlePublicProducts(w http.ResponseWriter, r *http.Request) {
	query := parsePageQuery(r)
	result, err := productService.ListPublicProducts(query)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result)
}

func handlePublicProductDetail(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/public/products/")
	productID, _ := strconv.ParseInt(idStr, 10, 64)

	product, err := productService.GetProductByID(productID)
	if err != nil {
		jsonError(w, "商品不存在", http.StatusNotFound)
		return
	}

	jsonResponse(w, product)
}

func handlePublicCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := productService.GetCategories()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, categories)
}

func handlePublicBanners(w http.ResponseWriter, r *http.Request) {
	banners, err := bannerService.GetActiveBanners()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, banners)
}

func handlePublicHome(w http.ResponseWriter, r *http.Request) {
	banners, _ := bannerService.GetActiveBanners()
	featured, _ := productService.GetFeaturedProducts(8)
	newProducts, _ := productService.GetNewProducts(8)
	categories, _ := productService.GetCategories()

	jsonResponse(w, map[string]interface{}{
		"banners":     banners,
		"featured":    featured,
		"new":         newProducts,
		"categories":  categories,
	})
}

// 商品管理处理器
func handleProducts(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	switch r.Method {
	case "GET":
		query := parsePageQuery(r)
		result, err := productService.ListProducts(query, userID)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, result)

	case "POST":
		var req model.ProductCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		product, err := productService.CreateProduct(userID, &req)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, product)

	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleProductByID(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/products/")

	// 处理子路由
	parts := strings.Split(idStr, "/")
	productID, _ := strconv.ParseInt(parts[0], 10, 64)

	if len(parts) > 1 && parts[1] == "change-logs" {
		logs, err := productService.GetProductChangeLogs(productID)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, logs)
		return
	}

	switch r.Method {
	case "GET":
		product, err := productService.GetProductByID(productID)
		if err != nil {
			jsonError(w, "商品不存在", http.StatusNotFound)
			return
		}
		jsonResponse(w, product)

	case "PUT":
		var req model.ProductCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		product, err := productService.UpdateProduct(productID, userID, &req)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, product)

	case "DELETE":
		if err := productService.DeleteProduct(productID, userID); err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, nil)

	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := productService.GetCategories()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, categories)
}

// 购物车处理器
func handleCart(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	switch r.Method {
	case "GET":
		items, err := cartService.ListCartItems(userID)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, items)

	case "POST":
		var req model.CartAddRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		item, err := cartService.AddToCart(userID, &req)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, item)

	case "DELETE":
		if err := cartService.ClearCart(userID); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, nil)

	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleCartItem(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/cart/")
	cartID, _ := strconv.ParseInt(idStr, 10, 64)

	switch r.Method {
	case "PUT":
		var req struct {
			Quantity int `json:"quantity"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if err := cartService.UpdateCartQuantity(cartID, userID, req.Quantity); err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, nil)

	case "DELETE":
		if err := cartService.RemoveFromCart(cartID, userID); err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, nil)

	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// 收货地址处理器
func handleAddresses(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	switch r.Method {
	case "GET":
		addresses, err := addressService.ListAddresses(userID)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, addresses)

	case "POST":
		var address model.Address
		if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		result, err := addressService.CreateAddress(userID, &address)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, result)

	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleAddressByID(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/addresses/")
	addressID, _ := strconv.ParseInt(idStr, 10, 64)

	switch r.Method {
	case "GET":
		address, err := addressService.GetAddressByID(addressID, userID)
		if err != nil {
			jsonError(w, "地址不存在", http.StatusNotFound)
			return
		}
		jsonResponse(w, address)

	case "PUT":
		var address model.Address
		if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		result, err := addressService.UpdateAddress(addressID, userID, &address)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, result)

	case "DELETE":
		if err := addressService.DeleteAddress(addressID, userID); err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, nil)

	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// 管理员快速下单
func handleAdminOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserID(r)
	var req model.AdminOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	order, err := orderService.CreateAdminOrder(userID, &req)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, order)
}

// 订单处理器
func handleOrders(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	switch r.Method {
	case "GET":
		query := parsePageQuery(r)
		result, err := orderService.ListOrders(query, userID)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, result)

	case "POST":
		var req model.OrderCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		order, err := orderService.CreateOrder(userID, &req)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, order)

	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleOrderByID(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/orders/"), "/")
	orderID, _ := strconv.ParseInt(pathParts[0], 10, 64)

	// 处理子路由
	if len(pathParts) > 1 {
		switch pathParts[1] {
		case "cancel":
			if r.Method == "POST" {
				if err := orderService.CancelOrder(orderID, userID); err != nil {
					jsonError(w, err.Error(), http.StatusBadRequest)
					return
				}
				jsonResponse(w, nil)
				return
			}
		case "pay":
			if r.Method == "POST" {
				var req struct {
					PayMethod string `json:"pay_method"`
				}
				json.NewDecoder(r.Body).Decode(&req)
				if err := orderService.PayOrder(orderID, userID, req.PayMethod); err != nil {
					jsonError(w, err.Error(), http.StatusBadRequest)
					return
				}
				jsonResponse(w, nil)
				return
			}
		case "receive":
			if r.Method == "POST" {
				if err := orderService.ConfirmReceive(orderID, userID); err != nil {
					jsonError(w, err.Error(), http.StatusBadRequest)
					return
				}
				jsonResponse(w, nil)
				return
			}
		case "status":
			if r.Method == "PUT" {
				var req model.OrderUpdateStatusRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					jsonError(w, "Invalid request", http.StatusBadRequest)
					return
				}
				if err := orderService.UpdateOrderStatus(orderID, userID, &req); err != nil {
					jsonError(w, err.Error(), http.StatusBadRequest)
					return
				}
				jsonResponse(w, nil)
				return
			}
		}
	}

	switch r.Method {
	case "GET":
		order, err := orderService.GetOrderByID(orderID, userID)
		if err != nil {
			jsonError(w, "订单不存在", http.StatusNotFound)
			return
		}
		jsonResponse(w, order)

	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// 收藏处理器
func handleFavorites(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	switch r.Method {
	case "GET":
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
		result, err := favoriteService.ListFavorites(userID, page, pageSize)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, result)

	case "POST":
		var req struct {
			ProductID int64 `json:"product_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if err := favoriteService.AddFavorite(userID, req.ProductID); err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, nil)

	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleFavoriteByID(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/favorites/")
	productID, _ := strconv.ParseInt(idStr, 10, 64)

	switch r.Method {
	case "DELETE":
		if err := favoriteService.RemoveFavorite(userID, productID); err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, nil)

	case "GET":
		isFav, _ := favoriteService.IsFavorite(userID, productID)
		jsonResponse(w, map[string]bool{"is_favorite": isFav})

	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// 评价处理器
func handleReviews(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	switch r.Method {
	case "GET":
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
		result, err := reviewService.ListUserReviews(userID, page, pageSize)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, result)

	case "POST":
		var review model.Review
		if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		result, err := reviewService.CreateReview(userID, &review)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, result)

	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleProductReviews(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/reviews/product/")
	productID, _ := strconv.ParseInt(idStr, 10, 64)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	result, err := reviewService.ListProductReviews(productID, page, pageSize)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result)
}

// 轮播图处理器
func handleBanners(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		banners, err := bannerService.ListBanners("")
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, banners)

	case "POST":
		var banner model.Banner
		if err := json.NewDecoder(r.Body).Decode(&banner); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		result, err := bannerService.CreateBanner(&banner)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, result)

	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleBannerByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/banners/")
	bannerID, _ := strconv.ParseInt(idStr, 10, 64)

	switch r.Method {
	case "GET":
		banner, err := bannerService.GetBannerByID(bannerID)
		if err != nil {
			jsonError(w, "轮播图不存在", http.StatusNotFound)
			return
		}
		jsonResponse(w, banner)

	case "PUT":
		var banner model.Banner
		if err := json.NewDecoder(r.Body).Decode(&banner); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		result, err := bannerService.UpdateBanner(bannerID, &banner)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, result)

	case "DELETE":
		if err := bannerService.DeleteBanner(bannerID); err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, nil)

	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// 报表处理器
func handleDashboard(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	stats, err := reportService.GetDashboardStats(userID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, stats)
}

func handleSalesReport(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	report, err := reportService.GetSalesReport(startDate, endDate)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, report)
}

func handleProductSalesRank(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	rank, err := reportService.GetProductSalesRank(startDate, endDate, limit)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, rank)
}

// 上传处理器
func handleUploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查是否是base64上传
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		var req struct {
			Image string `json:"image"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		url, err := uploadService.UploadBase64Image(req.Image)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, map[string]string{"url": url})
		return
	}

	// multipart上传
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
		jsonError(w, "文件太大", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		jsonError(w, "请选择文件", http.StatusBadRequest)
		return
	}
	defer file.Close()

	url, err := uploadService.UploadImage(file, header.Filename, header.Header.Get("Content-Type"))
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]string{"url": url})
}

// 辅助函数
func parsePageQuery(r *http.Request) *model.PageQuery {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	return &model.PageQuery{
		Page:     page,
		PageSize: pageSize,
		Keyword:  r.URL.Query().Get("keyword"),
		Category: r.URL.Query().Get("category"),
		Status:   r.URL.Query().Get("status"),
		SortBy:   r.URL.Query().Get("sort_by"),
		SortDir:  r.URL.Query().Get("sort_dir"),
	}
}

// ============ 二维码处理器 ============

func handleQRCodes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		platform := r.URL.Query().Get("platform")
		codes, err := qrcodeService.List(platform)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, codes)

	case "POST":
		var req model.QRCodeCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "无效的请求", http.StatusBadRequest)
			return
		}
		qr, err := qrcodeService.Create(&req)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, qr)

	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleQRCodeByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/qrcodes/")
	if path == "" || path == "pages" {
		return
	}
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		jsonError(w, "无效的ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		qr, err := qrcodeService.GetByID(id)
		if err != nil {
			jsonError(w, "二维码不存在", http.StatusNotFound)
			return
		}
		jsonResponse(w, qr)

	case "DELETE":
		if err := qrcodeService.Delete(id); err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, nil)

	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleQRCodePages(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := config.GetConfig()
	pages := []map[string]string{
		{"value": cfg.WechatMP.Pages.Home, "label": "首页"},
		{"value": cfg.WechatMP.Pages.Product, "label": "商品详情"},
		{"value": cfg.WechatMP.Pages.Cart, "label": "购物车"},
		{"value": cfg.WechatMP.Pages.Order, "label": "订单列表"},
		{"value": cfg.WechatMP.Pages.User, "label": "个人中心"},
	}
	jsonResponse(w, pages)
}

// 通用文件上传（图片+视频）
func handleUploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		jsonError(w, "文件太大，最大50MB", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		jsonError(w, "请选择文件", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileURL, fileType, err := uploadService.UploadFile(file, header.Filename, header.Header.Get("Content-Type"))
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonResponse(w, map[string]string{"url": fileURL, "type": fileType})
}
