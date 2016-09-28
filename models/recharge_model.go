package models

import "time"

type Recharge struct {
	RechargeId int64     `json:"rechargeid"`
	Amount     float32   `json:"amount"`
	Namespace  string    `json:"namespace"`
	LoginUser  string    `json:"LoginUser,omitempty"`
	CreateTime time.Time `json:"createtime,omitempty"`
	Status     string    `json:"status,omitempty"`
	StatusTime time.Time `json:"statustime,omitempty"`
}
