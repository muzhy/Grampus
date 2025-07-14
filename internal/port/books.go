package port

import (
	"Grampus/internal/books"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AddBookRouter(router *gin.Engine, bookRepo books.BookRepository) {
	// 创建书籍相关的路由组
	bookRouter := router.Group("/books")
	{
		// bookRouter.GET("/", GetAllBooks)
		// bookRouter.GET("/:id", GetBookByID)
		bookRouter.POST("/", CreateBook(bookRepo))
		// bookRouter.PUT("/:id", UpdateBook)
		// bookRouter.DELETE("/:id", DeleteBook)
	}
}

func CreateBook(repo books.BookRepository) func(c *gin.Context) {
	return func(c *gin.Context) {
		var book books.Book
		if err := c.ShouldBindJSON(&book); err != nil {
			c.JSON(400, gin.H{"error": "Invalid book data"})
			return
		}

		err := book.Create(repo)
		if err != nil {
			zap.L().Warn("Failed to create book", zap.Error(err))
			// TODO 需要根据错误类型，返回不同的错误码
			c.JSON(500, gin.H{"error": "Failed to create book"})
			return
		}

		zap.L().Info("Book created successfully",
			zap.String("name", book.Name),
			zap.Int32("ID", book.ID))

		c.JSON(201, gin.H{
			"message": "Book created successfully",
			"book":    book,
		})
	}
}
