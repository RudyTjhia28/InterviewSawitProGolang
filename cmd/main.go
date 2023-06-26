package main

import (
	"os"

	"interviewsawitprogolang/repository"

	"interviewsawitprogolang/handler"

	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, World!",
		})
	})

	router.Run(":1323")
	e.Logger.Fatal(e.Start(":1323"))

}

func newServer() *handler.Server {
	dbDsn := os.Getenv("DATABASE_URL")
	var repo repository.RepositoryInterface = repository.NewRepository(repository.NewRepositoryOptions{
		Dsn: dbDsn,
	})
	opts := handler.NewServerOptions{
		Repository: repo,
	}
	return handler.NewServer(opts)
}
