package models

import (
	"fmt"
	"ginchat/utils"
	"time"

	"gorm.io/gorm"
)

type UserBasic struct {
	gorm.Model
	Name          string
	PassWord      string
	Phone         string `valid:"matches(^1[3-9]{1}\\d{9}$)"`
	Email         string `valid:"email"`
	Identity      string
	CilentIp      string
	ClientPort    string
	Salt          string
	LoginTime     time.Time
	HeartbeatTime time.Time
	LoginOutTime  time.Time
	Avatar        string
	IsLogOut      bool
	DeviceInfo    string
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}

func GetUserList() []*UserBasic {
	//定义一个结构体数组
	data := make([]*UserBasic, 10)
	utils.DB.Find(&data)
	for _, v := range data {
		fmt.Println(v)
	}

	return data
}

func FindUserByNameAndPwd(name string, passWord string) UserBasic {
	user := UserBasic{}
	utils.DB.Where("name = ? and pass_word = ?", name, passWord).First(&user)
	//在校验姓名和密码也就是验证登录时，以当前时间加密后作为校验源(Identity)存入数据库
	str := fmt.Sprintf("%d", time.Now().Unix())
	temp := utils.MD5Encode(str)
	utils.DB.Model(&user).Where("id", user.ID).Update("identity", temp)
	return user
}

func FindUserByName(name string) UserBasic {
	user := UserBasic{}
	fmt.Println("enter the function of FindUserByName...")
	utils.DB.Where("name = ?", name).First(&user)
	return user
}

func FindUserByPhone(phone string) *gorm.DB {
	user := UserBasic{}
	return utils.DB.Where("phone", phone).First(&user)
}

func FindUserByEmail(email string) *gorm.DB {
	user := UserBasic{}
	return utils.DB.Where("email", email).First(&user)
}

func FindUserById(id uint) UserBasic {
	user := UserBasic{}
	utils.DB.Where("id", id).First(&user)
	return user
}

func CreateUser(user UserBasic) *gorm.DB {
	//print(user)
	return utils.DB.Create(&user)
}

func DeleteUser(user UserBasic) *gorm.DB {
	return utils.DB.Delete(&user)
}

func UpdateUser(user UserBasic) *gorm.DB {
	return utils.DB.Model(&user).Updates(UserBasic{Name: user.Name, PassWord: user.PassWord, Phone: user.Phone, Email: user.Email, Avatar: user.Avatar})
}
