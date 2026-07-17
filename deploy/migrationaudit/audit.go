package main

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

type Auditor struct {
	db *gorm.DB
}

type Report struct {
	Version     int      `json:"version"`
	Database    Database `json:"database"`
	Summary     Summary  `json:"summary"`
	Findings    Findings `json:"findings"`
	GeneratedAt string   `json:"generatedAt"`
}

type Database struct {
	Driver string `json:"driver"`
	Name   string `json:"name,omitempty"`
}

type Summary struct {
	FindingGroups   int  `json:"findingGroups"`
	AffectedUsers   int  `json:"affectedUsers"`
	RequiresCleanup bool `json:"requiresCleanup"`
}

type Findings struct {
	DuplicateFriendRequests []DuplicateFriendRequest `json:"duplicateFriendRequests"`
	DuplicateGroupRequests  []DuplicateGroupRequest  `json:"duplicateGroupRequests"`
	LegacyGroupInvitations  []LegacyGroupInvitation  `json:"legacyGroupInvitations"`
	DuplicateFriends        []DuplicateFriend        `json:"duplicateFriends"`
	OneWayFriends           []OneWayFriend           `json:"oneWayFriends"`
	GroupMemberUniqueIndex  UniqueIndexFinding       `json:"groupMemberUniqueIndex"`
}

type DuplicateFriendRequest struct {
	UserID          uint64   `json:"userId"`
	RequestUserID   uint64   `json:"requestUserId"`
	RecordIDs       []uint64 `json:"recordIds"`
	KeepID          uint64   `json:"keepId"`
	Rule            string   `json:"rule"`
	AffectedUserIDs []uint64 `json:"affectedUserIds"`
}

type DuplicateGroupRequest struct {
	GroupID        string  `json:"groupId"`
	ApplicantID    string  `json:"applicantId"`
	RecordIDs      []int64 `json:"recordIds"`
	KeepID         int64   `json:"keepId"`
	Rule           string  `json:"rule"`
	LegacyCategory string  `json:"legacyCategory"`
}

type LegacyGroupInvitation struct {
	ID             int64  `json:"id"`
	GroupID        string `json:"groupId"`
	InviteeID      string `json:"inviteeId"`
	InviterID      string `json:"inviterId,omitempty"`
	JoinSource     int64  `json:"joinSource"`
	Classification string `json:"classification"`
}

type DuplicateFriend struct {
	UserID          uint64   `json:"userId"`
	FriendUserID    uint64   `json:"friendUserId"`
	RecordIDs       []uint64 `json:"recordIds"`
	KeepID          uint64   `json:"keepId"`
	Rule            string   `json:"rule"`
	AffectedUserIDs []uint64 `json:"affectedUserIds"`
}

type OneWayFriend struct {
	UserID       uint64 `json:"userId"`
	FriendUserID uint64 `json:"friendUserId"`
	RecordID     uint64 `json:"recordId"`
}

type UniqueIndexFinding struct {
	Present bool     `json:"present"`
	Valid   bool     `json:"valid"`
	Names   []string `json:"names"`
	Columns []string `json:"columns"`
}

type idList struct {
	ID uint64
}

func NewAuditor(db *gorm.DB) *Auditor {
	return &Auditor{db: db}
}

func (a *Auditor) Run(ctx context.Context, database Database, generatedAt string) (Report, error) {
	report := Report{Version: 1, Database: database, GeneratedAt: generatedAt}
	var err error

	if report.Findings.DuplicateFriendRequests, err = a.duplicateFriendRequests(ctx); err != nil {
		return Report{}, err
	}
	if report.Findings.DuplicateGroupRequests, err = a.duplicateGroupRequests(ctx); err != nil {
		return Report{}, err
	}
	if report.Findings.LegacyGroupInvitations, err = a.legacyGroupInvitations(ctx); err != nil {
		return Report{}, err
	}
	if report.Findings.DuplicateFriends, err = a.duplicateFriends(ctx); err != nil {
		return Report{}, err
	}
	if report.Findings.OneWayFriends, err = a.oneWayFriends(ctx); err != nil {
		return Report{}, err
	}
	if report.Findings.GroupMemberUniqueIndex, err = a.groupMemberUniqueIndex(); err != nil {
		return Report{}, err
	}

	report.finalize()
	return report, nil
}

