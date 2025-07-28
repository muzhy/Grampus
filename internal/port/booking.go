package port

// "Grampus/internal/booking"

// func AddBookingRouter(router *gin.Engine, bookRepo booking.BookingRepository) {
// bookRouter := router.Group("/booking")
// router.GET("/book", GetAllBooks(bookRepo))
// router.POST("/book", CreateBook(bookRepo))
// // 不提供删除接口，账簿本身不支持更新，只是一个入口而已
// router.POST("/account", CreateAccount(bookRepo))
// }

// func GetAllBooks(repo booking.BookingRepository) func(c *gin.Context) {
// 	return func(c *gin.Context) {
// 		books, err := booking.GetAllBooks(repo)
// 		if err != nil {
// 			Fail(c, internal.ErrInternal, err.Error())
// 			return
// 		}

// 		Success(c, books)
// 	}
// }

// func CreateBook(repo booking.BookingRepository) func(c *gin.Context) {
// 	return func(c *gin.Context) {
// 		var book booking.Book
// 		if err := c.ShouldBindJSON(&book); err != nil {
// 			Fail(c, internal.ErrInvalidParam, err.Error())
// 			return
// 		}

// 		err := book.Create(repo)
// 		if err != nil {
// 			if bookErr, ok := err.(internal.Error); ok {
// 				switch bookErr.Code {
// 				case internal.ErrInternal:
// 					zap.L().Error("Failed to create book", zap.Error(err))
// 					Fail(c, internal.ErrInternal, "Failed to create book")
// 					return
// 				default:
// 					Fail(c, bookErr.Code, err.Error())
// 					return
// 				}
// 			}
// 			zap.L().Error("Failed to create book", zap.Error(err))
// 			Fail(c, internal.ErrInternal, "Failed to create book")
// 			return
// 		}

// 		zap.L().Info("Book created successfully",
// 			zap.String("name", book.Name),
// 			zap.String("ID", book.ID))

// 		Success(c, book)
// 	}
// }

// func CreateAccount(repo booking.BookingRepository) func(c *gin.Context) {
// 	return func(c *gin.Context) {
// 		var account booking.Account
// 		if err := c.BindJSON(&account); err != nil {
// 			Fail(c, internal.ErrInvalidParam, err.Error())
// 			return
// 		}
// 		err := account.Create(repo)
// 		if err != nil {
// 			if gerr, ok := err.(internal.Error); ok {
// 				Fail(c, gerr.Code, gerr.Error())
// 				return
// 			}
// 			Fail(c, internal.ErrInternal, err.Error())
// 			return
// 		}
// 		Success(c, account)
// 	}
// }
