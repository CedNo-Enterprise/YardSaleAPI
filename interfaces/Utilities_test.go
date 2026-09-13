package interfaces

import (
	"GarageSaleAPI/test"
	"bytes"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type placeholder struct {
	name string
}

type decodeTarget struct {
	Name string `json:"name"`
}

func TestDecodeBody(t *testing.T) {
	type args struct {
		body        string
		contentType string
	}
	tests := []struct {
		name         string
		args         args
		wantContinue bool
		wantCode     int
		wantBody     string
		wantName     string
	}{
		{
			name:         "valid body",
			args:         args{body: `{"name": "GarageSaleAPI"}`, contentType: "application/json"},
			wantContinue: true,
			wantCode:     http.StatusOK,
			wantBody:     "",
			wantName:     "GarageSaleAPI",
		},
		{
			name:         "wrong content type",
			args:         args{body: `{"name": "GarageSaleAPI"}`, contentType: "multipart/form-data"},
			wantContinue: false,
			wantCode:     http.StatusUnsupportedMediaType,
			wantBody:     "invalid content type\n",
		},
		{
			name:         "missing content type",
			args:         args{body: `{"name": "GarageSaleAPI"}`, contentType: ""},
			wantContinue: false,
			wantCode:     http.StatusUnsupportedMediaType,
			wantBody:     "invalid content type\n",
		},
		{
			name:         "malformed body",
			args:         args{body: `{"name": `, contentType: "application/json"},
			wantContinue: false,
			wantCode:     http.StatusBadRequest,
			wantBody:     "bad request body\n",
		},
		{
			name:         "empty body",
			args:         args{body: "", contentType: "application/json"},
			wantContinue: false,
			wantCode:     http.StatusBadRequest,
			wantBody:     "bad request body\n",
		},
		{
			name:         "unknown field",
			args:         args{body: `{"nmae": "typo"}`, contentType: "application/json"},
			wantContinue: false,
			wantCode:     http.StatusBadRequest,
			wantBody:     "bad request body\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := test.CreateRequest(http.MethodPost, "/", strings.NewReader(tt.args.body), tt.args.contentType)

			var target decodeTarget
			got := DecodeBody(w, r, &target)

			if got != tt.wantContinue {
				t.Errorf("DecodeBody() = %v, want %v", got, tt.wantContinue)
			}
			test.ValidateExpectedCodeAndBody(w, t, tt.wantCode, tt.wantBody)
			if tt.wantContinue && target.Name != tt.wantName {
				t.Errorf("decoded Name = %q, want %q", target.Name, tt.wantName)
			}
		})
	}
}

func TestDecodeBody_oversizedBody(t *testing.T) {
	w := httptest.NewRecorder()
	oversized := `{"name": "` + strings.Repeat("x", maxRequestBodyBytes+1) + `"}`
	r := test.CreateRequest(http.MethodPost, "/", strings.NewReader(oversized), "application/json")

	var target decodeTarget
	if DecodeBody(w, r, &target) {
		t.Errorf("DecodeBody() = true, want false for an oversized body")
	}
	test.ValidateExpectedCodeAndBody(w, t, http.StatusBadRequest, "bad request body\n")
}

func TestDecodeBody_guardStopsTheHandler(t *testing.T) {
	w := httptest.NewRecorder()
	r := test.CreateRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name": "x"}`), "text/plain")

	handlerRan := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		var target decodeTarget
		if !DecodeBody(w, r, &target) {
			return
		}
		handlerRan = true
		WriteResponse(w, target, http.StatusCreated, "application/json")
	}

	handler(w, r)

	if handlerRan {
		t.Errorf("handler continued past a rejected body")
	}
	test.ValidateExpectedCodeAndBody(w, t, http.StatusUnsupportedMediaType, "invalid content type\n")
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
