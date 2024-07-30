package main

import (
	"ginchat/router"

	"ginchat/utils"
)

func main() {
	//初始化
	utils.InitConfig()
	utils.InitMySql()
	utils.InitRedis()
	r := router.Router()
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
