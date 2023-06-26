package models

// Define the registration request structure
type RegistrationRequest struct {
	PhoneNumber string `json:"phone_number"`
	FullName    string `json:"full_name"`
	Password    string `json:"password"`
}

// Define the registration response structure
type RegistrationResponse struct {
	ID int64 `json:"id"`
}

type User struct {
	ID          int64  `json:"id"`
	PhoneNumber string `json:"phone_number"`
	FullName    string `json:"full_name"`
	Password    string `json:"password"`
}
