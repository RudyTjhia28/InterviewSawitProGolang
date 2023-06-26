package main

import (
	"database/sql"
	"interviewsawitprogolang/handler"
	"interviewsawitprogolang/repository"
	"net/http"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgresql://postgres:password1@localhost/interviewsawitpro?sslmode=disable")

	if err != nil {
		panic(err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	userHandler := handler.NewUserHandler(*userRepo)

	e := echo.New()
	e.POST("/register", userHandler.RegisterUser)
	e.GET("/login", userHandler.LoginUser)
	e.GET("/profile", userHandler.GetMyProfile, authenticateJWT)
	e.PUT("/updateProfile", userHandler.UpdateUserProfile, authenticateJWT)
	e.Start(":8080")

}

func authenticateJWT(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Missing authorization token")
		}

		// Verify and parse the JWT token
		_, err := handler.VerifyJWT(token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
		}

		// Retrieve the user ID from the token's claims

		// Perform any additional token validation or verification here
		// For example, you can check if the token is expired or if it has the required scope/permissions
		return next(c)
	}
}