func (a *Auditor) duplicateFriendRequests(ctx context.Context) ([]DuplicateFriendRequest, error) {
	type duplicate struct {
		UserID uint64
		ReqUID uint64
	}
	var duplicates []duplicate
	err := a.db.WithContext(ctx).Table("friend_requests").
		Select("user_id, req_uid").
		Where("COALESCE(handle_result, 0) = ? AND COALESCE(status, 1) = ?", 0, 1).
		Group("user_id, req_uid").Having("COUNT(*) > 1").
		Order("user_id, req_uid").Scan(&duplicates).Error
	if err != nil {
		return nil, fmt.Errorf("audit duplicate friend requests: %w", err)
	}

	findings := make([]DuplicateFriendRequest, 0, len(duplicates))
	for _, duplicate := range duplicates {
		var rows []idList
		if err := a.db.WithContext(ctx).Table("friend_requests").Select("id").
			Where("user_id = ? AND req_uid = ? AND COALESCE(handle_result, 0) = ? AND COALESCE(status, 1) = ?", duplicate.UserID, duplicate.ReqUID, 0, 1).
			Order("id").Scan(&rows).Error; err != nil {
			return nil, fmt.Errorf("list duplicate friend request ids: %w", err)
		}
		ids := uint64IDs(rows)
		findings = append(findings, DuplicateFriendRequest{
			UserID: duplicate.UserID, RequestUserID: duplicate.ReqUID, RecordIDs: ids,
			KeepID: ids[len(ids)-1], Rule: "keep_latest_normal_pending",
			AffectedUserIDs: sortedUint64s(duplicate.UserID, duplicate.ReqUID),
		})
	}
	return findings, nil
}

func (a *Auditor) duplicateGroupRequests(ctx context.Context) ([]DuplicateGroupRequest, error) {
	type duplicate struct {
		GroupID string
		ReqID   string
	}
	var duplicates []duplicate
	err := a.db.WithContext(ctx).Table("group_requests").
		Select("group_id, req_id").Where("COALESCE(handle_result, 0) = ?", 0).
		Group("group_id, req_id").Having("COUNT(*) > 1").
		Order("group_id, req_id").Scan(&duplicates).Error
	if err != nil {
		return nil, fmt.Errorf("audit duplicate group requests: %w", err)
	}

	findings := make([]DuplicateGroupRequest, 0, len(duplicates))
	for _, duplicate := range duplicates {
		type row struct{ ID int64 }
		var rows []row
		if err := a.db.WithContext(ctx).Table("group_requests").Select("id").
			Where("group_id = ? AND req_id = ? AND COALESCE(handle_result, 0) = ?", duplicate.GroupID, duplicate.ReqID, 0).
			Order("id").Scan(&rows).Error; err != nil {
			return nil, fmt.Errorf("list duplicate group request ids: %w", err)
		}
		ids := make([]int64, len(rows))
		for i := range rows {
			ids[i] = rows[i].ID
		}
		findings = append(findings, DuplicateGroupRequest{
			GroupID: duplicate.GroupID, ApplicantID: duplicate.ReqID, RecordIDs: ids,
			KeepID: ids[len(ids)-1], Rule: "keep_latest_pending", LegacyCategory: "mixed_sources_require_review",
		})
	}
	return findings, nil
}

func (a *Auditor) legacyGroupInvitations(ctx context.Context) ([]LegacyGroupInvitation, error) {
	type row struct {
		ID            int64
		GroupID       string
		ReqID         string
		InviterUserID *string
		JoinSource    *int64
	}
	var rows []row
	if err := a.db.WithContext(ctx).Table("group_requests").
		Select("id, group_id, req_id, inviter_user_id, join_source").
		Where("join_source IN ? OR inviter_user_id IS NOT NULL", []int{2, 3}).
		Order("id").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("audit legacy group invitations: %w", err)
	}

	findings := make([]LegacyGroupInvitation, 0, len(rows))
	for _, row := range rows {
		joinSource := int64(0)
		if row.JoinSource != nil {
			joinSource = *row.JoinSource
		}
		inviter := ""
		if row.InviterUserID != nil {
			inviter = *row.InviterUserID
		}
		classification := "manual_review"
		if joinSource == 2 && inviter != "" {
			classification = "recoverable_member_invitation"
		} else if joinSource == 3 {
			classification = "invite_link_not_user_invitation"
		}
		findings = append(findings, LegacyGroupInvitation{
			ID: row.ID, GroupID: row.GroupID, InviteeID: row.ReqID, InviterID: inviter,
			JoinSource: joinSource, Classification: classification,
		})
	}
	return findings, nil
}

func (a *Auditor) duplicateFriends(ctx context.Context) ([]DuplicateFriend, error) {
	type duplicate struct {
		UserID    uint64
		FriendUID uint64
	}
	var duplicates []duplicate
	if err := a.db.WithContext(ctx).Table("friends").Select("user_id, friend_uid").
		Group("user_id, friend_uid").Having("COUNT(*) > 1").Order("user_id, friend_uid").
		Scan(&duplicates).Error; err != nil {
		return nil, fmt.Errorf("audit duplicate friends: %w", err)
	}

	findings := make([]DuplicateFriend, 0, len(duplicates))
	for _, duplicate := range duplicates {
		var rows []idList
		if err := a.db.WithContext(ctx).Table("friends").Select("id").
			Where("user_id = ? AND friend_uid = ?", duplicate.UserID, duplicate.FriendUID).
			Order("id").Scan(&rows).Error; err != nil {
			return nil, fmt.Errorf("list duplicate friend ids: %w", err)
		}
		ids := uint64IDs(rows)
		findings = append(findings, DuplicateFriend{
			UserID: duplicate.UserID, FriendUserID: duplicate.FriendUID, RecordIDs: ids,
			KeepID: ids[0], Rule: "keep_earliest_relation_merge_metadata",
			AffectedUserIDs: sortedUint64s(duplicate.UserID, duplicate.FriendUID),
		})
	}
	return findings, nil
}

