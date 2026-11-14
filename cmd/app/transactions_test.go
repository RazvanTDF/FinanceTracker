package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atad-project/internal/store"
)

func post(t *testing.T, h http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path, strings.NewReader(body)))
	return rec
}

func TestCreateTransaction(t *testing.T) {
	h := newMux(store.New())
	rec := post(t, h, "/transactions", `{"date":"2026-10-01","amount":"-12.50","description":"coffee"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d; want 201: %s", rec.Code, rec.Body)
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["id"] != 1.0 || got["amount"] != "-12.50" || got["category"] != "uncategorised" {
		t.Errorf("unexpected body: %s", rec.Body)
	}
}

func TestCreateTransactionRejectsBadInput(t *testing.T) {
	bodies := map[string]string{
		"amount as number": `{"date":"2026-10-01","amount":-12.5,"description":"coffee"}`,
		"bad amount":       `{"date":"2026-10-01","amount":"12,50","description":"coffee"}`,
		"missing amount":   `{"date":"2026-10-01","description":"coffee"}`,
		"impossible date":  `{"date":"2026-13-45","amount":"1.00","description":"coffee"}`,
		"wrong date form":  `{"date":"01.10.2026","amount":"1.00","description":"coffee"}`,
		"no description":   `{"date":"2026-10-01","amount":"1.00","description":"  "}`,
		"not json":         `date=2026-10-01`,
	}
	for name, body := range bodies {
		rec := post(t, newMux(store.New()), "/transactions", body)
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), `"error":"validation"`) {
			t.Errorf("%s: status %d, body %s; want 400 validation", name, rec.Code, rec.Body)
		}
	}
}

func TestResetRestartsIDs(t *testing.T) {
	h := newMux(store.New())
	body := `{"date":"2026-10-01","amount":"1.00","description":"x"}`
	post(t, h, "/transactions", body)
	post(t, h, "/transactions", body)
	if rec := post(t, h, "/reset", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("reset status = %d; want 204", rec.Code)
	}
	if !strings.Contains(post(t, h, "/transactions", body).Body.String(), `"id":1,`) {
		t.Error("first id after reset is not 1")
	}
}
