package logic

import (
	"context"
	"encoding/json"
	"errors"
	"sort"

	"github.com/iceymoss/go-hichat-api/apps/trend/models"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GetTrendDiscussesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTrendDiscussesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTrendDiscussesLogic {
	return &GetTrendDiscussesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetTrendDiscusses 获取动态的多级评论树（树形结构）
func (l *GetTrendDiscussesLogic) GetTrendDiscusses(in *trend.GetDiscussesReq) (*trend.DiscussesTreeResp, error) {
	// 验证请求参数
	if in.TrendId == 0 {
		return nil, errors.New("动态ID不能为空")
	}

	// 检查动态是否存在
	_, err := l.svcCtx.Trend.FindOne(l.ctx, in.TrendId)
	if err != nil {
		zLog.Error("GetTrendDiscusses.FindOne: get trend failed", zap.Any("trendId", in.TrendId), zap.Error(err))
		return nil, errors.New("动态不存在")
	}

	// 使用 keyset 分页获取一级评论
	rootDiscusses, _, err := l.svcCtx.TrendDiscuss.FindByTrendWithKeyset(
		l.ctx,
		in.TrendId,
		1,                            // 一级评论
		uint64(in.Pagination.LastId), // 游标ID
		50,
	)
	if err != nil {
		zLog.Error("GetTrendDiscusses.FindByTrendWithKeyset: 获取一级评论失败", zap.Any("trendId", in.TrendId), zap.Error(err))
		return nil, errors.New("获取评论失败")
	}

	lastItem := &models.TrendDiscuss{}

	// 构建评论树
	var discusses []*trend.Discuss
	if len(rootDiscusses) > 0 {
		// 获取所有一级评论的ID
		rootIds := make([]uint64, 0, len(rootDiscusses))
		for _, rd := range rootDiscusses {
			rootIds = append(rootIds, rd.Id)
		}

		// 获取所有子评论（一次性查询提高性能）
		allChildren, err := l.svcCtx.TrendDiscuss.FindChildrenByRootIds(l.ctx, rootIds)
		if err != nil {
			zLog.Error("GetTrendDiscusses.FindChildrenByRootIds: 获取子评论失败", zap.Any("trendId", in.TrendId), zap.Error(err))
			return nil, errors.New("获取评论失败")
		}

		// 构建子评论映射：rootID => []*models.TrendDiscuss
		childrenMap := l.groupChildrenByRoot(allChildren)

		// 构建完整的树形结构
		discusses = l.buildDiscussTrees(rootDiscusses, childrenMap)

		if len(rootDiscusses) > 0 {
			// 设置下一页的游标
			lastItem = rootDiscusses[len(rootDiscusses)-1]
		}
	}

	// 返回响应
	return &trend.DiscussesTreeResp{
		Discusses: discusses,
		Pagination: &trend.Pagination{
			LastId:   int32(lastItem.Id),
			LastTime: lastItem.Createtime.Unix(),
		},
	}, nil

}

// groupChildrenByRoot 按根评论ID分组子评论
func (l *GetTrendDiscussesLogic) groupChildrenByRoot(children []*models.TrendDiscuss) map[uint64][]*models.TrendDiscuss {
	grouped := make(map[uint64][]*models.TrendDiscuss)

	for _, c := range children {
		rootID := c.Rootid
		grouped[rootID] = append(grouped[rootID], c)
	}

	return grouped
}

// buildDiscussTrees 构建评论树结构
func (l *GetTrendDiscussesLogic) buildDiscussTrees(roots []*models.TrendDiscuss, childrenMap map[uint64][]*models.TrendDiscuss) []*trend.Discuss {
	// 先按创建时间倒序排序一级评论（最新优先）
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].Createtime.After(roots[j].Createtime)
	})

	// 递归构建树结构
	var trees []*trend.Discuss
	for _, root := range roots {
		tree := l.convertModelToTree(root, childrenMap[root.Id])
		trees = append(trees, tree)
	}

	return trees
}

// 递归构建单棵评论树
func (l *GetTrendDiscussesLogic) convertModelToTree(root *models.TrendDiscuss, allChildren []*models.TrendDiscuss) *trend.Discuss {
	// 先创建根节点
	treeNode := convertToReply(root)

	// 如果存在子评论，递归构建子树
	if len(allChildren) > 0 {
		// 创建子节点映射（按父评论分组）
		childrenMap := make(map[uint64][]*models.TrendDiscuss)
		for _, c := range allChildren {
			fatherID := uint64(c.Father)
			childrenMap[fatherID] = append(childrenMap[fatherID], c)
		}

		// 递归构建树
		treeNode.Children = l.buildSubTree(root.Id, childrenMap)
	}

	return treeNode
}

// 构建子树（递归实现）
func (l *GetTrendDiscussesLogic) buildSubTree(
	fatherID uint64,
	childrenMap map[uint64][]*models.TrendDiscuss,
) []*trend.Discuss {
	// 获取直接子评论
	children := childrenMap[fatherID]
	if len(children) == 0 {
		return nil
	}

	// 先按创建时间升序排序（越早回复越靠前）
	sort.Slice(children, func(i, j int) bool {
		return children[i].Createtime.Before(children[j].Createtime)
	})

	// 递归处理每个子节点
	var treeNodes []*trend.Discuss
	for _, child := range children {
		node := convertToReply(child)
		node.Children = l.buildSubTree(child.Id, childrenMap)
		treeNodes = append(treeNodes, node)
	}

	return treeNodes
}

// 将模型转换为 protobuf 响应
func convertToReply(discuss *models.TrendDiscuss) *trend.Discuss {
	// 反序列化 @用户列表
	var atUserIds []uint64
	if discuss.Idlist != "" {
		_ = json.Unmarshal([]byte(discuss.Idlist), &atUserIds)
	}

	return &trend.Discuss{
		Id:           discuss.Id,
		TrendId:      uint64(discuss.Trendid),
		Father:       uint64(discuss.Father),
		RootId:       discuss.Rootid,
		Replyer:      uint64(discuss.Replyer),
		UserId:       discuss.Userid,
		Level:        uint32(discuss.Level),
		Content:      discuss.Content,
		AtUserIds:    atUserIds,
		AgreeCount:   discuss.AgreeCount,
		DiscussCount: uint64(discuss.DiscussCount),
		State:        uint32(discuss.State),
		Read:         discuss.Read != 0, // 0:未读 1:已读
		CreateTime:   discuss.Createtime.Unix(),
		UpdateTime:   discuss.Updatetime.Unix(),
		Children:     []*trend.Discuss{}, // 初始化为空切片
	}
}
