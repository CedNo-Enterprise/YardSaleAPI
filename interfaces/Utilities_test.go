package interfaces

import (
	"GarageSaleAPI/test"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestWriteResponse(t *testing.T) {
	type args struct {
		w           *httptest.ResponseRecorder
		response    any
		status      int
		contentType string
	}
	tests := []struct {
		name            string
		args            args
		wantStatusCode  int
		wantBody        string
		wantContentType string
	}{
		{
			name: "Write response",
			args: args{
				w:           httptest.NewRecorder(),
				response:    "{name: username, email: valid@email}",
				status:      http.StatusOK,
				contentType: "application/json",
			},
			wantStatusCode:  http.StatusOK,
			wantBody:        `"{name: username, email: valid@email}"`,
			wantContentType: "application/json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			WriteResponse(tt.args.w, tt.args.response, tt.args.status, tt.args.contentType)

			if tt.args.w.Code != tt.wantStatusCode {
				t.Errorf("Expected %d, got %d", tt.wantStatusCode, tt.args.w.Code)
			}

			if strings.TrimSpace(tt.args.w.Body.String()) != tt.wantBody {
				t.Errorf("Expected %s, got %s", tt.wantBody, tt.args.w.Body.String())
			}

			if tt.args.w.Header().Get("Content-Type") != tt.wantContentType {
				t.Errorf("Expected %s, got %s", tt.wantContentType, tt.args.w.Header().Get("Content-Type"))
			}
		})
	}
}
