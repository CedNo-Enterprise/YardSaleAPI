package test

import (
	"net/http/httptest"
	"testing"
)

func setupW(code int, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	w.WriteHeader(code)
	_, err := w.WriteString(body)
	if err != nil {
		panic(err)
	}

	return w
}

func TestValidateExpectedCodeAndBody(t *testing.T) {
	type args struct {
		w            *httptest.ResponseRecorder
		t            *testing.T
		expectedCode int
		expectedBody string
	}

	var fakeT = &testing.T{}
	fakeT.Cleanup(func() {
		fakeT = &testing.T{}
	})

	tests := []struct {
		name          string
		args          args
		wantErr       bool
		wantErrString string
	}{
		{
			name: "matching code and body",
			args: args{
				w:            setupW(200, "{matching body}"),
				t:            fakeT,
				expectedCode: 200,
				expectedBody: "{matching body}",
			},
			wantErr: false,
		},
		{
			name: "non matching code",
			args: args{
				w:            setupW(415, "{matching body}"),
				t:            fakeT,
				expectedCode: 200,
				expectedBody: "{matching body}",
			},
			wantErr:       true,
			wantErrString: `{"Expected 200, got 415"}`,
		},
		{
			name: "non matching body",
			args: args{
				w:            setupW(200, "{matching body}"),
				t:            fakeT,
				expectedCode: 200,
				expectedBody: "{non matching body}",
			},
			wantErr:       true,
			wantErrString: `{"Expected {non matching body}, got {matching body}"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ValidateExpectedCodeAndBody(tt.args.w, tt.args.t, tt.args.expectedCode, tt.args.expectedBody)
			if fakeT.Failed() != tt.wantErr {
				t.Errorf("ValidateExpectedCodeAndBody() error = %v, wantErr %v", fakeT.Failed(), tt.wantErr)
			}
		})
	}
}