func (a *Auditor) oneWayFriends(ctx context.Context) ([]OneWayFriend, error) {
	type row struct {
		ID        uint64
		UserID    uint64
		FriendUID uint64
	}
	var rows []row
	err := a.db.WithContext(ctx).Table("friends AS f").
		Select("f.id, f.user_id, f.friend_uid").
		Where("NOT EXISTS (?)", a.db.Table("friends AS reverse").Select("1").
			Where("reverse.user_id = f.friend_uid AND reverse.friend_uid = f.user_id")).
		Order("f.user_id, f.friend_uid, f.id").Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("audit one-way friends: %w", err)
	}

	findings := make([]OneWayFriend, len(rows))
	for i, row := range rows {
		findings[i] = OneWayFriend{UserID: row.UserID, FriendUserID: row.FriendUID, RecordID: row.ID}
	}
	return findings, nil
}

func (a *Auditor) groupMemberUniqueIndex() (UniqueIndexFinding, error) {
	indexes, err := a.db.Migrator().GetIndexes("group_members")
	if err != nil {
		return UniqueIndexFinding{}, fmt.Errorf("audit group member indexes: %w", err)
	}

	finding := UniqueIndexFinding{Names: []string{}, Columns: []string{}}
	for _, index := range indexes {
		columns := index.Columns()
		unique, ok := index.Unique()
		if !ok || !unique || len(columns) != 2 || columns[0] != "group_id" || columns[1] != "user_id" {
			continue
		}
		finding.Present = true
		finding.Valid = true
		finding.Names = append(finding.Names, index.Name())
		finding.Columns = columns
	}
	sort.Strings(finding.Names)
	return finding, nil
}

func (r *Report) finalize() {
	affected := make(map[string]struct{})
	for _, finding := range r.Findings.DuplicateFriendRequests {
		for _, id := range finding.AffectedUserIDs {
			affected[fmt.Sprintf("user:%d", id)] = struct{}{}
		}
	}
	for _, finding := range r.Findings.DuplicateGroupRequests {
		affected["user:"+finding.ApplicantID] = struct{}{}
	}
	for _, finding := range r.Findings.LegacyGroupInvitations {
		affected["user:"+finding.InviteeID] = struct{}{}
		if finding.InviterID != "" {
			affected["user:"+finding.InviterID] = struct{}{}
		}
	}
	for _, finding := range r.Findings.DuplicateFriends {
		for _, id := range finding.AffectedUserIDs {
			affected[fmt.Sprintf("user:%d", id)] = struct{}{}
		}
	}
	for _, finding := range r.Findings.OneWayFriends {
		affected[fmt.Sprintf("user:%d", finding.UserID)] = struct{}{}
		affected[fmt.Sprintf("user:%d", finding.FriendUserID)] = struct{}{}
	}

	r.Summary.FindingGroups = len(r.Findings.DuplicateFriendRequests) + len(r.Findings.DuplicateGroupRequests) +
		len(r.Findings.LegacyGroupInvitations) + len(r.Findings.DuplicateFriends) + len(r.Findings.OneWayFriends)
	if !r.Findings.GroupMemberUniqueIndex.Valid {
		r.Summary.FindingGroups++
	}
	r.Summary.AffectedUsers = len(affected)
	r.Summary.RequiresCleanup = r.Summary.FindingGroups > 0
}

func uint64IDs(rows []idList) []uint64 {
	ids := make([]uint64, len(rows))
	for i := range rows {
		ids[i] = rows[i].ID
	}
	return ids
}

func sortedUint64s(values ...uint64) []uint64 {
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	return values
}

func databaseName(driver, dsn string) string {
	if driver == "sqlite" {
		return filepath.Base(strings.SplitN(dsn, "?", 2)[0])
	}
	if parsed, err := url.Parse(dsn); err == nil && parsed.Scheme != "" {
		return strings.TrimPrefix(parsed.Path, "/")
	}
	if driver == "postgres" {
		for _, field := range strings.Fields(dsn) {
			key, value, ok := strings.Cut(field, "=")
			if ok && (key == "dbname" || key == "database") {
				return strings.Trim(value, "'\"")
			}
		}
		return ""
	}
	parts := strings.Split(strings.TrimRight(dsn, "/"), "/")
	name := parts[len(parts)-1]
	if index := strings.IndexByte(name, '?'); index >= 0 {
		name = name[:index]
	}
	return name
}
