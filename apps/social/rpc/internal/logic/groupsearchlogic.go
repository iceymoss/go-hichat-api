package logic

import (
	"context"
	"strconv"
	"strings"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GroupSearchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupSearchLogic {
	return &GroupSearchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GroupSearch 搜索群：群号精确 OR 群名模糊，分页，并回填群成员数
func (l *GroupSearchLogic) GroupSearch(in *social.GroupSearchReq) (*social.GroupSearchResp, error) {
	keyword := strings.TrimSpace(in.Keyword)
	if keyword == "" {
		return &social.GroupSearchResp{}, nil
	}

	page := in.Page
	if page < 1 {
		page = 1
	}
	size := in.Size
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	groups, total, err := l.svcCtx.GroupsModel.SearchGroups(l.ctx, keyword, offset, size)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "search groups err %v keyword %v", err, keyword)
	}
	if len(groups) == 0 {
		return &social.GroupSearchResp{Total: total}, nil
	}

	// 批量统计成员数（group_members.group_id 为字符串列）
	ids := make([]string, 0, len(groups))
	for _, g := range groups {
		ids = append(ids, strconv.Itoa(g.Id))
	}
	countMap, err := l.svcCtx.GroupMembersModel.CountByGroupIds(l.ctx, ids)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "count group members err %v ids %v", err, ids)
	}

	list := make([]*social.GroupSearchItem, 0, len(groups))
	for _, g := range groups {
		gid := strconv.Itoa(g.Id)
		list = append(list, &social.GroupSearchItem{
			Id:          gid,
			Name:        g.Name,
			Icon:        g.Icon,
			Description: g.Description,
			CreatorUid:  g.CreatorUid,
			MemberCount: countMap[gid],
		})
	}

	return &social.GroupSearchResp{
		List:  list,
		Total: total,
	}, nil
}
