package test

import (
	"io"
	"net/http"
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

func CreateRequest(method string, target string, body io.Reader, contentType string) *http.Request {
	request := httptest.NewRequest(
		method,
		target,
		body,
	)
	request.Header.Set("Content-Type", contentType)
	return request
}

func CreateRequestWithPathParam(
	method string, target string, body io.Reader,
	pathParam string, pathParamValue string,
) *http.Request {
	request := httptest.NewRequest(method, target, body)

	request.SetPathValue(pathParam, pathParamValue)

	return request
}
