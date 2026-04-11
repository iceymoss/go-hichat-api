package emoji

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/user/models"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteEmojiLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteEmojiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteEmojiLogic {
	return &DeleteEmojiLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DeleteEmojiLogic) DeleteEmoji(req *types.DeleteEmojiReq) (*types.DeleteEmojiResp, error) {
	uid := ctxdata.GetUId(l.ctx)
	userId, _ := strconv.ParseUint(uid, 10, 64)
	if userId == 0 {
		return nil, libErr.New(xerr.ErrBadRequest, "用户未登录")
	}

	m := models.NewUserEmojiModel()
	if err := m.Delete(l.ctx, req.Id, userId); err != nil {
		return nil, libErr.New(xerr.ErrInternalServer, "删除失败")
	}

	return &types.DeleteEmojiResp{}, nil
}
