package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"go-wind-admin/app/admin/service/internal/data"
	"go-wind-admin/pkg/middleware/auth"
	"io"
	"math"
	"net/http"
	"path"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	adminV1 "go-wind-admin/api/gen/go/admin/service/v1"
	fileV1 "go-wind-admin/api/gen/go/file/service/v1"

	"go-wind-admin/pkg/oss"
)

type FileTransferService struct {
	adminV1.FileTransferServiceHTTPServer

	log *log.Helper

	mc       *oss.MinIOClient
	fileRepo *data.FileRepo
}

func NewFileTransferService(
	ctx *bootstrap.Context,
	mc *oss.MinIOClient,
	fileRepo *data.FileRepo,
) *FileTransferService {
	return &FileTransferService{
		log:      ctx.NewLoggerHelper("file-transfer/service/admin-service"),
		mc:       mc,
		fileRepo: fileRepo,
	}
}

func parseKey(key string) (folder, filename, ext string) {
	if key == "" {
		return "", "", ""
	}

	// 统一去除前导斜杠，但保留中间路径
	key = strings.TrimPrefix(key, "/")

	// 如果以 '/' 结尾，则视为目录
	if strings.HasSuffix(key, "/") {
		f := strings.TrimSuffix(key, "/")
		return f, "", ""
	}

	// 目录部分
	dir := path.Dir(key)
	if dir == "." {
		dir = ""
	}

	base := path.Base(key)

	// 处理点文件（如 .env）：当且仅当只有一个前导点且没有其他点，视为无扩展名
	if strings.HasPrefix(base, ".") && strings.Count(base, ".") == 1 {
		return dir, base, ""
	}

	// 查找最后一个点作为扩展名分隔（点在开头不算）
	idx := strings.LastIndex(base, ".")
	if idx <= 0 {
		// 无扩展名或点在首位（已处理首位点情况）
		return dir, base, ""
	}

	name := base[:idx]
	ext = base[idx:] // 包含前导 '.'

	if dir == "" {
		dir = "/"
	}

	return dir, name, ext
}

// formatSize 返回归一化后的数值和单位。
// 例如：formatSize(1536) -> 1.5, "KB"；formatSize(512) -> 512, "B"
func formatSize(size int64) (float64, string) {
	if size <= 0 {
		return 0, "B"
	}

	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	s := float64(size)
	i := 0
	for s >= 1024 && i < len(units)-1 {
		s /= 1024
		i++
	}

	if i == 0 {
		// 字节单位返回整数
		return math.Round(s), units[i]
	}
	// 非字节单位保留最多两位小数
	v := math.Round(s*100) / 100
	return v, units[i]
}

// recordFile 记录文件元数据到数据库
func (s *FileTransferService) recordFile(
	ctx context.Context,
	tenantID, userID uint32,
	fileData []byte,
	info minio.UploadInfo,
	downloadUrl string,
) error {

	sum := md5.Sum(fileData)             // md5.Sum 返回 [16]byte
	md5Hex := hex.EncodeToString(sum[:]) // 转为十六进制字符串

	dir, name, ext := parseKey(info.Key)

	sizeNum, sizeFormat := formatSize(info.Size)

	if err := s.fileRepo.Create(ctx, &fileV1.CreateFileRequest{
		Data: &fileV1.File{
			Provider:      trans.Ptr(fileV1.OSSProvider_MINIO),
			BucketName:    trans.Ptr(info.Bucket),
			SaveFileName:  trans.Ptr(info.Key),
			Md5:           trans.Ptr(md5Hex),
			FileDirectory: trans.Ptr(dir),
			FileName:      trans.Ptr(name),
			Extension:     trans.Ptr(ext),
			FileGuid:      trans.Ptr(uuid.New().String()),
			Size:          trans.Ptr(uint64(sizeNum)),
			SizeFormat:    trans.Ptr(sizeFormat),
			LinkUrl:       trans.Ptr(downloadUrl),
			CreatedBy:     trans.Ptr(userID),
			TenantId:      trans.Ptr(tenantID),
		},
	}); err != nil {
		s.log.Errorf("Failed to create file record: %v", err)
		return err
	}
	return nil
}

