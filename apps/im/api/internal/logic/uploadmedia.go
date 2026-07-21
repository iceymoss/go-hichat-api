package logic

import (
	"context"
	"io"

	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/storage"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"
)

// maxUploadSize 单个富媒体文件大小上限：100MB
const maxUploadSize = 100 * 1024 * 1024

// uploadMedia 校验并上传一个富媒体文件，按类型归档到 im/<kind> 目录。
// 返回访问 URL 与媒体类型（image/video/voice/file）。
func uploadMedia(ctx context.Context, store storage.FileStorage, filename string, size int64, file io.Reader) (url, kind string, err error) {
	if filename == "" {
		return "", "", libErr.New(xerr.ErrBadRequest, "请选择要上传的文件")
	}
	if size > maxUploadSize {
		return "", "", libErr.New(xerr.ErrBadRequest, "文件大小不能超过100MB")
	}

	kind = storage.ClassifyMedia(filename)
	folder := "im/" + kind

	url, err = store.UploadFile(ctx, file, filename, folder)
	if err != nil {
		return "", "", libErr.New(xerr.ErrInternalServer, "上传文件失败: "+err.Error())
	}
	return url, kind, nil
}
