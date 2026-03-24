package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Dailiduzhou/library_manage_sys/models"
	"github.com/Dailiduzhou/library_manage_sys/pkg/logger"
	"github.com/Dailiduzhou/library_manage_sys/services"
	"github.com/Dailiduzhou/library_manage_sys/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// BookHandler handles book-related HTTP requests.
type BookHandler struct {
	bookService services.BookService
}

// NewBookHandler creates a new BookHandler.
func NewBookHandler(bookService services.BookService) *BookHandler {
	return &BookHandler{
		bookService: bookService,
	}
}

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	userService services.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// BorrowHandler handles borrow-related HTTP requests.
type BorrowHandler struct {
	borrowService services.BorrowService
}

// NewBorrowHandler creates a new BorrowHandler.
func NewBorrowHandler(borrowService services.BorrowService) *BorrowHandler {
	return &BorrowHandler{borrowService: borrowService}
}

// @Summary 用户注册
// @Description 创建新用户账号
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "注册请求"
// @Success 200 {object} models.Response "注册成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 409 {object} models.Response "用户已存在"
// @Failure 500 {object} models.Response "服务器错误"
// @Router /api/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code: 400,
			Msg:  "参数设定错误",
		})
		return
	}

	newUser, err := h.userService.Register(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrUserAlreadyExists) {
			c.JSON(http.StatusBadRequest, models.Response{
				Code: 400,
				Msg:  "用户已存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.Response{
			Code: 500,
			Msg:  "创建用户失败",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: 200,
		Msg:  "注册成功",
		Data: gin.H{
			"username": newUser.Username,
			"user_id":  newUser.ID,
		},
	})
}

// @Summary 用户登录
// @Description 用户身份验证
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "登录请求"
// @Success 200 {object} models.Response "登录成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 403 {object} models.Response "认证失败"
// @Failure 500 {object} models.Response "服务器错误"
// @Router /api/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code: 400,
			Msg:  "参数设定错误",
		})
		return
	}

	user, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			c.JSON(http.StatusForbidden, models.Response{
				Code: 403,
				Msg:  "用户不存在",
			})
			return
		}
		c.JSON(http.StatusForbidden, models.Response{
			Code: 403,
			Msg:  "密码错误",
		})
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("role", user.Role)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code: 500,
			Msg:  "鉴权组件错误",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: 200,
		Msg:  "登陆成功",
		Data: gin.H{
			"use_id": user.ID,
			"role":   user.Role,
		},
	})
}

// @Summary 用户登出
// @Description 清除会话，登出当前用户
// @Tags auth
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} models.Response "登出成功"
// @Failure 500 {object} models.Response "服务器错误"
// @Router /api/auth/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code: 500,
			Msg:  "登出失败",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: 200,
		Msg:  "登出成功",
	})
}

