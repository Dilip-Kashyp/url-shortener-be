package config

import (
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func ConnectRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"), // e.g., localhost:6379
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err := RedisClient.Ping(RedisClient.Context()).Result()
	if err != nil {
		log.Println("Failed to connect to Redis:", err)
	} else {
		log.Println("Redis connected successfully")
	}
}
