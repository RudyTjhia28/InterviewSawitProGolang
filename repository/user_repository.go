package repository

import (
	"database/sql"
	"errors"
	"interviewsawitprogolang/models"
	"strconv"
	"strings"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (r *UserRepository) CreateUser(user models.User) (int64, error) {
	query := `INSERT INTO crudassignment.users (phone_number, full_name, password,successful_logins) 
              VALUES ($1, $2, $3, 0) 
              RETURNING id`

	var id int64
	err := r.DB.QueryRow(query, user.PhoneNumber, user.FullName, user.Password).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetUserByPhoneNumber retrieves a user from the database based on the phone number
func (r *UserRepository) GetUserByPhoneNumber(phoneNumber string) (models.User, error) {
	query := `SELECT id, phone_number, full_name, password FROM crudassignment.users WHERE phone_number = $1`
	user := models.User{}
	err := r.DB.QueryRow(query, phoneNumber).Scan(&user.ID, &user.PhoneNumber, &user.FullName, &user.Password)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

// IncrementSuccessfulLogins increments the number of successful logins for the user
func (r *UserRepository) IncrementSuccessfulLogins(userID int64) error {
	query := `UPDATE crudassignment.users SET successful_logins = successful_logins + 1 WHERE id = $1`
	_, err := r.DB.Exec(query, userID)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) GetUserByID(userID float64) (*models.User, error) {
	query := "SELECT phone_number, full_name FROM crudassignment.users WHERE id = $1"

	row := r.DB.QueryRow(query, userID)
	user := &models.User{}
	err := row.Scan(&user.PhoneNumber, &user.FullName)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, err
	}

	return user, err
}

func (r *UserRepository) CheckPhoneNumberExists(phoneNumber string) (bool, error) {
	var exists bool

	err := r.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM crudassignment.users WHERE phone_number = $1)", phoneNumber).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *UserRepository) UpdateUserProfile(userID float64, phoneNumber *string, fullName *string) error {

	query, params := buildUpdateQuery(userID, phoneNumber, fullName)

	// Execute the update query
	_, err := r.DB.Exec(query, params...)
	if err != nil {
		return err
	}

	return nil
}

func buildUpdateQuery(userID float64, phoneNumber *string, fullName *string) (string, []interface{}) {
	var queryBuilder strings.Builder
	var params []interface{}

	// Build the query dynamically
	queryBuilder.WriteString("UPDATE crudassignment.users SET ")

	// Build the query and params for phoneNumber
	if phoneNumber != nil && *phoneNumber != "" {
		queryBuilder.WriteString("phone_number = $")
		queryBuilder.WriteString(strconv.Itoa(len(params) + 1))
		params = append(params, *phoneNumber)
	}

	// Build the query and params for fullName
	if fullName != nil && *fullName != "" {
		if len(params) > 0 {
			queryBuilder.WriteString(", ")
		}
		queryBuilder.WriteString("full_name = $")
		queryBuilder.WriteString(strconv.Itoa(len(params) + 1))
		params = append(params, *fullName)
	}

	// Add the user ID as the last parameter
	queryBuilder.WriteString(" WHERE ID = $")
	queryBuilder.WriteString(strconv.Itoa(len(params) + 1))
	params = append(params, userID)

	return queryBuilder.String(), params
}
