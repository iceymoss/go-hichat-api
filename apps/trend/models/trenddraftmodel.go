package models

import (
	"context"
	"errors"
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/transaction"

	"gorm.io/gorm"
)

const trendDraftTable = "trend_drafts"

type TrendDraft struct {
	Id           uint64    `db:"id"`
	UserId       uint64    `db:"user_id"`
	Type         int64     `db:"type"`
	Content      string    `db:"content"`
	Title        string    `db:"title"`
	Resources    string    `db:"resources"`
	CoverUrl     string    `db:"cover_url"`
	ShareUrl     string    `db:"share_url"`
	PositionName string    `db:"position_name"`
	Longitude    float64   `db:"longitude"`
	Latitude     float64   `db:"latitude"`
	Scope        int64     `db:"scope"`
	OpenReply    int64     `db:"open_reply"`
	State        int64     `db:"state"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type TrendDraftModel interface {
	Save(ctx context.Context, data *TrendDraft) (*TrendDraft, error)
	FindLatestByUser(ctx context.Context, userId uint64) (*TrendDraft, error)
	FindOneByUser(ctx context.Context, userId uint64, id uint64) (*TrendDraft, error)
	SoftDelete(ctx context.Context, userId uint64, id uint64) error
}

type defaultTrendDraftModel struct {
	table string
}

func NewTrendDraftModel() TrendDraftModel {
	return &defaultTrendDraftModel{table: trendDraftTable}
}

func (m *defaultTrendDraftModel) db(ctx context.Context) *gorm.DB {
	return transaction.GetTransactionOrDB(ctx, db.GetMysqlConn(db.MYSQL_DB_HICHAT2)).Table(m.table)
}

func (m *defaultTrendDraftModel) Save(ctx context.Context, data *TrendDraft) (*TrendDraft, error) {
	now := time.Now()
	if data.State == 0 {
		data.State = 1
	}
	data.UpdatedAt = now

	conn := m.db(ctx)
	if data.Id == 0 {
		data.CreatedAt = now
		if err := conn.Create(data).Error; err != nil {
			return nil, err
		}
		return data, nil
	}

	updates := map[string]interface{}{
		"type":          data.Type,
		"content":       data.Content,
		"title":         data.Title,
		"resources":     data.Resources,
		"cover_url":     data.CoverUrl,
		"share_url":     data.ShareUrl,
		"position_name": data.PositionName,
		"longitude":     data.Longitude,
		"latitude":      data.Latitude,
		"scope":         data.Scope,
		"open_reply":    data.OpenReply,
		"state":         data.State,
		"updated_at":    data.UpdatedAt,
	}
	res := conn.Where("id = ? AND user_id = ? AND state != ?", data.Id, data.UserId, 0).Updates(updates)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return m.FindOneByUser(ctx, data.UserId, data.Id)
}

func (m *defaultTrendDraftModel) FindLatestByUser(ctx context.Context, userId uint64) (*TrendDraft, error) {
	var data TrendDraft
	err := m.db(ctx).Where("user_id = ? AND state = ?", userId, 1).Order("updated_at DESC").First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &data, err
}

func (m *defaultTrendDraftModel) FindOneByUser(ctx context.Context, userId uint64, id uint64) (*TrendDraft, error) {
	var data TrendDraft
	err := m.db(ctx).Where("id = ? AND user_id = ? AND state != ?", id, userId, 0).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &data, err
}

func (m *defaultTrendDraftModel) SoftDelete(ctx context.Context, userId uint64, id uint64) error {
	res := m.db(ctx).Where("id = ? AND user_id = ? AND state != ?", id, userId, 0).Updates(map[string]interface{}{
		"state":      0,
		"updated_at": time.Now(),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
