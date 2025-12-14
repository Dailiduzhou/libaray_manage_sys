package middleware

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
)

func InitSession(r *gin.Engine) error {
	useRedis := os.Getenv("USE_REDIS")
	if useRedis == "" {
		useRedis = "true"
	}

	sessionSecret := make([]byte, 32)

	sessionSecret = []byte(os.Getenv("SESSION_SECRET"))
	if len(sessionSecret) == 0 {
		sessionSecret = []byte("aebd2a80a82c5067554dc481e1dc7615d8d30c075e1424d6843d934479e4786e")
	}

	var store sessions.Store
	var err error

	store, err = initRedisStore(sessionSecret)
	if err != nil {
		log.Printf("Redis 存储初始化失败，回退到 Cookie 存储: %v", err)
		store = initCookieStore(sessionSecret)
	}

	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   gin.Mode() == gin.ReleaseMode,
		SameSite: 0,
	})

	r.Use(sessions.Sessions("mysession", store))

	log.Println("Session 中间件初始化完成")
	return nil
}

func initRedisStore(sessionSecret []byte) (sessions.Store, error) {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	store, err := redis.NewStore(
		10,
		"tcp",
		redisAddr,
		"",
		redisPassword,
		sessionSecret,
	)

	if err != nil {
		return nil, fmt.Errorf("连接 Redis 失败: %w", err)
	}

	log.Println("Redis 存储初始化成功")
	return store, nil
}

func initCookieStore(sessionSecret []byte) sessions.Store {
	log.Println("使用 Cookie 存储（开发环境）")
	store := cookie.NewStore(sessionSecret)
	return store
}
