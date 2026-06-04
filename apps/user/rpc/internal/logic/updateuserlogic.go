package logic

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/user/models"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/encrypt"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/message/verification"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type UpdateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserLogic) UpdateUser(in *user.UpdateUserReq) (*user.UpdateUserResp, error) {
	if in.Id == "" {
		return nil, libErr.New(xerr.ErrBadRequest, "用户id不能为空")
	}

	userId, err := strconv.Atoi(in.Id)
	if err != nil {
		logger.Error("atoi error", zap.Any("id", in.Id), zap.Error(err))
		return nil, err
	}

	// get user
	userEntity, err := l.svcCtx.UserModels.FindOne(l.ctx, uint64(userId))
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, libErr.New(xerr.ErrNotFound, "用户不存在")
		}
		logger.Error("user not found", zap.Any("id", in.Id), zap.Error(err))
		return nil, err
	}

	if userEntity.Status != 1 {
		return nil, libErr.New(10005, "用户已禁用")
	}

	now := time.Now()
	userObj := models.Users{
		Id:        uint64(userId),
		UpdatedAt: now,
	}

	if in.Name != "" {
		userObj.Nickname = in.Name
	}

	// Phone 不允许通过此 API 更新（手机号不可修改）
	// if in.Phone != "" {
	// 	userObj.Phone = in.Phone
	// }

	// 邮箱更新逻辑：只有在提供了 EmailCode 时才允许更新邮箱
	// 这是为了支持邮箱绑定 API，普通更新 API 不应该传递 EmailCode
	if in.Email != "" && in.EmailCode != "" {
		currentEmail := getStringValue(userEntity.Email) // 获取当前邮箱（处理 sql.NullString）
		if in.Email != currentEmail {
			// 验证邮箱验证码
			codeSender, err := verification.GetCodeSender(verification.CodeTypeEmail)
			if err != nil {
				logger.Error("获取邮件验证码发送器失败", zap.Error(err))
				return nil, libErr.New(xerr.ErrInternalServer, "获取验证码发送器失败")
			}
			
			rdb := db.GetRedisConn()
			key := verification.GetRedisKey(verification.CodeTypeEmail, in.Email)
			pass, err := codeSender.VerifyCode(l.ctx, rdb, key, in.EmailCode)
			if err != nil {
				logger.Error("验证邮箱验证码失败", zap.Any("email", in.Email), zap.Error(err))
				return nil, libErr.New(xerr.ErrInternalServer, "验证码验证失败")
			}
			
			if !pass {
				return nil, libErr.New(xerr.ErrInvalidInput, "邮箱验证码错误或已过期")
			}
			
			// 验证通过，更新邮箱（设置为 sql.NullString）
			userObj.Email = sql.NullString{String: in.Email, Valid: true}
		}
	}
	// 如果没有提供 EmailCode，即使传递了 Email 也不更新（普通更新 API 不应该更新邮箱）

	if in.Sex != 0 {
		userObj.Sex = int(in.Sex)
	}

	if in.Password != "" {
		genPassword, err := encrypt.GenPasswordHash([]byte(in.Password))
		if err != nil {
			return nil, err
		}
		userObj.Password = string(genPassword)
	}

	if in.Avatar != "" {
		userObj.Avatar = in.Avatar
	}

	if in.Introduction != "" {
		userObj.Introduction = in.Introduction
	}

	if in.MomentsCover != "" {
		userObj.MomentsCover = in.MomentsCover
	}

	if in.Type != "" {
		userType, err := strconv.ParseInt(in.Type, 10, 64)
		if err == nil {
			userObj.Type = uint64(userType)
		}
	}

	// 处理新字段：region, occupation, tags（使用 sql.NullString）
	// 注意：由于 proto 的限制，无法区分字段未设置和空字符串
	// 策略：如果字段非空（去除空格后），设置为该值；否则不更新（保持原值）
	// 这样至少不会意外清空字段，但用户也无法通过 API 清空字段（需要特殊处理）
	if strings.TrimSpace(in.Region) != "" {
		userObj.Region = sql.NullString{String: strings.TrimSpace(in.Region), Valid: true}
	}
	// 如果为空字符串，不更新该字段（保持原值）

	if strings.TrimSpace(in.Occupation) != "" {
		userObj.Occupation = sql.NullString{String: strings.TrimSpace(in.Occupation), Valid: true}
	}

	if strings.TrimSpace(in.Tags) != "" {
		userObj.Tags = sql.NullString{String: strings.TrimSpace(in.Tags), Valid: true}
	}

	err = l.svcCtx.UserModels.UpdateByID(l.ctx, &userObj)
	if err != nil {
		logger.Error("UpdateUser.UpdateByID: update user failed", zap.Any("userId", userId), zap.Error(err))
		
		// 检查是否是邮箱唯一约束冲突
		if isEmailDuplicateError(err) {
			return nil, libErr.New(xerr.ErrInvalidInput, "该邮箱已被其他用户使用，请更换其他邮箱")
		}
		
		return nil, libErr.New(xerr.ErrInternalServer, "更新用户信息失败")
	}

	return &user.UpdateUserResp{}, nil
}

// isEmailDuplicateError 检查是否是邮箱唯一约束冲突错误
func isEmailDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	// MySQL 唯一约束错误：Error 1062 (23000): Duplicate entry 'xxx' for key 'users.uk_email'
	return strings.Contains(errStr, "Duplicate entry") && 
		   (strings.Contains(errStr, "uk_email") || strings.Contains(errStr, "email"))
}
