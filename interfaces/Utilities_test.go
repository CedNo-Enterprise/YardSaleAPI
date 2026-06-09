package interfaces

import (
	"GarageSaleAPI/test"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

type placeholder struct {
	name string
}

func Test_givenValidBody_whenDecode_thenReturnOk(t *testing.T) {
	w := httptest.NewRecorder()
	_, err := w.Write([]byte(`{"name": "GarageSaleAPI"}`))
	if err != nil {
		t.Fatal(err)
	}

	decoder := json.NewDecoder(w.Body)
	var p placeholder

	Decode(w, decoder, &p)

	expectedCode := http.StatusOK
	expectedBody := ""

	test.ValidateExpectedCodeAndBody(w, t, expectedCode, expectedBody)
}

func Test_givenInvalidBody_whenDecode_thenReturnsBadRequest(t *testing.T) {
	w := httptest.NewRecorder()

	decoder := json.NewDecoder(w.Body)
	var p placeholder

	Decode(w, decoder, &p)

	expectedCode := http.StatusBadRequest
	expectedBody := "bad request body\n"

	test.ValidateExpectedCodeAndBody(w, t, expectedCode, expectedBody)
}

func Test_givenValidObject_whenMarshal_thenReturnOk(t *testing.T) {
	w := httptest.NewRecorder()

	Marshal(w, placeholder{"GarageSaleAPI"})

	expectedCode := http.StatusOK
	expectedBody := ""

	test.ValidateExpectedCodeAndBody(w, t, expectedCode, expectedBody)
}

func Test_givenInvalidObject_whenMarshal_thenReturnBadRequest(t *testing.T) {
	w := httptest.NewRecorder()

	Marshal(w, math.NaN())

	expectedCode := http.StatusInternalServerError
	expectedBody := "internal server error\n"

	test.ValidateExpectedCodeAndBody(w, t, expectedCode, expectedBody)
}

func Test_givenMatchingContentType_whenValidateContentType_thenReturnOk(t *testing.T) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	ValidateContentType(w, req, "application/json")

	expectedCode := http.StatusOK
	expectedBody := ""

	test.ValidateExpectedCodeAndBody(w, t, expectedCode, expectedBody)
}

func Test_givenNonMatchingContentType_whenValidateContentType_thenReturnOk(t *testing.T) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	ValidateContentType(w, req, "multipart/form-data")

	expectedCode := http.StatusUnsupportedMediaType
	expectedBody := "invalid content type\n"

	test.ValidateExpectedCodeAndBody(w, t, expectedCode, expectedBody)
}
