package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Teryn-Guzman/Lab-3/internal/validator"
)

type CustomerModel struct {
	DB *sql.DB
}

type Customer struct {
	ID          int64     `json:"customer_id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	CreatedAt   time.Time `json:"created_at"`
	NoShowCount int       `json:"no_show_count"`
	PenaltyFlag bool      `json:"penalty_flag"`
}

func (m CustomerModel) Insert(customer *Customer) error {

	query := `
		INSERT INTO customers (first_name, last_name, email, phone)
		VALUES ($1, $2, $3, $4)
		RETURNING customer_id, created_at, no_show_count, penalty_flag
	`

	args := []any{
		customer.FirstName,
		customer.LastName,
		customer.Email,
		customer.Phone,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(
		&customer.ID,
		&customer.CreatedAt,
		&customer.NoShowCount,
		&customer.PenaltyFlag,
	)
}
func (m CustomerModel) Get(id int64) (*Customer, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT customer_id, first_name, last_name, email, phone,
		       created_at, no_show_count, penalty_flag
		FROM customers
		WHERE customer_id = $1
	`

	var customer Customer

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&customer.ID,
		&customer.FirstName,
		&customer.LastName,
		&customer.Email,
		&customer.Phone,
		&customer.CreatedAt,
		&customer.NoShowCount,
		&customer.PenaltyFlag,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &customer, nil
}

func ValidateCustomer(v *validator.Validator, c *Customer) {

	v.Check(c.FirstName != "", "first_name", "must be provided")
	v.Check(len(c.FirstName) <= 100, "first_name", "must not exceed 100 characters")

	v.Check(c.LastName != "", "last_name", "must be provided")
	v.Check(len(c.LastName) <= 100, "last_name", "must not exceed 100 characters")

	if c.Email != "" {
		v.Check(len(c.Email) <= 255, "email", "must not exceed 255 characters")
	}
}