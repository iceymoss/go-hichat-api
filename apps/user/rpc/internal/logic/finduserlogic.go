package logic

import (
	"context"
	"database/sql"

	"github.com/iceymoss/go-hichat-api/apps/user/models"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type FindUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFindUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindUserLogic {
	return &FindUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FindUserLogic) FindUser(in *user.FindUserReq) (*user.FindUserResp, error) {
	var (
		userEntitys []*models.Users
		userMap     = make(map[uint64]*models.Users) // 用于去重
	)

	// 支持同时搜索多个条件，使用 map 去重
	// 1. 按手机号精准搜索
	if in.Phone != "" {
		userEntity, err := l.svcCtx.UserModels.FindOneByPhone(l.ctx, in.Phone)
		if err == nil && userEntity != nil {
			userMap[userEntity.Id] = userEntity
		}
	}

	// 2. 按邮箱精准搜索
	if in.Email != "" {
		userEntity, emailErr := l.svcCtx.UserModels.FindOneByEmail(l.ctx, sql.NullString{
			String: in.Email,
			Valid:  true,
		})
		if emailErr == nil && userEntity != nil {
			userMap[userEntity.Id] = userEntity
		}
	}

	// 3. 按名称模糊搜索（支持多个结果）
	if in.Name != "" {
		nameUsers, nameErr := l.svcCtx.UserModels.ListByName(l.ctx, in.Name)
		if nameErr == nil {
			for _, u := range nameUsers {
				if u != nil {
					userMap[u.Id] = u
				}
			}
		}
	}

	// 4. 按 ID 列表搜索
	if len(in.Ids) > 0 {
		idsUsers, idsErr := l.svcCtx.UserModels.ListByIds(l.ctx, in.Ids)
		if idsErr == nil {
			for _, u := range idsUsers {
				if u != nil {
					userMap[u.Id] = u
				}
			}
		}
	}

	// 将 map 转换为 slice
	for _, u := range userMap {
		userEntitys = append(userEntitys, u)
	}

	var resp []*user.UserEntity
	resp = Copier(userEntitys)

	return &user.FindUserResp{
		User: resp,
	}, nil
}
