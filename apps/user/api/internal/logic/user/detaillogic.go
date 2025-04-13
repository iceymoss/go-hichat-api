package user

import (
	"context"
	"fmt"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDetailLogic 获取用户信息
func NewDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DetailLogic {
	return &DetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DetailLogic) Detail(req *types.UserInfoReq) (resp *types.UserInfoResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	fmt.Println("用户id:", uid, "ctx:", l.ctx)

	rpcUserInfoResp, err := l.svcCtx.User.GetUserInfo(l.ctx, &user.GetUserInfoReq{
		Id: uid,
	})
	if err != nil {
		return nil, err
	}

	res := types.User{
		Id:           rpcUserInfoResp.User.Id,
		Mobile:       rpcUserInfoResp.User.Phone,
		Nickname:     rpcUserInfoResp.User.Nickname,
		Sex:          int(rpcUserInfoResp.User.Sex),
		Avatar:       rpcUserInfoResp.User.Avatar,
		LastLogin:    strconv.Itoa(int(rpcUserInfoResp.User.LastLogin)),
		Introduction: rpcUserInfoResp.User.Introduction,
		Email:        rpcUserInfoResp.User.Email,
	}

	return &types.UserInfoResp{Info: res}, nil
}
