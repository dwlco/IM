package models

import (
	"fmt"
	"ginchat/utils"

	"gorm.io/gorm"
)

// 人员关系
type Contact struct {
	gorm.Model
	OwnerId  uint //谁的关系信息
	TargetId uint //对应的谁
	Type     int  //对应的类型 1好友 2群 3
	Desc     string
}

func (table *Contact) TableName() string {
	return "contact"
}

func SearchFriend(userId uint) []UserBasic {
	Contacts := make([]Contact, 0)
	utils.DB.Where("owner_id = ? and type = 1", userId).Find(&Contacts)
	objIds := make([]uint64, 0)
	for _, v := range Contacts {
		fmt.Println(v)
		objIds = append(objIds, uint64(v.TargetId))
	}
	users := make([]UserBasic, 0)
	utils.DB.Where("id in ?", objIds).Find(&users)
	return users
}