func (s *FileTransferService) UploadFile(ctx context.Context, req *fileV1.UploadFileRequest) (*fileV1.UploadFileResponse, error) {
	if req.File == nil {
		return nil, fileV1.ErrorUploadFailed("unknown fileData")
	}

	// 获取操作人信息
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.BucketName == nil {
		req.BucketName = trans.Ptr(s.mc.ContentTypeToBucketName(req.GetMime()))
	}
	if req.ObjectName == nil {
		req.ObjectName = trans.Ptr(req.GetSourceFileName())
	}

	info, downloadUrl, err := s.mc.UploadFile(ctx, req.GetBucketName(), req.GetObjectName(), req.GetFile())
	if err != nil {
		return nil, err
	}

	if err = s.recordFile(
		ctx,
		operator.GetTenantId(), operator.GetUserId(),
		req.GetFile(),
		info, downloadUrl); err != nil {
	}

	return &fileV1.UploadFileResponse{
		Url: downloadUrl,
	}, err
}

// downloadFileFromURL 从指定的 URL 下载文件内容
func (s *FileTransferService) downloadFileFromURL(ctx context.Context, downloadUrl string) (*fileV1.DownloadFileResponse, error) {
	if downloadUrl == "" {
		return nil, fileV1.ErrorDownloadFailed("empty download url")
	}

	// 如果需要支持断点续传，可在此构造请求并设置 Range 头
	httpReq, err := http.NewRequestWithContext(ctx, "GET", downloadUrl, nil)
	if err != nil {
		return nil, fileV1.ErrorDownloadFailed(err.Error())
	}
	// 示例：如果你要设置 Range（可选）
	// httpReq.Header.Set("Range", "bytes=100-")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fileV1.ErrorDownloadFailed(err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return nil, fileV1.ErrorDownloadFailed("unexpected status: " + resp.Status)
	}

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fileV1.ErrorDownloadFailed(err.Error())
	}

	return &fileV1.DownloadFileResponse{
		Content: &fileV1.DownloadFileResponse_File{
			File: fileData,
		},
	}, nil
}

// DownloadFile 下载文件
func (s *FileTransferService) DownloadFile(ctx context.Context, req *fileV1.DownloadFileRequest) (*fileV1.DownloadFileResponse, error) {
	switch req.Selector.(type) {
	case *fileV1.DownloadFileRequest_FileId:
		return nil, fileV1.ErrorDownloadFailed("unsupported file ID download")

	case *fileV1.DownloadFileRequest_StorageObject:
		return s.mc.DownloadFile(ctx, req)

	case *fileV1.DownloadFileRequest_DownloadUrl:
		return s.downloadFileFromURL(ctx, req.GetDownloadUrl())

	default:
		return nil, fileV1.ErrorDownloadFailed("unknown download selector")
	}
}

func (s *FileTransferService) UEditorUploadFile(ctx context.Context, req *fileV1.UEditorUploadRequest) (*fileV1.UEditorUploadResponse, error) {
	//s.log.Infof("上传文件： %s", req.GetFile())

	if req.File == nil {
		return nil, fileV1.ErrorUploadFailed("unknown file")
	}

	var bucketName string
	switch req.GetAction() {
	default:
		fallthrough
	case fileV1.UEditorAction_uploadFile.String():
		bucketName = "files"
	case fileV1.UEditorAction_uploadImage.String(), fileV1.UEditorAction_uploadScrawl.String(), fileV1.UEditorAction_catchImage.String():
		bucketName = "images"
	case fileV1.UEditorAction_uploadVideo.String():
		bucketName = "videos"
	}

	_, downloadUrl, err := s.mc.UploadFile(ctx, bucketName, req.GetSourceFileName(), req.GetFile())
	if err != nil {
		return &fileV1.UEditorUploadResponse{
			State: trans.Ptr(err.Error()),
		}, err
	}

	return &fileV1.UEditorUploadResponse{
		State:    trans.Ptr(StateOK),
		Original: trans.Ptr(req.GetSourceFileName()),
		Title:    trans.Ptr(req.GetSourceFileName()),
		Url:      trans.Ptr(downloadUrl),
		Size:     trans.Ptr(int32(len(req.GetFile()))),
		Type:     trans.Ptr(path.Ext(req.GetSourceFileName())),
	}, nil
}
