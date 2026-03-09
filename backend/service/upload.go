package service

import (
	"bytes"
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

var imageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

var videoExts = map[string]bool{
	".mp4": true, ".mov": true, ".avi": true, ".webm": true, ".mkv": true,
}

func (s *UploadService) UploadImage(file io.Reader, filename string, contentType string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = s.getExtFromContentType(contentType)
	}
	if !imageExts[ext] {
		return "", fmt.Errorf("不支持的图片格式，仅支持 jpg/png/gif/webp")
	}
	return s.upload(file, "ring/images/", ext)
}

func (s *UploadService) UploadVideo(file io.Reader, filename string, contentType string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = s.getExtFromVideoContentType(contentType)
	}
	if !videoExts[ext] {
		return "", fmt.Errorf("不支持的视频格式，仅支持 mp4/mov/avi/webm/mkv")
	}
	return s.upload(file, "ring/videos/", ext)
}

// UploadFile 通用上传，自动判断图片或视频
func (s *UploadService) UploadFile(file io.Reader, filename string, contentType string) (string, string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = s.getExtFromContentType(contentType)
	}

	if imageExts[ext] {
		u, err := s.upload(file, "ring/images/", ext)
		return u, "image", err
	}
	if videoExts[ext] {
		u, err := s.upload(file, "ring/videos/", ext)
		return u, "video", err
	}
	return "", "", fmt.Errorf("不支持的文件格式: %s", ext)
}

func (s *UploadService) upload(file io.Reader, prefix string, ext string) (string, error) {
	cfg := config.GetConfig()
	newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	if cfg.TencentCloud.COS.Enabled {
		return s.uploadToCOS(file, prefix+newFilename)
	}
	return s.uploadToLocal(file, newFilename)
}

func (s *UploadService) UploadBase64Image(base64Data string) (string, error) {
	parts := strings.SplitN(base64Data, ",", 2)
	var data string
	var ext string

	if len(parts) == 2 {
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

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("invalid base64 data: %v", err)
	}

	cfg := config.GetConfig()
	newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	if cfg.TencentCloud.COS.Enabled {
		return s.uploadToCOS(bytes.NewReader(decoded), "ring/images/"+newFilename)
	}
	return s.uploadToLocalBytes(decoded, newFilename)
}

func (s *UploadService) uploadToLocal(file io.Reader, filename string) (string, error) {
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}
	filePath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()
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

func (s *UploadService) uploadToCOS(file io.Reader, key string) (string, error) {
	cfg := config.GetConfig()
	u, _ := url.Parse(cfg.TencentCloud.COS.BaseURL)
	client := cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.TencentCloud.COS.SecretID,
			SecretKey: cfg.TencentCloud.COS.SecretKey,
		},
	})

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

func (s *UploadService) getExtFromVideoContentType(contentType string) string {
	switch contentType {
	case "video/mp4":
		return ".mp4"
	case "video/quicktime":
		return ".mov"
	case "video/x-msvideo":
		return ".avi"
	case "video/webm":
		return ".webm"
	default:
		return ".mp4"
	}
}

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
	client := cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.TencentCloud.COS.SecretID,
			SecretKey: cfg.TencentCloud.COS.SecretKey,
		},
	})
	key := strings.TrimPrefix(imageURL, cfg.TencentCloud.COS.BaseURL+"/")
	_, err := client.Object.Delete(context.Background(), key)
	return err
}
