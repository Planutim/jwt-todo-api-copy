package main

import (
	"jwt_rewrite/handlers"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
)

var (
	router = gin.Default()
	client *redis.Client
)

func init() {
	godotenv.Load()
}
func main() {
	mh, err := handlers.NewJwtHandler()
	if err != nil {
		panic(err)
	}
	router.POST("/login", mh.Login)
	router.POST("/todo", mh.TokenAuthMiddleWare(), mh.CreateTodo)
	router.GET("/todo", mh.TokenAuthMiddleWare(), mh.ListTodo)

	router.POST("/logout", mh.TokenAuthMiddleWare(), mh.Logout)
	router.POST("/token/refresh", mh.Refresh)
	router.POST("/register", mh.Register)

	log.Fatal(router.Run(":7000"))
}

func ConnectToRedis() error {
	dsn := "localhost:6379"
	client = redis.NewClient(&redis.Options{
		Addr: dsn,
	})
	_, err := client.Ping().Result()
	return err
}
