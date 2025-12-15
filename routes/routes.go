package routes

import (
	controller "github.com/Dailiduzhou/library_manage_sys/controllers"
	"github.com/Dailiduzhou/library_manage_sys/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		// 认证路由 (无需认证)
		auth := api.Group("/auth")
		{
			auth.POST("/register", controller.Register)
			auth.POST("/login", controller.Login)
		}

		// 需要认证的路由
		api.Use(middleware.AuthRequired())
		{
			// 用户登出路由
			api.POST("/auth/logout", controller.Logout)

			// 图书资源路由
			books := api.Group("/books")
			{
				// 获取图书列表 (带查询参数)
				// 支持标题、作者和简介的模糊搜索
				books.GET("", controller.GetBooks)

				// 需要管理员权限的路由
				books.Use(middleware.AdminRequired())
				{
					// 创建新图书
					books.POST("", controller.CreateBook)

					// 更新特定图书 (ID 从路径获取)
					books.PUT("/:id", controller.UpdateBook)

					// 删除特定图书 (ID 从路径获取)
					books.DELETE("/:id", controller.DeleteBooks)
				}
			}

			// 借阅资源路由
			borrows := api.Group("/borrows")
			{
				// 创建借阅记录 (借书)
				borrows.POST("", controller.BorrowBook)

				// 创建归还记录 (还书) - 非标准但保持逻辑
				borrows.POST("/return", controller.ReturnBook)
			}
		}
	}
}
