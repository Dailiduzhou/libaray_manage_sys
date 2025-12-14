package main

import (
	"github.com/Dailiduzhou/libaray_manage_sys/config"
	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDB()

	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
}
