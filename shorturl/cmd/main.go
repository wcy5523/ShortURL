package main

import (
	"log"
	"net/http"
	"shorturl/config"
	"shorturl/handler"
	"shorturl/middleware"
	"shorturl/service"
	"shorturl/util"

	"github.com/gin-gonic/gin"
)

func main() {
	config.Load()

	if err := config.InitRedis(); err != nil {
		log.Fatalf("init redis failed: %v", err)
	}

	if err := config.InitBloomFilter(); err != nil {
		log.Fatalf("init bloom filter failed: %v", err)
	}

	if err := config.InitMySQL(); err != nil {
		log.Fatalf("init mysql failed: %v", err)
	}

	snowflake, err := util.NewSnowflake(config.AppConfig.Snowflake.WorkerID)
	if err != nil {
		log.Fatalf("init snowflake failed: %v", err)
	}

	shortURLService := service.NewShortURLService(snowflake)
	statsService := service.NewStatsService()
	statsService.Start()
	defer statsService.Stop()

	shortURLHandler := handler.NewShortURLHandler(shortURLService, statsService)
	userHandler := handler.NewUserHandler()

	gin.SetMode(config.AppConfig.Server.Mode)
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})

	r.Use(middleware.RateLimitMiddleware(
		config.RedisClient,
		config.AppConfig.RateLimit.WindowSeconds,
		config.AppConfig.RateLimit.MaxRequests,
	))

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/captcha", userHandler.SendCaptcha)
			auth.POST("/login/captcha", userHandler.LoginWithCaptcha)
			auth.POST("/forgot-password", userHandler.ForgotPassword)
			auth.POST("/reset-password", userHandler.ResetPassword)
		}

		authRequired := api.Group("", middleware.JWT())
		{
			authRequired.POST("/create", shortURLHandler.Create)
			authRequired.GET("/links", shortURLHandler.List)
			authRequired.DELETE("/links/:code", shortURLHandler.Delete)
		}
	}

	r.GET("/s/:code", shortURLHandler.Redirect)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	r.Static("/static", "./static")
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/admin")
	})
	r.GET("/admin", func(c *gin.Context) {
		c.File("./static/admin.html")
	})

	log.Printf("server starting on :%s", config.AppConfig.Server.Port)
	if err := r.Run(":" + config.AppConfig.Server.Port); err != nil {
		log.Fatalf("server start failed: %v", err)
	}
}
