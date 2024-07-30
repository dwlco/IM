package router

import (
	"ginchat/docs"
	"ginchat/service"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
)

func Router() *gin.Engine {
	r := gin.Default()
	docs.SwaggerInfo.BasePath = ""
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	//静态资源
	r.Static("/asset", "asset/")
	r.LoadHTMLGlob("views/*/*")

	//首页
	r.GET("/", service.GetIndex)
	r.GET("/index", service.GetIndex)
	//注册
	r.GET("/toRegister", service.ToRegister)

	r.POST("/searchFriends", service.SearchFriend)

	//聊天页面
	r.GET("/toChat", service.ToChat)
	r.GET("/chat", service.Chat)
	r.GET("/user/getUserList", service.GetUserList)
	r.POST("/user/createUser", service.CreateUser)
	r.GET("/user/deleteUser", service.DeleteUser)
	r.POST("/user/updateUser", service.UpdateUser)
	r.POST("/user/findUserByNameAndPwd", service.FindUserByNameAndPwd)

	//发送消息
	r.GET("/user/sendMsg", service.SendMessage)
	//私聊发送消息
	r.GET("/user/sendUserMsg", service.SendUserMsg)
	return r
}
