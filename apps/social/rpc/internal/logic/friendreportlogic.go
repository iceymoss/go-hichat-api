package logic

import (
	"context"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type FriendReportLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendReportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendReportLogic {
	return &FriendReportLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FriendReportLogic) FriendReport(in *social.FriendReportReq) (*social.FriendReportResp, error) {
	reporterUid, err := strconv.ParseUint(in.UserId, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewReqParamErr(), "invalid reporter uid:%s err:%v", in.UserId, err)
	}

	targetUid, err := strconv.ParseUint(in.FriendUid, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewReqParamErr(), "invalid target uid:%s err:%v", in.FriendUid, err)
	}

	now := time.Now()
	report := &objects.FriendReport{
		ReporterUID: reporterUid,
		TargetUID:   targetUid,
		Reason:      in.Reason,
		CreatedAt:   &now,
	}

	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	if err := mysqlConn.Create(report).Error; err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "insert friend report err:%v req:%v", err, in)
	}

	return &social.FriendReportResp{}, nil
}
