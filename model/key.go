package model

type Key struct {
	Id       uint   `json:"id"`
	Key      string `json:"key"`
	Status   int    `json:"status"`
	LockTime int    `json:"lock_time"`
	Type     string `json:"type"`
}