// @Summary 创建图书
// @Description 添加新图书（管理员权限）
// @Tags books
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "书名"
// @Param author formData string true "作者"
// @Param summary formData string false "简介"
// @Param cover formData file false "封面图片"
// @Param initial_stock formData integer true "初始库存" minimum(0)
// @Success 200 {object} models.Response{data=models.Book} "创建成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 409 {object} models.Response "图书已存在"
// @Failure 500 {object} models.Response "服务器错误"
// @Router /api/admin/books [post]
func (h *BookHandler) CreateBook(c *gin.Context) {
	var req models.CreateBookRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code: 400,
			Msg:  "参数设定错误",
		})
		return
	}

	finalCoverPath := models.DefaultCoverPath
	if req.Cover != nil && req.Cover.Size > 0 {
		logger.Infof("有封面文件上传，大小: %d", req.Cover.Size)
		savePath, err := utils.SaveImages(c, req.Cover)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code: 500,
				Msg:  "图片保存失败",
			})
			return
		}
		finalCoverPath = savePath
	} else {
		logger.Infof("没有封面文件上传或文件为空，使用默认路径: %s", finalCoverPath)
	}

	finalSummary := req.Summary
	if finalSummary == "" {
		finalSummary = models.DefaultSummary
	}

	newBook, err := h.bookService.CreateBook(req.Title, req.Author, finalSummary, finalCoverPath, req.InitialStock)
	if err != nil {
		if req.Cover != nil && req.Cover.Size != 0 {
			if removeErr := utils.RemoveFile(finalCoverPath); removeErr != nil {
				logger.Infof("封面删除失败: %v", removeErr)
			}
		}

		if errors.Is(err, services.ErrBookAlreadyExists) {
			c.JSON(http.StatusConflict, models.Response{
				Code: 409,
				Msg:  "该图书已存在(书名和作者相同)",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.Response{
			Code: 500,
			Msg:  "创建图书失败",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: 200,
		Msg:  "图书创建成功",
		Data: newBook,
	})
}

// @Summary 获取图书列表
// @Description 按条件查询图书
// @Tags books
// @Security ApiKeyAuth
// @Produce json
// @Param title query string false "按书名模糊查询"
// @Param author query string false "按作者模糊查询"
// @Param summary query string false "按简介模糊查询"
// @Success 200 {object} models.Response{data=[]models.Book} "查询成功"
// @Failure 500 {object} models.Response "数据库错误"
// @Router /api/books [get]
func (h *BookHandler) GetBooks(c *gin.Context) {
	title := c.Query("title")
	author := c.Query("author")
	summary := c.Query("summary")

	books, err := h.bookService.GetBooks(title, author, summary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code: 500,
			Msg:  "数据库查询失败",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: 200,
		Msg:  "查询成功",
		Data: books,
	})
}

// @Summary 更新图书
// @Description 修改图书信息（管理员权限）
// @Tags books
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path uint true "图书ID"
// @Param title formData string false "新书名"
// @Param author formData string false "新作者"
// @Param summary formData string false "新简介"
// @Param cover formData file false "新封面"
// @Param stock formData integer false "当前库存" minimum(0)
// @Param total_stock formData integer false "总库存" minimum(0)
// @Success 200 {object} models.Response{data=models.Book} "更新成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 404 {object} models.Response "图书不存在"
// @Failure 500 {object} models.Response "服务器错误"
// @Router /api/admin/books/{id} [put]
func (h *BookHandler) UpdateBook(c *gin.Context) {
	id := c.Param("id")
	bookID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code: 400,
			Msg:  "无效的图书ID",
		})
		return
	}

	var req models.UpdateBookRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code: 400,
			Msg:  "参数设定错误",
		})
		return
	}
	req.ID = uint(bookID)

	var finalCoverPath string
	if req.Cover != nil && req.Cover.Size > 0 {
		logger.Infof("有封面文件上传，大小: %d", req.Cover.Size)
		savePath, err := utils.SaveImages(c, req.Cover)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code: 500,
				Msg:  "图片保存失败",
			})
			return
		}
		finalCoverPath = savePath
	}

	book, err := h.bookService.UpdateBook(req.ID, req.Title, req.Author, req.Summary, finalCoverPath, req.Stock, req.TotalStock)
	if err != nil {
		if errors.Is(err, services.ErrBookNotFound) {
			c.JSON(http.StatusNotFound, models.Response{
				Code: 404,
				Msg:  "图书不存在",
			})
			return
		}
		if errors.Is(err, services.ErrStockInvalid) {
			c.JSON(http.StatusBadRequest, models.Response{
				Code: 400,
				Msg:  err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.Response{
			Code: 500,
			Msg:  "更新失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: 200,
		Msg:  "图书更新成功",
		Data: *book,
	})
}

// @Summary 删除图书
// @Description 删除指定图书（管理员权限）
// @Tags books
// @Security ApiKeyAuth
// @Produce json
// @Param id path uint true "图书ID"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 404 {object} models.Response "图书不存在"
// @Failure 409 {object} models.Response "图书仍在借阅中"
// @Failure 500 {object} models.Response "服务器错误"
// @Router /api/admin/books/{id} [delete]
func (h *BookHandler) DeleteBooks(c *gin.Context) {
	id := c.Param("id")
	bookID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code: 400,
			Msg:  "无效的图书ID",
		})
		return
	}

	book, err := h.bookService.GetBookByID(uint(bookID))
	if err != nil {
		if errors.Is(err, services.ErrBookNotFound) {
			c.JSON(http.StatusNotFound, models.Response{
				Code: 404,
				Msg:  "图书不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.Response{
			Code: 500,
			Msg:  "系统繁忙,请稍后再试",
		})
		return
	}

	if book.Stock != book.TotalStock {
		c.JSON(http.StatusConflict, models.Response{
			Code: 409,
			Msg:  "图书仍在借阅中",
		})
		return
	}

	if err := h.bookService.DeleteBook(uint(bookID)); err != nil {
		if errors.Is(err, services.ErrBookNotFound) {
			c.JSON(http.StatusNotFound, models.Response{
				Code: 404,
				Msg:  "图书不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.Response{
			Code: 500,
			Msg:  "删除图书失败",
		})
		return
	}

	if book.CoverPath != models.DefaultCoverPath {
		if err := utils.RemoveFile(book.CoverPath); err != nil {
			logger.Infof("封面删除失败: %v", err)
		}
	}

	c.JSON(http.StatusOK, models.Response{
		Code: 200,
		Msg:  "删除图书成功",
	})
}

// @Summary 借阅图书
// @Description 创建借阅记录，用户借阅指定图书 (return_date 初始为 null)
// @Tags borrows
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body models.FindBookRequest true "借阅请求"
// @Success 200 {object} models.Response{data=models.BorrowRecord} "借阅成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 404 {object} models.Response "图书不存在"
// @Failure 409 {object} models.Response "库存不足"
// @Failure 500 {object} models.Response "服务器错误"
// @Router /api/borrows [post]
func (h *BorrowHandler) BorrowBook(c *gin.Context) {
	var req models.FindBookRequest
	userID := c.GetUint("user_id")
	if userID == 0 {
		logger.Info("严重错误: 上下文中没有获取到 user_id")
		c.JSON(http.StatusUnauthorized, models.Response{
			Code: 401,
			Msg:  "用户未登录或认证失效",
		})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code: 400,
			Msg:  "参数设定错误",
		})
		return
	}

	borrowRecord, err := h.borrowService.BorrowBook(userID, req.ID)
	if err != nil {
		if errors.Is(err, services.ErrBookNotFound) {
			c.JSON(http.StatusNotFound, models.Response{
				Code: 404,
				Msg:  "图书不存在",
			})
			return
		}
		if errors.Is(err, services.ErrNoStock) {
			c.JSON(http.StatusConflict, models.Response{
				Code: 409,
				Msg:  "图书库存不足",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.Response{
			Code: 500,
			Msg:  "系统繁忙,请稍后再试",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: 200,
		Msg:  "借书成功",
		Data: borrowRecord,
	})
}

// @Summary 归还图书
// @Description 更新借阅状态，用户归还图书
// @Tags borrows
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body models.FindBookRequest true "归还请求"
// @Success 200 {object} models.Response{data=models.BorrowRecord} "归还成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 404 {object} models.Response "记录不存在"
// @Failure 500 {object} models.Response "服务器错误"
// @Router /api/borrows/return [post]
func (h *BorrowHandler) ReturnBook(c *gin.Context) {
	var req models.FindBookRequest

	userID := c.GetUint("user_id")
	if userID == 0 {
		logger.Info("严重错误: 上下文中没有获取到 user_id")
		c.JSON(http.StatusUnauthorized, models.Response{
			Code: 401,
			Msg:  "用户未登录或认证失效",
		})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code: 400,
			Msg:  "参数设定错误",
		})
		return
	}

	borrowRecord, err := h.borrowService.ReturnBook(userID, req.ID)
	if err != nil {
		if errors.Is(err, services.ErrBookNotFound) {
			c.JSON(http.StatusNotFound, models.Response{
				Code: 404,
				Msg:  "图书不存在",
			})
			return
		}
		if errors.Is(err, services.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, models.Response{
				Code: 404,
				Msg:  "未找到该书的借阅记录或已归还",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.Response{
			Code: 500,
			Msg:  "系统错误",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: 200,
		Msg:  "还书成功",
		Data: borrowRecord,
	})
}

// @Summary 查询个人借书记录
// @Description 查询当前登录用户的借阅记录（需登录）
// @Tags borrows
// @Security ApiKeyAuth
// @Produce json
// @Param id path uint true "用户ID"
// @Success 200 {object} models.Response{data=[]models.BorrowRecord} "查询成功"
// @Failure 400 {object} models.Response "权限不足，无法查询他人记录"
// @Failure 404 {object} models.Response "查询成功,无借书记录"
// @Failure 500 {object} models.Response "用户ID解析错误或数据库查询失败"
// @Router /api/records/{id} [post]
func (h *BorrowHandler) BorrowRecords(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.ParseUint(id, 10, 32)

	session := sessions.Default(c)
	userIDSession := session.Get("user_id")
	if userIDSession == nil || uint(userID) != userIDSession.(uint) {
		c.JSON(http.StatusBadRequest, models.Response{
			Code: 400,
			Msg:  "仅可查询自己的借书记录",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code: 500,
			Msg:  "用户ID解析错误",
		})
		return
	}

	records, err := h.borrowService.GetUserRecords(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code: 500,
			Msg:  "数据库查询失败",
		})
		return
	}

	if len(records) == 0 {
		c.JSON(http.StatusNotFound, models.Response{
			Code: 404,
			Msg:  "查询成功,无借书记录",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: 200,
		Msg:  "查询成功",
		Data: records,
	})
}

// @Summary 查询所有借书记录
// @Description 查询所有借阅记录（需管理员权限）
// @Tags records
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} models.Response{data=[]models.BorrowRecord} "查询成功"
// @Failure 404 {object} models.Response "查询成功,无借书记录"
// @Failure 500 {object} models.Response "数据库查询失败"
// @Router /api/admin/records [get]
func (h *BorrowHandler) GetAllBorrowRecords(c *gin.Context) {
	records, err := h.borrowService.GetAllRecords()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code: 500,
			Msg:  "数据库查询失败",
		})
		return
	}

	if len(records) == 0 {
		c.JSON(http.StatusNotFound, models.Response{
			Code: 404,
			Msg:  "查询成功,无借书记录",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: 200,
		Msg:  "查询成功",
		Data: records,
	})
}

// @Summary 按用户ID查询借阅记录
// @Description 查询指定用户ID的借阅记录（需管理员权限）
// @Tags records
// @Security ApiKeyAuth
// @Produce json
// @Param id path uint true "用户ID"
// @Success 200 {object} models.Response{data=[]models.BorrowRecord} "查询成功"
// @Failure 404 {object} models.Response "查询成功,无借书记录"
// @Failure 500 {object} models.Response "用户ID解析错误或数据库查询失败"
// @Router /api/admin/records/{id} [post]
func (h *BorrowHandler) BorrowRecordsByID(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code: 500,
			Msg:  "用户ID解析错误",
		})
		return
	}

	records, err := h.borrowService.GetUserRecords(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code: 500,
			Msg:  "数据库查询失败",
		})
		return
	}

	if len(records) == 0 {
		c.JSON(http.StatusNotFound, models.Response{
			Code: 404,
			Msg:  "查询成功,无借书记录",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: 200,
		Msg:  "查询成功",
		Data: records,
	})
}
