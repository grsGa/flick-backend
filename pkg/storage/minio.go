package storage

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
)

// MinioClient 包装了Minio客户端
type MinioClient struct {
	client         *minio.Client
	bucketName     string
	endpoint       string
	publicEndpoint string // 用于前端访问的公共端点
	useSSL         bool
}

// MinioConfig 配置Minio客户端
type MinioConfig struct {
	Endpoint       string
	AccessKey      string
	SecretKey      string
	BucketName     string
	UseSSL         bool
	PublicEndpoint string // 用于前端访问的公共端点
}

// NewMinioClient 创建一个新的Minio客户端
func NewMinioClient(config MinioConfig) (*MinioClient, error) {
	// 创建MinIO客户端
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("无法创建MinIO客户端: %w", err)
	}

	// 检查存储桶是否存在，不存在则创建
	exists, err := client.BucketExists(context.Background(), config.BucketName)
	if err != nil {
		return nil, fmt.Errorf("检查存储桶失败: %w", err)
	}

	if !exists {
		err = client.MakeBucket(context.Background(), config.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("创建存储桶失败: %w", err)
		}
		log.Info().Msgf("已创建存储桶: %s", config.BucketName)

		// 设置存储桶策略，允许公共读取
		policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::` + config.BucketName + `/*"]}]}`
		err = client.SetBucketPolicy(context.Background(), config.BucketName, policy)
		if err != nil {
			log.Warn().Err(err).Msgf("设置存储桶策略失败: %s", config.BucketName)
		}
	}

	return &MinioClient{
		client:         client,
		bucketName:     config.BucketName,
		endpoint:       config.Endpoint,
		publicEndpoint: config.PublicEndpoint,
		useSSL:         config.UseSSL,
	}, nil
}

// 生成文件的唯一存储路径
func generateObjectName(userID, fileType string, filename string) string {
	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".jpg" // 默认扩展名
	}

	// 生成更具唯一性的文件名，包含时间戳、随机字符串和用户ID的部分哈希
	timestamp := time.Now().UnixNano()
	randomString := make([]byte, 8)
	for i := range randomString {
		randomString[i] = "abcdefghijklmnopqrstuvwxyz0123456789"[rand.Intn(36)]
	}

	// 计算用户ID的简单哈希
	userHash := 0
	for _, c := range userID {
		userHash = userHash*31 + int(c)
	}
	userHash = userHash & 0xFFFFFF // 保留低24位

	// 组合文件名部分
	objectName := fmt.Sprintf("%s/%s/%d_%s_%x%s",
		userID,
		fileType,
		timestamp,
		randomString,
		userHash,
		ext)

	log.Debug().
		Str("user_id", userID).
		Str("file_type", fileType).
		Str("original_filename", filename).
		Str("object_name", objectName).
		Msg("生成对象存储路径")

	return objectName
}

// UploadFile 上传文件到MinIO
func (m *MinioClient) UploadFile(ctx context.Context, userID, fileType string, fileSize int64, fileReader io.Reader, originalFileName string) (string, error) {
	// 生成唯一的对象名称
	objectName := generateObjectName(userID, fileType, originalFileName)

	// 获取文件内容类型
	contentType := ""
	if seeker, ok := fileReader.(io.ReadSeeker); ok {
		buffer := make([]byte, 512)
		_, err := seeker.Read(buffer)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("无法读取文件头: %w", err)
		}
		contentType = http.DetectContentType(buffer)
		_, err = seeker.Seek(0, io.SeekStart)
		if err != nil {
			return "", fmt.Errorf("无法重置文件读取位置: %w", err)
		}
	}

	// 如果不能确定内容类型，设置为空默认值让MinIO自动检测
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// 上传文件到MinIO
	_, err := m.client.PutObject(ctx, m.bucketName, objectName, fileReader, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	// 使用公共URL而不是预签名URL，这样图片可以永久访问
	return m.GetPublicURL(objectName), nil
}

// UploadFileWithObjectName 上传文件到MinIO并返回公共URL和对象名称
func (m *MinioClient) UploadFileWithObjectName(ctx context.Context, userID, fileType string, fileSize int64, fileReader io.Reader, originalFileName string) (string, string, error) {
	// 生成唯一的对象名称
	objectName := generateObjectName(userID, fileType, originalFileName)

	// 获取文件内容类型
	contentType := ""
	if seeker, ok := fileReader.(io.ReadSeeker); ok {
		buffer := make([]byte, 512)
		_, err := seeker.Read(buffer)
		if err != nil && err != io.EOF {
			return "", "", fmt.Errorf("无法读取文件头: %w", err)
		}
		contentType = http.DetectContentType(buffer)
		_, err = seeker.Seek(0, io.SeekStart)
		if err != nil {
			return "", "", fmt.Errorf("无法重置文件读取位置: %w", err)
		}
	}

	// 如果不能确定内容类型，设置为空默认值让MinIO自动检测
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// 上传文件到MinIO
	_, err := m.client.PutObject(ctx, m.bucketName, objectName, fileReader, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", "", err
	}

	// 使用公共URL而不是预签名URL，这样图片可以永久访问
	return m.GetPublicURL(objectName), objectName, nil
}

// GetFileURL 生成带过期时间的文件预签名URL
func (m *MinioClient) GetFileURL(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucketName, objectName, expires, nil)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

// GetPublicURL 生成永久有效的公共URL
func (m *MinioClient) GetPublicURL(objectName string) string {
	// 使用公共端点生成URL
	endpoint := m.endpoint
	if m.publicEndpoint != "" {
		endpoint = m.publicEndpoint
	}

	// 构建公共URL格式: http://localhost:9000/flick-bucket/objectName
	protocol := "http"
	if m.useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, endpoint, m.bucketName, objectName)
}

// DeleteFile 从MinIO删除文件
func (m *MinioClient) DeleteFile(ctx context.Context, objectName string) error {
	err := m.client.RemoveObject(ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	log.Info().Msgf("已删除文件: %s", objectName)
	return nil
}

// ExtractObjectNameFromURL 从MinIO生成的URL中提取对象名称
func ExtractObjectNameFromURL(url string, bucketName string) string {
	// 先移除查询参数
	urlPath := url
	if queryIdx := strings.Index(url, "?"); queryIdx != -1 {
		urlPath = url[:queryIdx]
	}

	// 基本URL格式通常是: http(s)://endpoint/bucketName/objectName
	parts := strings.Split(urlPath, "/"+bucketName+"/")
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimPrefix(parts[1], "/")
}
