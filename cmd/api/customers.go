package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Teryn-Guzman/Lab-3/internal/data"
	"github.com/Teryn-Guzman/Lab-3/internal/validator"
)
func (a *applicationDependencies) createCustomerHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	var input struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
	}

	err := a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	customer := data.Customer{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Phone:     input.Phone,
	}

	v := validator.New()
	data.ValidateCustomer(v, &customer)

	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.customerModel.Insert(&customer)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/customers/%d", customer.ID))

	err = a.writeJSON(w, http.StatusCreated,
		envelope{"customer": customer}, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
func (a *applicationDependencies) displayCustomerHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	customer, err := a.customerModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	data := envelope{
		"customer": customer,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
func (a *applicationDependencies) updateCustomerHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Get the ID from the URL
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Fetch the existing customer
	customer, err := a.customerModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// Decode incoming JSON into a temporary struct
	var input struct {
		FirstName   *string `json:"first_name"`
		LastName    *string `json:"last_name"`
		Email       *string `json:"email"`
		Phone       *string `json:"phone"`
		NoShowCount *int    `json:"no_show_count"`
		PenaltyFlag *bool   `json:"penalty_flag"`
	}

	err = a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	//  Update only the fields provided
	if input.FirstName != nil {
		customer.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		customer.LastName = *input.LastName
	}
	if input.Email != nil {
		customer.Email = *input.Email
	}
	if input.Phone != nil {
		customer.Phone = *input.Phone
	}
	if input.NoShowCount != nil {
		customer.NoShowCount = *input.NoShowCount
	}
	if input.PenaltyFlag != nil {
		customer.PenaltyFlag = *input.PenaltyFlag
	}

	//  Validate the updated customer
	v := validator.New()
	data.ValidateCustomer(v, customer)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	//  Save the updates to the database
	err = a.customerModel.Update(customer)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	//  Return the updated customer
	dataEnvelope := envelope{
		"customer": customer,
	}

	err = a.writeJSON(w, http.StatusOK, dataEnvelope, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) deleteCustomerHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	//  Get the ID from the URL
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	//  Delete the customer
	err = a.customerModel.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	//  Return a success message
	data := envelope{
		"message": "customer successfully deleted",
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
func (a *applicationDependencies) listCustomersHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Create a struct to hold the query parameters
	var queryParametersData struct {
		FirstName string
		LastName  string
		data.Filters
	}

	// Get the query parameters from the URL
	queryParameters := r.URL.Query()

	// Load the query parameters into our struct with defaults
	queryParametersData.FirstName = a.getSingleQueryParameter(
		queryParameters,
		"first_name",
		"",
	)

	queryParametersData.LastName = a.getSingleQueryParameter(
		queryParameters,
		"last_name",
		"",
	)

	// Create a new validator instance
   	v := validator.New()

	queryParametersData.Filters.Page = a.getSingleIntegerParameter(
                                       queryParameters, "page", 1, v)
    queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(
                                       queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(
		queryParameters, "sort", "customer_id")

	queryParametersData.Filters.SortSafeList = []string{"customer_id", "first_name",
		"-customer_id", "-first_name"}
 
	// Check if our filters are valid
    data.ValidateFilters(v, queryParametersData.Filters)
    if !v.IsEmpty() {
       a.failedValidationResponse(w, r, v.Errors)
       return
   }


	//  Call the model's GetAll with optional filters
	customers, metadata, err := a.customerModel.GetAll(
		queryParametersData.FirstName,
		queryParametersData.LastName,
		queryParametersData.Filters,
	)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	//  Wrap the results in an envelope and return JSON
	data := envelope{
		"customers": customers,
		"@metadata": metadata,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}