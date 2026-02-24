package main

import (
	"net/http"
	"time"

	"github.com/Teryn-Guzman/Lab-3/internal/data"
	"github.com/Teryn-Guzman/Lab-3/internal/validator"
)

func (a *applicationDependencies) createReservationHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	var incomingData struct {
		Customer  string `json:"customer"`
		TableID   int    `json:"table_id"`
		TimeSlot  string `json:"time_slot"`
		PartySize int    `json:"party_size"`
	}

	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	

	parsedTime, err := time.Parse(time.RFC3339, incomingData.TimeSlot)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	reservation := data.Reservation{
		ID:        time.Now().Unix(),
		Customer:  incomingData.Customer,
		TableID:   incomingData.TableID,
		TimeSlot:  parsedTime,
		PartySize: incomingData.PartySize,
		Status:    "confirmed",
		CreatedAt: time.Now(),
	}
	v := validator.New()
	data.ValidateReservation(v, &reservation)

	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	dataEnvelope := envelope{
		"reservation": reservation,
	}

	err = a.writeJSON(w, http.StatusCreated, dataEnvelope, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
