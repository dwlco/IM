package models

import (
	"fmt"
	"ginchat/utils"

	"gorm.io/gorm"
)

type Community struct {
	gorm.Model
	Name    string
	OwnerId uint
	Img     string
	Desc    string
}

func CreateCommunity(community Community) (int, string) {
	if len(community.Name) == 0 {
		return -1, "群名称不能为空"
	}
	if community.OwnerId == 0 {
		return -1, "请先登录"
	}

	if err := utils.DB.Create(&community).Error; err != nil {
		fmt.Println(err)
		return -1, "建群失败"
	}
	return 0, "建群成功"
}

func LoadCommunity(ownerId uint) ([]Community, string) {
	data := make([]Community, 10)
	utils.DB.Where("owner_id", ownerId).Find(&data)
	return data, "查询成功"
}

func JoinGroup(userId uint, comId string) (int, string) {
	contact := Contact{}
	contact.OwnerId = userId

	contact.Type = 2
	community := Community{}
	//首先查找此群是否存在
	//然后查找数据库中是否有这条记录

	//添加群成员
	utils.DB.Where("id = ? or name = ?", comId, comId).Find(&community)
	if community.Name == "" {
		return -1, "没找到群"
	}
	contact.TargetId = community.ID
	utils.DB.Where("owner_id = ? and target_id = ? ad type = 2", userId, contact.TargetId).Find(&contact)
	if !contact.CreatedAt.IsZero() {
		return -1, "已加过此群"
	} else {
		utils.DB.Create(&contact)
		return 0, "加群成功"
	}

}
