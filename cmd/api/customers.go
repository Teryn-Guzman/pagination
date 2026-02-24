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