// Code generated by goctl. DO NOT EDIT.
// goctl 1.8.2

package types

type LoginReq struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginResp struct {
	Token  string `json:"token"`
	Expire int64  `json:"expire"`
}

type RegisterReq struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Sex      int    `json:"sex"`
	Avatar   string `json:"avatar"`
}

type RegisterResp struct {
	Token  string `json:"token"`
	Expire int64  `json:"expire"`
}

type User struct {
	Id           string `json:"id"`
	Mobile       string `json:"mobile"`
	Nickname     string `json:"nickname"`
	Sex          int    `json:"sex"`
	Avatar       string `json:"avatar"`
	LastLogin    string `json:"lastLogin"`
	Introduction string `json "introduction"`
	Email        string `json:"email"`
}

type UserInfoReq struct {
}

type UserInfoResp struct {
	Info User `json:"info"`
}
