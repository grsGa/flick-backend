package service

import (
	"context"
	"github.com/flick/backend/services/media/proto"
)

// MediaService 定义媒体服务接口
type MediaService interface {
	// UploadFile 上传文件
	UploadFile(ctx context.Context, req *proto.UploadFileRequest) (*proto.UploadFileResponse, error)
	
	// GetFile 获取文件信息
	GetFile(ctx context.Context, req *proto.GetFileRequest) (*proto.GetFileResponse, error)
	
	// DeleteFile 删除文件
	DeleteFile(ctx context.Context, req *proto.DeleteFileRequest) (*proto.DeleteFileResponse, error)
	
	// ListFiles 获取文件列表
	ListFiles(ctx context.Context, req *proto.ListFilesRequest) (*proto.ListFilesResponse, error)
}
