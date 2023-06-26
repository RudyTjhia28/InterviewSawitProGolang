package handler

import (
	"encoding/json"
	"interviewsawitprogolang/models"
	"interviewsawitprogolang/repository"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	UserRepository repository.UserRepository
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{
		UserRepository: userRepo,
	}
}

func (h *UserHandler) RegisterUser(c echo.Context) error {
	// Parse the request body
	req := new(models.RegistrationRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request payload")
	}

	// Validation checks
	var validationErrors []string

	// Phone number validation
	if len(req.PhoneNumber) < 10 || len(req.PhoneNumber) > 13 {
		validationErrors = append(validationErrors, "Phone number must be between 10 and 13 characters")
	}

	if !strings.HasPrefix(req.PhoneNumber, "+62") {
		validationErrors = append(validationErrors, "Phone number must start with the Indonesia country code '+62'")
	}

	// Full name validation
	if len(req.FullName) < 3 || len(req.FullName) > 60 {
		validationErrors = append(validationErrors, "Full name must be between 3 and 60 characters")
	}

	// Password validation
	if len(req.Password) < 6 || len(req.Password) > 64 {
		validationErrors = append(validationErrors, "Password must be between 6 and 64 characters")
	} else {
		hasUppercase := false
		hasNumber := false
		hasSpecial := false

		for _, ch := range req.Password {
			switch {
			case 'A' <= ch && ch <= 'Z':
				hasUppercase = true
			case '0' <= ch && ch <= '9':
				hasNumber = true
			case ch == '!' || ch == '@' || ch == '#' || ch == '$' || ch == '%' || ch == '&':
				hasSpecial = true
			}
		}

		if !hasUppercase || !hasNumber || !hasSpecial {
			validationErrors = append(validationErrors, "Password must contain at least 1 uppercase letter, 1 number, and 1 special character")
		}
	}

	// If there are validation errors, return the error response
	if len(validationErrors) > 0 {
		return c.JSON(http.StatusBadRequest, validationErrors)
	}

	// Hash and salt the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to hash password")
	}

	// Create a new user
	user := models.User{
		PhoneNumber: req.PhoneNumber,
		FullName:    req.FullName,
		Password:    string(hashedPassword),
	}

	// Store the user in the database
	userID, err := h.UserRepository.CreateUser(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to create user")
	}

	// Return the user ID
	return c.JSON(http.StatusOK, map[string]interface{}{
		"id": userID,
	})
}

// LoginUser handles the login endpoint
func (h *UserHandler) LoginUser(c echo.Context) error {
	// Parse the request body
	req := new(models.LoginRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request payload")
	}

	// Retrieve the user from the database based on the phone number
	user, err := h.UserRepository.GetUserByPhoneNumber(req.PhoneNumber)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid phone number")
	}

	// Compare the provided password with the stored hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid password")
	}

	// Increment the number of successful logins for the user
	err = h.UserRepository.IncrementSuccessfulLogins(user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	// Generate JWT token
	token, err := generateJWT(user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	// Return the user ID and JWT token
	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":    user.ID,
		"token": token,
	})
}

// Generate JWT token
func generateJWT(userID int64) (string, error) {
	// Set the claims for the token, here i use jwt-go to make it easier rather than generating private-key
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour),
	}
	// Create a new JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with a predefined secret key
	signedToken, err := token.SignedString([]byte("secretKey"))
	if err != nil {
		return "", err
	}

	// Return the encoded token
	return signedToken, nil
}

func (h *UserHandler) GetMyProfile(c echo.Context) error {
	token := c.Request().Header.Get("Authorization")
	claims, err := VerifyJWT(token)
	var userInfo *models.User
	if err != nil {
		return c.JSON(http.StatusForbidden, "Invalid or expired token")
	}

	if userID, ok := claims["sub"].(float64); ok {
		userInfo, err = h.UserRepository.GetUserByID(userID)
	} else {
		return c.JSON(http.StatusInternalServerError, "Failed to extract user data")
	}

	if err != nil {
		return err
	}
	// Return the user's name and phone number
	return c.JSON(http.StatusOK, map[string]interface{}{
		"name":         userInfo.FullName,
		"phone_number": userInfo.PhoneNumber,
	})
}

func VerifyJWT(tokenString string) (jwt.MapClaims, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Provide the secret key or public key for token verification
		return []byte("secretKey"), nil
	})

	// Check for parsing errors
	if err != nil {
		return nil, err
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, err
	}

	// Retrieve the claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, err
	}

	return claims, nil
}

func (h *UserHandler) UpdateUserProfile(c echo.Context) error {
	token := c.Request().Header.Get("Authorization")
	claims, err := VerifyJWT(token)
	// Parse the request body
	var updateData struct {
		PhoneNumber *string `json:"phone_number"`
		FullName    *string `json:"full_name"`
	}
	userID := claims["sub"].(float64)
	err = json.NewDecoder(c.Request().Body).Decode(&updateData)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request body")
	}

	// Check if any field is provided
	if updateData.PhoneNumber == nil && updateData.FullName == nil {
		return c.JSON(http.StatusBadRequest, "No fields provided for update")
	}

	// Check if phone number already exists
	if updateData.PhoneNumber != nil {
		exists, err := h.UserRepository.CheckPhoneNumberExists(*updateData.PhoneNumber)
		if err != nil {
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, "Failed to update user profile")
		}
		if exists {
			return c.JSON(http.StatusConflict, "Phone number already exists")
		}
	}

	err = h.UserRepository.UpdateUserProfile(userID, updateData.PhoneNumber, updateData.FullName)
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, "Failed to update user profile")
	}

	return c.JSON(http.StatusOK, "User profile updated successfully")
}
