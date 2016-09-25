package models

import "time"

type Recharge struct {
	Recharge_id int `json:"recharge_id, omitenpty"`
	Money       string
	FromUser    string
	ToUser      string
	Create_time time.Time
	Status      string
	Status_time time.Time
}
