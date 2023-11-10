package model

import "time"

type GetMemberListResponse struct {
	StartAt time.Time `json:"start_at"`
	Members []Member  `json:"members"`
}

type SetMemberListRequest struct {
	StartAt time.Time `json:"start_at"`
	Members []Member  `json:"members"`
}
