package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"atad-project/internal/csvimport"
	"atad-project/internal/money"
	"atad-project/internal/store"
)

type transactionInput struct {
	Date        string        `json:"date"`
	Amount      *money.Amount `json:"amount"`
	Description string        `json:"description"`
	Category    string        `json:"category"`
}

func createTransaction(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in transactionInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			msg := "body must be a JSON object with date, amount and description"
			if errors.Is(err, money.ErrInvalid) {
				msg = err.Error()
			}
			writeError(w, http.StatusBadRequest, "validation", msg)
			return
		}
		tx, err := in.validate()
		if err != nil {
			writeError(w, http.StatusBadRequest, "validation", err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, st.Add(tx))
	}
}

func (in transactionInput) validate() (store.Transaction, error) {
	if _, err := time.Parse(time.DateOnly, in.Date); err != nil {
		return store.Transaction{}, errors.New("date must be a real date written YYYY-MM-DD")
	}
	if in.Amount == nil {
		return store.Transaction{}, errors.New("amount is required")
	}
	desc := strings.TrimSpace(in.Description)
	if desc == "" {
		return store.Transaction{}, errors.New("description is required")
	}
	return store.Transaction{Date: in.Date, Amount: *in.Amount, Description: desc, Category: categorise(in.Category)}, nil
}

// categorise keeps the category the user gave, or falls back to "uncategorised".
func categorise(category string) string {
	category = strings.TrimSpace(category)
	if category == "" {
		return "uncategorised"
	}
	return category
}

func importTransactions(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			CSV *string `json:"csv"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.CSV == nil {
			writeError(w, http.StatusBadRequest, "validation", `body must be a JSON object with the file as a "csv" string`)
			return
		}
		rows, rejected := csvimport.Parse(*in.CSV)
		for _, row := range rows {
			st.Add(store.Transaction{
				Date:        row.Date,
				Amount:      row.Amount,
				Description: row.Description,
				Category:    categorise(row.Category),
			})
		}
		if rejected == nil {
			rejected = []csvimport.Rejection{}
		}
		writeJSON(w, http.StatusOK, map[string]any{"imported": len(rows), "rejected": rejected})
	}
}
