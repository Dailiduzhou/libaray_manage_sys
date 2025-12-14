package models

import "mime/multipart"

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreateBookRequest struct {
	Title   string                `form:"title" binding:"required"`
	Author  string                `form:"author" binding:"required"`
	Summary string                `form:"summary" binding:"omitempty"`
	Cover   *multipart.FileHeader `form:"cover" binding:"omitempty"`
}
