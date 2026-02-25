package routes

import (
	controller "github.com/Dailiduzhou/library_manage_sys/controllers"
	"github.com/Dailiduzhou/library_manage_sys/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterBookRouters(r *gin.Engine, bookHandler *controller.BookHandler, borrowHandler *controller.BorrowHandler) {
	api := r.Group("/api")

	authGroup := api.Group("/")
	authGroup.Use(middleware.AuthRequired())
	{
		authGroup.POST("/records/:id", borrowHandler.BorrowRecords)
		borrows := authGroup.Group("/borrows")
		{
			// 创建借阅记录 (借书)
			borrows.POST("", borrowHandler.BorrowBook)
			borrows.POST("/return", borrowHandler.ReturnBook)
		}

		authGroup.GET("/books", bookHandler.GetBooks)

		adminGroup := authGroup.Group("/admin")
		adminGroup.Use(middleware.AdminRequired())
		{
			// POST /books 创建
			adminGroup.POST("/books", bookHandler.CreateBook)
			// PUT /books/:id 更新
			adminGroup.PUT("/books/:id", bookHandler.UpdateBook)
			// DELETE /books/:id 删除
			adminGroup.DELETE("/books/:id", bookHandler.DeleteBooks)

			adminGroup.GET("/records", borrowHandler.GetAllBorrowRecords)

			adminGroup.POST("/records/:id", borrowHandler.BorrowRecordsByID)
		}
	}
}
