package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
	"itsyourturnring/config"
)

type UploadService struct{}

func NewUploadService() *UploadService {
	return &UploadService{}
}

// UploadImage 上传图片
func (s *UploadService) UploadImage(file io.Reader, filename string, contentType string) (string, error) {
	cfg := config.GetConfig()

	// 生成新文件名
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = s.getExtFromContentType(contentType)
	}
	newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	if cfg.TencentCloud.COS.Enabled {
		return s.uploadToCOS(file, newFilename)
	}

	return s.uploadToLocal(file, newFilename)
}

// UploadBase64Image 上传Base64图片
func (s *UploadService) UploadBase64Image(base64Data string) (string, error) {
	// 解析Base64数据
	parts := strings.SplitN(base64Data, ",", 2)
	var data string
	var ext string

	if len(parts) == 2 {
		// 格式: data:image/jpeg;base64,xxxxx
		meta := parts[0]
		data = parts[1]
		if strings.Contains(meta, "image/jpeg") {
			ext = ".jpg"
		} else if strings.Contains(meta, "image/png") {
			ext = ".png"
		} else if strings.Contains(meta, "image/gif") {
			ext = ".gif"
		} else if strings.Contains(meta, "image/webp") {
			ext = ".webp"
		} else {
			ext = ".jpg"
		}
	} else {
		data = base64Data
		ext = ".jpg"
	}

	// 解码
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("invalid base64 data: %v", err)
	}

	// 生成文件名
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	cfg := config.GetConfig()
	if cfg.TencentCloud.COS.Enabled {
		return s.uploadToCOS(strings.NewReader(string(decoded)), filename)
	}

	return s.uploadToLocalBytes(decoded, filename)
}

func (s *UploadService) uploadToLocal(file io.Reader, filename string) (string, error) {
	// 确保目录存在
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	// 创建文件
	filePath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// 写入文件
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return "/uploads/" + filename, nil
}

func (s *UploadService) uploadToLocalBytes(data []byte, filename string) (string, error) {
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	filePath := filepath.Join(uploadDir, filename)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", err
	}

	return "/uploads/" + filename, nil
}

func (s *UploadService) uploadToCOS(file io.Reader, filename string) (string, error) {
	cfg := config.GetConfig()

	// 创建COS客户端
	u, _ := url.Parse(cfg.TencentCloud.COS.BaseURL)
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.TencentCloud.COS.SecretID,
			SecretKey: cfg.TencentCloud.COS.SecretKey,
		},
	})

	// 上传文件
	key := "ring/" + filename
	_, err := client.Object.Put(context.Background(), key, file, nil)
	if err != nil {
		return "", err
	}

	return cfg.TencentCloud.COS.BaseURL + "/" + key, nil
}

func (s *UploadService) getExtFromContentType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".jpg"
	}
}

// DeleteImage 删除图片
func (s *UploadService) DeleteImage(imageURL string) error {
	cfg := config.GetConfig()

	if cfg.TencentCloud.COS.Enabled && strings.HasPrefix(imageURL, cfg.TencentCloud.COS.BaseURL) {
		return s.deleteFromCOS(imageURL)
	}

	if strings.HasPrefix(imageURL, "/uploads/") {
		return s.deleteFromLocal(imageURL)
	}

	return nil
}

func (s *UploadService) deleteFromLocal(imageURL string) error {
	filePath := "." + imageURL
	return os.Remove(filePath)
}

func (s *UploadService) deleteFromCOS(imageURL string) error {
	cfg := config.GetConfig()

	u, _ := url.Parse(cfg.TencentCloud.COS.BaseURL)
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.TencentCloud.COS.SecretID,
			SecretKey: cfg.TencentCloud.COS.SecretKey,
		},
	})

	key := strings.TrimPrefix(imageURL, cfg.TencentCloud.COS.BaseURL+"/")
	_, err := client.Object.Delete(context.Background(), key)
	return err
}
