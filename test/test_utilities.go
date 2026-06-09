package test

import (
	"net/http/httptest"
	"testing"
)

func ValidateExpectedCodeAndBody(w *httptest.ResponseRecorder, t *testing.T, expectedCode int, expectedBody string) {
	if w.Code != expectedCode {
		t.Errorf("Expected %d, got %d", expectedCode, w.Code)
	}

	if w.Body.String() != expectedBody {
		t.Errorf("Expected %s, got %s", expectedBody, w.Body.String())
	}
}
