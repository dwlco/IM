package service

import (
	"fmt"
	"ginchat/models"
	"ginchat/utils"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// GetIndex
// @Tags 首页
// @Success 200 {string} welcome
// @Router /index [get]
func GetIndex(c *gin.Context) {
	ind, err := template.ParseFiles("index.html")
	if err != nil {
		panic(err)
	}
	ind.Execute(c.Writer, "index")
	// c.JSON(http.StatusOK, gin.H{
	// 	"message": "welcome",
	// })

}

// 注册
func ToRegister(c *gin.Context) {
	ind, err := template.ParseFiles("views/user/register.html")
	if err != nil {
		panic(err)
	}
	ind.Execute(c.Writer, "register")
	// c.JSON(http.StatusOK, gin.H{
	// 	"message": "welcome",
	// })

}

// 消息页面
func ToChat(c *gin.Context) {
	ind, err := template.ParseFiles("views/chat/index.html",
		"views/chat/head.html",
		"views/chat/foot.html",
		"views/chat/tabmenu.html",
		"views/chat/concat.html",
		"views/chat/group.html",
		"views/chat/profile.html",
		"views/chat/main.html",
		"views/chat/createcom.html",
		"views/chat/userinfo.html")
	if err != nil {
		panic(err)
	}
	userId, _ := strconv.Atoi(c.Query("userId"))
	token := c.Query("token")
	user := models.UserBasic{}
	user.ID = uint(userId)
	user.Identity = token
	fmt.Print("ToChat>>>>>>>")
	ind.Execute(c.Writer, user)
	// c.JSON(http.StatusOK, gin.H{
	// 	"message": "welcome",
	// })

}

// GetUserList
// @Tags user
// @Success 200 {string} json{"code","message"}
// @Router /user/getUserList [get]
func GetUserList(c *gin.Context) {
	data := make([]*models.UserBasic, 10)
	data = models.GetUserList()
	c.JSON(http.StatusOK, gin.H{
		"message": data,
	})
}

// CreateUser
// @Summary 新增用户
// @Tags user
// @param name query string false "用户名"
// @param password query string false "密码"
// @param repassword query string false "确认密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/createUser [get]
func CreateUser(c *gin.Context) {
	user := models.UserBasic{}
	user.Name = c.PostForm("name")
	password := c.PostForm("password")
	repassword := c.PostForm("repassword")

	//校验输入的用户名和密码
	if user.Name == "" || password == "" {
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "用户名和密码不能为空",
		})
		return
	}

	//获取加密随机数
	salt := fmt.Sprintf("%06d", rand.Int31())

	if password != repassword {
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "两次密码不一致！",
		})
		return
	}

	data := models.FindUserByName((user.Name))
	if data.Name != "" {
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "用户名已注册！",
		})
		return
	}
	//user.Password = password
	user.PassWord = utils.MakePassword(password, salt)
	user.Salt = salt
	//print(user)
	models.CreateUser(user)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "创建成员成功",
	})
}

// DeleteUser
// @Summary 删除用户
// @Tags user
// @param id query string false "id"
// @Success 200 {string} json{"code","message"}
// @Router /user/deleteUser [get]
func DeleteUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.Query("id"))
	user.ID = uint(id)

	models.DeleteUser(user)
	c.JSON(http.StatusOK, gin.H{
		"message": "删除成员成功",
	})
}

// UpdateUser
// @Summary 更新用户信息
// @Tags user
// @param id formData string false "id"
// @param name formData string false "name"
// @param password formData string false "password"
// @param phone formData string false "phone"
// @param email formData string false "email"
// @Success 200 {string} json{"code","message"}
// @Router /user/updateUser [post]
func UpdateUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.PostForm("id"))
	user.ID = uint(id)
	user.Name = c.PostForm("name")
	user.PassWord = c.PostForm("password")
	user.Phone = c.PostForm("phone")
	user.Email = c.PostForm("email")
	_, err := govalidator.ValidateStruct(user)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{
			"message": "修改参数不匹配",
		})
	} else {
		models.UpdateUser(user)
		c.JSON(http.StatusOK, gin.H{
			"message": "更新成功",
		})
	}

}

// FindUserByNameAndPwd
// @Summary 查询用户信息
// @Tags user
// @param name formData string false "name"
// @param password formData string false "password"
// @Success 200 {string} json{"code","data"}
// @Router /user/findUserByNameAndPwd [post]
func FindUserByNameAndPwd(c *gin.Context) {
	data := models.UserBasic{}

	name := c.PostForm("name")

	password := c.PostForm("password")

	user := models.FindUserByName(name)
	if user.Name == "" {
		c.JSON(200, gin.H{
			"msg": "该用户不存在",
		})
		return
	}

	flag := utils.ValidPassword(password, user.Salt, user.PassWord)
	if !flag {
		c.JSON(200, gin.H{
			"msg": "密码不正确",
		})
		return
	}

	pwd := utils.MakePassword(password, user.Salt)
	data = models.FindUserByNameAndPwd(name, pwd)
	c.JSON(200, gin.H{
		"code":    0, //0成功，-1失败
		"message": "登录成功",
		"data":    data,
	})
}

// 防止跨域站点伪造请求
var upGrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func SendMessage(c *gin.Context) {
	ws, err := upGrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func(ws *websocket.Conn) {
		err = ws.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}(ws)
	Msghandler(ws, c)
}

func Msghandler(ws *websocket.Conn, c *gin.Context) {
	msg, err := utils.Subscribe(c, utils.PublishKey)
	if err != nil {
		fmt.Println(err)
	}
	tm := time.Now().Format("2006-01-02 15:04:05")
	m := fmt.Sprintf("[ws][%s]:%s", tm, msg)
	err = ws.WriteMessage(1, []byte(m))
	if err != nil {
		fmt.Println(err)
	}
}

func SendUserMsg(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}

func SearchFriend(c *gin.Context) {
	userIdInt, _ := strconv.Atoi(c.PostForm("userId"))
	userId := uint(userIdInt)
	users := models.SearchFriend(userId)
	utils.RespOKList(c.Writer, users, len(users))

}

func AddFriend(c *gin.Context) {
	userIdInt, _ := strconv.Atoi(c.PostForm("userId"))
	targetIdInt, _ := strconv.Atoi(c.PostForm("targetId"))
	userId := uint(userIdInt)
	targetId := uint(targetIdInt)
	code, msg := models.AddFriend(userId, targetId)
	if code == 1 {
		utils.RespOK(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}

}

func CreateCommunity(c *gin.Context) {
	ownerIdInt, _ := strconv.Atoi(c.PostForm("ownerId"))
	ownerId := uint(ownerIdInt)
	name := c.PostForm("name")
	community := models.Community{}

	community.OwnerId = ownerId
	community.Name = name
	fmt.Println(community)
	code, msg := models.CreateCommunity(community)
	if code == 1 {
		utils.RespOK(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}

}

// 加载群列表
func LoadCommunity(c *gin.Context) {
	ownerId, _ := strconv.Atoi(c.PostForm("ownerId"))
	data, msg := models.LoadCommunity(uint(ownerId))
	if len(data) == 0 {
		utils.RespFail(c.Writer, msg)
	} else {
		utils.RespOKList(c.Writer, data, len(data))
	}

}

// 加入群 userId uint, comId, uint
func JoinGroup(c *gin.Context) {
	userIdint, _ := strconv.Atoi(c.PostForm("userId"))
	comId := c.PostForm("comId")

	data, msg := models.JoinGroup(uint(userIdint), comId)
	if data == 0 {
		utils.RespOK(c.Writer, data, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}
