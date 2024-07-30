package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB  *gorm.DB
	Red *redis.Client
)

func InitConfig() {
	viper.SetConfigName("app")
	viper.AddConfigPath("config")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("config app inited......")

}

func InitMySql() {

	//自定义日志模板，打印sql语句
	newLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logger.Info,
		Colorful:      true,
	})

	DB, _ = gorm.Open(mysql.Open(viper.GetString("mysql.dns")), &gorm.Config{Logger: newLogger})
	fmt.Println("config mysql inited......")
	// user := &models.UserBasic{}
	// DB.Find(&user)
	// fmt.Println(user)
}

func InitRedis() {
	var ctx = context.Background()
	Red = redis.NewClient(&redis.Options{
		Addr:         viper.GetString("redis.addr"),
		Password:     viper.GetString("redis.password"),
		DB:           viper.GetInt("redis.DB"),
		PoolSize:     viper.GetInt("redis.poolSize"),
		MinIdleConns: viper.GetInt("redis.minIdleConns"),
	})

	pong, err := Red.Ping(ctx).Result()
	if err != nil {
		fmt.Println("init redis error", err)
	} else {
		fmt.Println("inited redis......", pong)
	}

}

const (
	PublishKey = "websocket"
)

// Publish发布消息到redis
func Publish(ctx context.Context, channel string, msg string) error {
	var err error
	fmt.Println("Publish....", msg)
	err = Red.Publish(ctx, channel, msg).Err()
	return err
}

// Subscribe订阅redis消息
func Subscribe(ctx context.Context, channel string) (string, error) {
	sub := Red.Subscribe(ctx, channel)

	msg, err := sub.ReceiveMessage(ctx)
	fmt.Println(msg)
	fmt.Println("Subscribe....", msg.Payload)
	return msg.Payload, err
}
