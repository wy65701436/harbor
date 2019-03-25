package models

import (
	"time"
)

type OIDCUserMetaData struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	UserID       int64     `orm:"column(user_id)" json:"user_id"`
	Name         string    `orm:"column(name)" json:"name"`
	Value        string    `orm:"column(value)" json:"value"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}
