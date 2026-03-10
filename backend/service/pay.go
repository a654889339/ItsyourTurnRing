package service

import (
	"crypto"
	"crypto/md5"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"itsyourturnring/config"
	"itsyourturnring/database"
)

type PayService struct {
	orderService *OrderService
}

func NewPayService() *PayService {
	return &PayService{orderService: NewOrderService()}
}

// PrepayResult 预支付结果
type PrepayResult struct {
	NeedPay   bool              `json:"need_pay"`
	PayMethod string            `json:"pay_method"`
	PayParams map[string]string `json:"pay_params,omitempty"`
}

// Prepay 发起预支付，根据支付方式和配置决定走真实支付还是模拟
func (s *PayService) Prepay(orderID, userID int64, payMethod string) (*PrepayResult, error) {
	cfg := config.GetConfig()

	switch payMethod {
	case "wechat":
		if cfg.WechatPay.Enabled {
			return s.wechatPrepay(orderID, userID)
		}
	case "alipay":
		if cfg.AlipayPay.Enabled {
			return s.alipayPrepay(orderID, userID)
		}
	default:
		return nil, errors.New("不支持的支付方式: " + payMethod)
	}

	// 未启用真实支付 -> 模拟支付
	if err := s.orderService.PayOrder(orderID, userID, payMethod); err != nil {
		return nil, err
	}
	return &PrepayResult{NeedPay: false, PayMethod: payMethod}, nil
}

// ==================== 微信支付 V2 JSAPI ====================

