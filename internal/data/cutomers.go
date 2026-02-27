package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func (m CustomerModel) Update(customer *Customer) error {
	query := `
		UPDATE customers
		SET first_name = $1,
		    last_name = $2,
		    email = $3,
		    phone = $4,
		    no_show_count = $5,
		    penalty_flag = $6
		WHERE customer_id = $7
		RETURNING customer_id
	`

	args := []any{
		customer.FirstName,
		customer.LastName,
		customer.Email,
		customer.Phone,
		customer.NoShowCount,
		customer.PenaltyFlag,
		customer.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int64
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	return nil
}
func (m CustomerModel) Delete(id int64) error {
	//  Check if the ID is valid
	if id < 1 {
		return ErrRecordNotFound
	}

	//  SQL query to delete a customer by ID
	query := `
		DELETE FROM customers
		WHERE customer_id = $1
	`

	//  Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the delete query
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Check how many rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows were deleted, the ID probably doesn't exist
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
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

func (m CustomerModel) GetAll(firstName string, lastName string, filters Filters) ([]*Customer, Metadata, error) {
	// SQL query to get all customers with optional filtering by first_name and last_name
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), customer_id, first_name, last_name, email, phone,
		       created_at, no_show_count, penalty_flag
		FROM customers
		WHERE ($1 = '' OR first_name ILIKE '%%' || $1 || '%%')
		  AND ($2 = '' OR last_name ILIKE '%%' || $2 || '%%')
		ORDER BY %s %s, customer_id ASC 
        LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the query
	rows, err := m.DB.QueryContext(ctx, query, firstName, lastName, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()
	totalRecords := 0
	
	// Prepare slice to store customers
	customers := []*Customer{}

	// Iterate over rows
	for rows.Next() {
		var c Customer
		err := rows.Scan(
			&totalRecords,
			&c.ID,
			&c.FirstName,
			&c.LastName,
			&c.Email,
			&c.Phone,
			&c.CreatedAt,
			&c.NoShowCount,
			&c.PenaltyFlag,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		customers = append(customers, &c)
	}

	// Check for errors during iteration
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// Calculate metadata
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return customers, metadata, nil
	}