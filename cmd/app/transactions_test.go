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

func TestImportCourseExample(t *testing.T) {
	st := store.New()
	rec := post(t, newMux(st), "/transactions/import",
		`{"csv":"date,amount,description\n2026-10-01,-0.10,coffee\n2026-10-01,-0.20,tea\n2026-13-45,xx,broken"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200: %s", rec.Code, rec.Body)
	}
	var got struct {
		Imported int `json:"imported"`
		Rejected []struct {
			Line   int    `json:"line"`
			Reason string `json:"reason"`
		} `json:"rejected"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Imported != 2 || len(got.Rejected) != 1 || got.Rejected[0].Line != 4 || got.Rejected[0].Reason == "" {
		t.Errorf("unexpected body: %s", rec.Body)
	}
	if n := len(st.All()); n != 2 {
		t.Errorf("%d transactions stored; want 2", n)
	}
}

func TestImportWithNothingRejected(t *testing.T) {
	rec := post(t, newMux(store.New()), "/transactions/import", `{"csv":"2026-10-01,-1.00,bread"}`)
	if !strings.Contains(rec.Body.String(), `"rejected":[]`) {
		t.Errorf("rejected should be an empty list, not null: %s", rec.Body)
	}
}

func TestImportRejectsBadBody(t *testing.T) {
	for _, body := range []string{`{}`, `{"csv":42}`, `not json`} {
		rec := post(t, newMux(store.New()), "/transactions/import", body)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("body %s: status %d; want 400", body, rec.Code)
		}
	}
}
