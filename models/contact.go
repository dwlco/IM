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

func AddFriend(userId uint, targetId uint) (int, string) {
	user := UserBasic{}

	if userId == targetId {
		return -1, "不能添加自己为好友"
	}

	if targetId != 0 {
		user = FindUserById(targetId)
		//存在目标账号
		if user.ID != 0 {
			//判断是否已存在好友关系
			contact0 := Contact{}
			utils.DB.Where("owner_id = ? and target_id = ? and type = 1", userId, targetId).Find(&contact0)
			if contact0.ID != 0 {
				return -1, "不能重复添加"
			}
			//已存在，返回错误提示
			//不存在，添加好友
			tx := utils.DB.Begin()
			//事务一旦开始，不论出现什么异常最终都会rollback
			//使用recover方法
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
				}
			}()

			contact := Contact{}
			contact.OwnerId = userId
			contact.TargetId = targetId
			contact.Type = 1
			if err := utils.DB.Create(&contact).Error; err != nil {
				tx.Rollback()
				return -1, "添加好友失败"
			}

			contact1 := Contact{}
			contact1.OwnerId = targetId
			contact1.TargetId = userId
			contact1.Type = 1
			if err := utils.DB.Create(&contact1).Error; err != nil {
				tx.Rollback()
				return -1, "添加好友失败"
			}
			tx.Commit()
			return 0, "添加成功"
		}
		return -1, "未找到此用户"
	}
	return -1, "好友id不能为空"
}
