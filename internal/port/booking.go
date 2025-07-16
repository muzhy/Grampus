package port

import (
	Grampus "Grampus"
	"Grampus/internal/booking"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AddBookingRouter(router *gin.Engine, bookRepo booking.BookingRepository) {
	bookRouter := router.Group("/booking")
	{
		// bookRouter.GET("/", GetAllBooks)
		// bookRouter.GET("/:id", GetBookByID)
		bookRouter.POST("/", CreateBook(bookRepo))
		// bookRouter.PUT("/:id", UpdateBook)
		// bookRouter.DELETE("/:id", DeleteBook)
	}

}

func AllCurrenty(repo booking.BookingRepository) func(c *gin.Context) {
	return nil
}

func CreateBook(repo booking.BookingRepository) func(c *gin.Context) {
	return func(c *gin.Context) {
		var book booking.Book
		if err := c.ShouldBindJSON(&book); err != nil {
			Fail(c, Grampus.ErrInvalidParam, err.Error())
			return
		}

		err := book.Create(repo)
		if err != nil {
			if bookErr, ok := err.(Grampus.Error); ok {
				switch bookErr.Code {
				case Grampus.ErrInternal:
					zap.L().Error("Failed to create book", zap.Error(err))
					Fail(c, Grampus.ErrInternal, "Failed to create book")
					return
				default:
					Fail(c, bookErr.Code, err.Error())
					return
				}
			}
			zap.L().Error("Failed to create book", zap.Error(err))
			Fail(c, Grampus.ErrInternal, "Failed to create book")
			return
		}

		zap.L().Info("Book created successfully",
			zap.String("name", book.Name),
			zap.String("ID", book.ID))

		Success(c, book)
	}
}
