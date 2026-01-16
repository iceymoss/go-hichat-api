package user

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSearchUserLogic 搜索用户
func NewSearchUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchUserLogic {
	return &SearchUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchUserLogic) SearchUser(req *types.SearchUserReq) (resp *types.SearchUserResp, err error) {
	// 验证至少提供一个搜索条件
	if req.Name == "" && req.Phone == "" && req.Email == "" && len(req.Ids) == 0 {
		return nil, errors.New(xerr.ErrBadRequest, "请至少提供一个搜索条件（名称、手机号、邮箱或用户ID）")
	}

	// 调用 RPC 搜索用户
	findResp, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{
		Name:  req.Name,
		Phone: req.Phone,
		Email: req.Email,
		Ids:   req.Ids,
	})
	if err != nil {
		return nil, errors.New(xerr.ErrInternalServer, "搜索用户失败，请稍后重试")
	}

	// 转换响应
	var users []types.User
	if findResp != nil && findResp.User != nil {
		for _, u := range findResp.User {
			if u != nil {
				users = append(users, types.User{
					Id:           u.Id,
					Mobile:       u.Phone,
					Nickname:     u.Nickname,
					Sex:          int(u.Sex),
					Avatar:       u.Avatar,
					LastLogin:    "",
					Introduction: u.Introduction,
					Email:        u.Email,
					Region:       u.Region,
					Occupation:   u.Occupation,
					Tags:         u.Tags,
				})
			}
		}
	}

	return &types.SearchUserResp{
		Users: users,
	}, nil
}
