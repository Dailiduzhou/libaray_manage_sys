package main

import controller "github.com/Dailiduzhou/library_manage_sys/controllers"

type appHandlers struct {
	book   *controller.BookHandler
	user   *controller.UserHandler
	borrow *controller.BorrowHandler
}
