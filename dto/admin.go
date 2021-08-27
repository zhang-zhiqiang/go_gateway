package dto

import "time"

type AdminInfoOutput struct {
	Id           int       `json:"id"`
	Name         string    `json:"name"`
	Avatar       string    `json:"avatar"`
	Introduction string    `json:"introduction"`
	Roles        []string  `json:"roles"`
	LoginTime    time.Time `json:"login_time"`
}
