//go:build wireinject
// +build wireinject

package main

import (
	controller "github.com/Dailiduzhou/library_manage_sys/controllers"
	"github.com/Dailiduzhou/library_manage_sys/repositories"
	"github.com/Dailiduzhou/library_manage_sys/services"
	"github.com/google/wire"
	"gorm.io/gorm"
)

var repositorySet = wire.NewSet(
	repositories.NewGormBookRepository,
	repositories.NewGormUserRepository,
	repositories.NewGormBorrowRepository,
	repositories.NewGormTransactor,
)

var serviceSet = wire.NewSet(
	services.NewBookService,
	services.NewUserService,
	services.NewBorrowService,
)

var handlerSet = wire.NewSet(
	controller.NewBookHandler,
	controller.NewUserHandler,
	controller.NewBorrowHandler,
)

func initializeHandlers(db *gorm.DB) (*appHandlers, error) {
	wire.Build(
		repositorySet,
		serviceSet,
		handlerSet,
		wire.Struct(new(appHandlers), "*"),
	)
	return nil, nil
}
