package utils

import (
	"context"
	"strconv"
)

const Identify = "hichat2.com"

func GetUserString(ctx context.Context) string {
	uid := ctx.Value(Identify).(string)
	return uid
}

func GetUser(ctx context.Context) int {
	uidStr := ctx.Value(Identify).(string)
	uid, _ := strconv.Atoi(uidStr)
	return uid
}