func (s *PayService) wechatPrepay(orderID, userID int64) (*PrepayResult, error) {
	cfg := config.GetConfig()
	db := database.GetDB()

	order, err := s.orderService.GetOrderByID(orderID, userID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.Status != "pending" {
		return nil, errors.New("订单状态不正确")
	}

	var openid string
	if err := db.QueryRow("SELECT COALESCE(wechat_openid,'') FROM users WHERE id = ?", userID).Scan(&openid); err != nil || openid == "" {
		return nil, errors.New("未绑定微信账号，无法发起微信支付")
	}

	totalFee := int(math.Round(order.PayPrice * 100))
	if totalFee <= 0 {
		return nil, errors.New("支付金额异常")
	}

	nonceStr := generateNonceStr(32)
	body := "RingShop订单-" + order.OrderNo

	params := map[string]string{
		"appid":            cfg.WechatMP.AppID,
		"mch_id":           cfg.WechatPay.MchID,
		"nonce_str":        nonceStr,
		"body":             body,
		"out_trade_no":     order.OrderNo,
		"total_fee":        strconv.Itoa(totalFee),
		"spbill_create_ip": "127.0.0.1",
		"notify_url":       cfg.WechatPay.NotifyURL,
		"trade_type":       "JSAPI",
		"openid":           openid,
		"sign_type":        "MD5",
	}
	params["sign"] = wxSign(params, cfg.WechatPay.APIKey)

	xmlBody := mapToXML(params)
	resp, err := httpClient.Post("https://api.mch.weixin.qq.com/pay/unifiedorder", "application/xml", strings.NewReader(xmlBody))
	if err != nil {
		return nil, fmt.Errorf("请求微信统一下单失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var wxResp struct {
		ReturnCode string `xml:"return_code"`
		ReturnMsg  string `xml:"return_msg"`
		ResultCode string `xml:"result_code"`
		ErrCode    string `xml:"err_code"`
		ErrCodeDes string `xml:"err_code_des"`
		PrepayID   string `xml:"prepay_id"`
	}
	if err := xml.Unmarshal(respBody, &wxResp); err != nil {
		return nil, fmt.Errorf("解析微信响应失败: %w", err)
	}
	if wxResp.ReturnCode != "SUCCESS" {
		return nil, fmt.Errorf("微信统一下单失败: %s", wxResp.ReturnMsg)
	}
	if wxResp.ResultCode != "SUCCESS" {
		return nil, fmt.Errorf("微信统一下单业务失败: %s %s", wxResp.ErrCode, wxResp.ErrCodeDes)
	}

	// 生成前端支付参数（二次签名）
	timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	payNonce := generateNonceStr(32)
	payParams := map[string]string{
		"appId":     cfg.WechatMP.AppID,
		"timeStamp": timeStamp,
		"nonceStr":  payNonce,
		"package":   "prepay_id=" + wxResp.PrepayID,
		"signType":  "MD5",
	}
	payParams["paySign"] = wxSign(payParams, cfg.WechatPay.APIKey)

	return &PrepayResult{
		NeedPay:   true,
		PayMethod: "wechat",
		PayParams: payParams,
	}, nil
}

// HandleWechatNotify 处理微信支付回调
func (s *PayService) HandleWechatNotify(r *http.Request) (string, error) {
	cfg := config.GetConfig()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return wxNotifyFail("读取请求失败"), err
	}

	var notify struct {
		ReturnCode    string `xml:"return_code"`
		ResultCode    string `xml:"result_code"`
		OutTradeNo    string `xml:"out_trade_no"`
		TransactionID string `xml:"transaction_id"`
		TotalFee      string `xml:"total_fee"`
		Sign          string `xml:"sign"`
		// 需要所有字段做验签
	}
	if err := xml.Unmarshal(body, &notify); err != nil {
		return wxNotifyFail("解析XML失败"), err
	}

	// 验签
	notifyMap := xmlToMap(string(body))
	receivedSign := notifyMap["sign"]
	delete(notifyMap, "sign")
	expectedSign := wxSign(notifyMap, cfg.WechatPay.APIKey)
	if receivedSign != expectedSign {
		return wxNotifyFail("签名验证失败"), errors.New("微信回调签名不一致")
	}

	if notify.ReturnCode != "SUCCESS" || notify.ResultCode != "SUCCESS" {
		return wxNotifyFail("支付失败"), errors.New("微信回调状态非成功")
	}

	// 查询订单并校验金额
	db := database.GetDB()
	var orderID, orderUserID int64
	var payPrice float64
	var status string
	err = db.QueryRow("SELECT id, user_id, pay_price, status FROM orders WHERE order_no = ?", notify.OutTradeNo).
		Scan(&orderID, &orderUserID, &payPrice, &status)
	if err != nil {
		return wxNotifyFail("订单不存在"), err
	}

	if status != "pending" {
		log.Printf("微信回调: 订单 %s 状态已是 %s, 跳过", notify.OutTradeNo, status)
		return wxNotifySuccess(), nil
	}

	expectedFee := int(math.Round(payPrice * 100))
	actualFee, _ := strconv.Atoi(notify.TotalFee)
	if expectedFee != actualFee {
		return wxNotifyFail("金额不一致"), fmt.Errorf("金额不一致: 期望 %d, 实际 %d", expectedFee, actualFee)
	}

	// 更新订单状态
	_, err = db.Exec(`UPDATE orders SET status = 'paid', pay_status = 'paid', pay_method = 'wechat',
		pay_time = CURRENT_TIMESTAMP, transaction_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'pending'`, notify.TransactionID, orderID)
	if err != nil {
		return wxNotifyFail("更新订单失败"), err
	}

	// 更新销量
	s.updateSales(orderID)

	log.Printf("微信支付成功: 订单 %s, 交易号 %s", notify.OutTradeNo, notify.TransactionID)
	return wxNotifySuccess(), nil
}

// ==================== 支付宝支付 ====================

func (s *PayService) alipayPrepay(orderID, userID int64) (*PrepayResult, error) {
	cfg := config.GetConfig()
	db := database.GetDB()

	order, err := s.orderService.GetOrderByID(orderID, userID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.Status != "pending" {
		return nil, errors.New("订单状态不正确")
	}

	var alipayUID string
	if err := db.QueryRow("SELECT COALESCE(alipay_userid,'') FROM users WHERE id = ?", userID).Scan(&alipayUID); err != nil || alipayUID == "" {
		return nil, errors.New("未绑定支付宝账号，无法发起支付宝支付")
	}

	privateKey, err := loadPrivateKeyFromFile(cfg.AlipayPay.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("加载支付宝私钥失败: %w", err)
	}

	bizContent := map[string]string{
		"out_trade_no": order.OrderNo,
		"total_amount": fmt.Sprintf("%.2f", order.PayPrice),
		"subject":      "RingShop订单-" + order.OrderNo,
		"buyer_id":     alipayUID,
	}
	bizContentJSON, _ := json.Marshal(bizContent)

	params := map[string]string{
		"app_id":      cfg.AlipayPay.AppID,
		"method":      "alipay.trade.create",
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  cfg.AlipayPay.NotifyURL,
		"biz_content": string(bizContentJSON),
	}

	signStr := buildAlipaySignString(params)
	sign, err := rsaSignSHA256(signStr, privateKey)
	if err != nil {
		return nil, fmt.Errorf("支付宝签名失败: %w", err)
	}
	params["sign"] = sign

	formValues := url.Values{}
	for k, v := range params {
		formValues.Set(k, v)
	}

	resp, err := httpClient.PostForm("https://openapi.alipay.com/gateway.do", formValues)
	if err != nil {
		return nil, fmt.Errorf("请求支付宝API失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Response struct {
			Code    string `json:"code"`
			Msg     string `json:"msg"`
			SubCode string `json:"sub_code"`
			SubMsg  string `json:"sub_msg"`
			TradeNo string `json:"trade_no"`
		} `json:"alipay_trade_create_response"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析支付宝响应失败: %w", err)
	}

	if result.Response.Code != "10000" {
		return nil, fmt.Errorf("支付宝创建交易失败: %s %s", result.Response.SubCode, result.Response.SubMsg)
	}

	return &PrepayResult{
		NeedPay:   true,
		PayMethod: "alipay",
		PayParams: map[string]string{
			"tradeNO": result.Response.TradeNo,
		},
	}, nil
}

// HandleAlipayNotify 处理支付宝支付回调
func (s *PayService) HandleAlipayNotify(r *http.Request) error {
	cfg := config.GetConfig()

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("解析支付宝回调参数失败: %w", err)
	}

	params := make(map[string]string)
	for k, v := range r.PostForm {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	// 验签
	sign := params["sign"]
	delete(params, "sign")
	delete(params, "sign_type")

	signStr := buildAlipaySignString(params)
	publicKey, err := loadPublicKeyFromFile(cfg.AlipayPay.AlipayPublicKeyPath)
	if err != nil {
		return fmt.Errorf("加载支付宝公钥失败: %w", err)
	}

	if err := rsaVerifySHA256(signStr, sign, publicKey); err != nil {
		return fmt.Errorf("支付宝回调验签失败: %w", err)
	}

	tradeStatus := params["trade_status"]
	if tradeStatus != "TRADE_SUCCESS" && tradeStatus != "TRADE_FINISHED" {
		log.Printf("支付宝回调: trade_status=%s, 非成功状态，跳过", tradeStatus)
		return nil
	}

	outTradeNo := params["out_trade_no"]
	tradeNo := params["trade_no"]
	totalAmount := params["total_amount"]

	db := database.GetDB()
	var orderID, orderUserID int64
	var payPrice float64
	var status string
	err = db.QueryRow("SELECT id, user_id, pay_price, status FROM orders WHERE order_no = ?", outTradeNo).
		Scan(&orderID, &orderUserID, &payPrice, &status)
	if err != nil {
		return fmt.Errorf("订单 %s 不存在", outTradeNo)
	}

	if status != "pending" {
		log.Printf("支付宝回调: 订单 %s 状态已是 %s, 跳过", outTradeNo, status)
		return nil
	}

	expectedAmount := fmt.Sprintf("%.2f", payPrice)
	if totalAmount != expectedAmount {
		return fmt.Errorf("金额不一致: 期望 %s, 实际 %s", expectedAmount, totalAmount)
	}

	_, err = db.Exec(`UPDATE orders SET status = 'paid', pay_status = 'paid', pay_method = 'alipay',
		pay_time = CURRENT_TIMESTAMP, transaction_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'pending'`, tradeNo, orderID)
	if err != nil {
		return fmt.Errorf("更新订单失败: %w", err)
	}

	s.updateSales(orderID)

	log.Printf("支付宝支付成功: 订单 %s, 交易号 %s", outTradeNo, tradeNo)
	return nil
}

// ==================== 辅助函数 ====================

func (s *PayService) updateSales(orderID int64) {
	db := database.GetDB()
	rows, err := db.Query("SELECT product_id, quantity FROM order_items WHERE order_id = ?", orderID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var pid int64
		var qty int
		if rows.Scan(&pid, &qty) == nil {
			db.Exec("UPDATE products SET sales = sales + ? WHERE id = ?", qty, pid)
		}
	}
}

func wxSign(params map[string]string, apiKey string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" || params[k] == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(params[k])
	}
	buf.WriteString("&key=")
	buf.WriteString(apiKey)

	hash := md5.Sum([]byte(buf.String()))
	return strings.ToUpper(fmt.Sprintf("%x", hash))
}

func mapToXML(params map[string]string) string {
	var buf strings.Builder
	buf.WriteString("<xml>")
	for k, v := range params {
		buf.WriteString(fmt.Sprintf("<%s><![CDATA[%s]]></%s>", k, v, k))
	}
	buf.WriteString("</xml>")
	return buf.String()
}

func xmlToMap(xmlStr string) map[string]string {
	result := make(map[string]string)
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	var currentKey string
	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := token.(type) {
		case xml.StartElement:
			currentKey = t.Name.Local
		case xml.CharData:
			if currentKey != "" && currentKey != "xml" {
				val := strings.TrimSpace(string(t))
				if val != "" {
					result[currentKey] = val
				}
			}
		case xml.EndElement:
			currentKey = ""
		}
	}
	return result
}

func wxNotifySuccess() string {
	return `<xml><return_code><![CDATA[SUCCESS]]></return_code><return_msg><![CDATA[OK]]></return_msg></xml>`
}

func wxNotifyFail(msg string) string {
	return fmt.Sprintf(`<xml><return_code><![CDATA[FAIL]]></return_code><return_msg><![CDATA[%s]]></return_msg></xml>`, msg)
}

func generateNonceStr(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[r.Intn(len(chars))]
	}
	return string(b)
}

func loadPrivateKeyFromFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func loadPublicKeyFromFile(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		raw := strings.ReplaceAll(string(data), "-----BEGIN PUBLIC KEY-----", "")
		raw = strings.ReplaceAll(raw, "-----END PUBLIC KEY-----", "")
		raw = strings.ReplaceAll(raw, "\n", "")
		raw = strings.ReplaceAll(raw, "\r", "")
		raw = strings.TrimSpace(raw)
		decoded, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return nil, fmt.Errorf("base64解码公钥失败: %w", err)
		}
		pub, err := x509.ParsePKIXPublicKey(decoded)
		if err != nil {
			return nil, err
		}
		rsaPub, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("非RSA公钥")
		}
		return rsaPub, nil
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("非RSA公钥")
	}
	return rsaPub, nil
}

func rsaVerifySHA256(content, signB64 string, pubKey *rsa.PublicKey) error {
	sigBytes, err := base64.StdEncoding.DecodeString(signB64)
	if err != nil {
		return err
	}
	h := sha256.Sum256([]byte(content))
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, h[:], sigBytes)
}
