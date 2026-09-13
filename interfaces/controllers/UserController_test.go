package controllers

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/domain/token"
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/infrastructure/persistence/memory"
	"GarageSaleAPI/infrastructure/ratelimit"
	"GarageSaleAPI/interfaces"
	"GarageSaleAPI/test"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
)

func Test_addUser(t *testing.T) {
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}

	repo := &memory.InMemoryUserRepository{}
	tokenService := services.NewTokenService([]byte("f81d4fae-7dec-11d0-a765-00a0c91e6bf6"), 24*time.Hour)
	sessionService := services.NewSessionService(&memory.InMemoryRevokedTokenRepository{})
	controller := *NewUserController(services.NewUserService(repo, tokenService), sessionService, testLoginThrottle(), interfaces.NewAuthenticationMiddleware(tokenService, sessionService))

	tests := []struct {
		name       string
		args       args
		wantStatus int
	}{
		{
			name: "Add valid user",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequest(
					http.MethodPost,
					"/user/add",
					bytes.NewBufferString(`{
						"Username":  "Edgouille",
						"Password":  "MDP!@#111111111",
						"Email":     "email@gmail.com"
					}`),
					"application/json",
				),
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "Add user with invalid content-type",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequest(
					http.MethodPost,
					"/user/add",
					bytes.NewBufferString(`{
						"Username":  "Edgouille",
						"Password":  "MDP!@#111111111",
						"Email":     "email@gmail.com"
					}`),
					"",
				),
			},
			wantStatus: http.StatusUnsupportedMediaType,
		},
		{
			name: "Add invalid user",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequest(
					http.MethodPost,
					"/user/add",
					bytes.NewBufferString(`{
						"Username":  "Edgouille",
						"Password":  "MDP!@#111111111",
						"Email":     "email"
					}`),
					"application/json",
				),
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller.addUser(tt.args.w, tt.args.r)
			if tt.args.w.Code != tt.wantStatus {
				t.Errorf("addUser() got status code = %v, want = %v",
					tt.args.w.Code, tt.wantStatus)
			}
		})
	}
}

func Test_getUser(t *testing.T) {
	ctx := test.CreateTestContext(t)
	userRepo := &memory.InMemoryUserRepository{}
	tokenService := services.NewTokenService([]byte("f81d4fae-7dec-11d0-a765-00a0c91e6bf6"), 24*time.Hour)
	sessionService := services.NewSessionService(&memory.InMemoryRevokedTokenRepository{})
	service := services.NewUserService(userRepo, tokenService)
	controller := *NewUserController(service, sessionService, testLoginThrottle(), interfaces.NewAuthenticationMiddleware(tokenService, sessionService))
	creationTime := time.Now()
	userId := uuid.NewString()
	userToAdd := user.CreateUser(userId, "Edgouille", "MDP!@#111111111", "email@email.com", creationTime)

	e := userRepo.Create(ctx, userToAdd)
	if e != nil {
		t.Fatal(e.Error())
	}

	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name           string
		args           args
		wantStatusCode int
		wantBody       string
	}{
		{
			name: "Get user",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequestWithPathParam(http.MethodGet, "/user/", nil, "id", userId),
			},
			wantStatusCode: http.StatusOK,
			wantBody:       fmt.Sprintf(`{"username":"Edgouille","email":"email@email.com","created_at":"%v","updated_at":"%v"}`+"\n", creationTime.Format(time.RFC3339Nano), creationTime.Format(time.RFC3339Nano)),
		},
		{
			name: "Get nonexistent user",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequestWithPathParam(http.MethodGet, "/user/", nil, "id", uuid.NewString()),
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       "user not found\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller.getUser(tt.args.w, tt.args.r)
			if tt.wantStatusCode != tt.args.w.Code {
				t.Errorf("getUser() got status code = %v, want = %v", tt.args.w.Code, tt.wantStatusCode)
			}
			if tt.wantBody != tt.args.w.Body.String() {
				t.Errorf("getUser() got body = %v, want = %v", tt.args.w.Body, tt.wantBody)
			}
		})
	}
}

func TestUserController_login(t *testing.T) {
	ctx := test.CreateTestContext(t)
	userRepo := &memory.InMemoryUserRepository{}
	tokenService := services.NewTokenService([]byte("f81d4fae-7dec-11d0-a765-00a0c91e6bf6"), 24*time.Hour)
	sessionService := services.NewSessionService(&memory.InMemoryRevokedTokenRepository{})
	service := services.NewUserService(userRepo, tokenService)
	controller := *NewUserController(service, sessionService, testLoginThrottle(), interfaces.NewAuthenticationMiddleware(tokenService, sessionService))
	creationTime := time.Now()
	userToAdd := user.CreateUser(
		uuid.NewString(),
		"Edgouille",
		"$2a$14$/IhjU2PxRypamw1kLypfIeB28u32sgVtTL2EvCl8Ar.sUlPk77drO",
		"email@email.com",
		creationTime)

	e := userRepo.Create(ctx, userToAdd)
	if e != nil {
		t.Fatal(e.Error())
	}

	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name           string
		args           args
		wantStatusCode int
	}{
		{
			name: "Login",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequest(
					http.MethodPost,
					"/login",
					bytes.NewBufferString(`{
						"Email":     "email@email.com",
						"Password":  "MDP!@#111111111"
					}`),
					"application/json",
				),
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "Login",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequest(
					http.MethodPost,
					"/login",
					bytes.NewBufferString(`{
						"email":     "email@email.com",
						"password":  "invalidPassword"
					}`),
					"application/json",
				),
			},
			wantStatusCode: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller.login(tt.args.w, tt.args.r)

			if tt.wantStatusCode != tt.args.w.Code {
				t.Errorf("getUser() got status code = %v, want = %v", tt.args.w.Code, tt.wantStatusCode)
			}
		})
	}
}

// logoutTestServer wires a controller behind a real mux so logout requests go
// through the authentication middleware, the way they do in production.
func logoutTestServer(t *testing.T) (*http.ServeMux, *services.TokenService) {
	t.Helper()

	userRepo := &memory.InMemoryUserRepository{}
	tokenService := services.NewTokenService([]byte("f81d4fae-7dec-11d0-a765-00a0c91e6bf6"), 24*time.Hour)
	sessionService := services.NewSessionService(&memory.InMemoryRevokedTokenRepository{})
	service := services.NewUserService(userRepo, tokenService)
	controller := NewUserController(service, sessionService, testLoginThrottle(), interfaces.NewAuthenticationMiddleware(tokenService, sessionService))

	mux := http.NewServeMux()
	controller.AddUserHandlersToMux(mux)

	return mux, tokenService
}

func logoutRequest(authHeader string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/logout", nil)
	if authHeader != "" {
		r.Header.Set("Authorization", authHeader)
	}
	return r
}

func TestUserController_logout(t *testing.T) {
	mux, tokenService := logoutTestServer(t)

	tokenStr, _, err := tokenService.Generate(uuid.NewString())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, logoutRequest("Bearer "+tokenStr))

	if w.Code != http.StatusNoContent {
		t.Errorf("logout() got status code = %v, want = %v", w.Code, http.StatusNoContent)
	}
	if w.Body.Len() != 0 {
		t.Errorf("logout() got body = %q, want empty", w.Body.String())
	}
}

func TestUserController_logout_TokenIsDeadAfterwards(t *testing.T) {
	mux, tokenService := logoutTestServer(t)

	tokenStr, _, err := tokenService.Generate(uuid.NewString())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	first := httptest.NewRecorder()
	mux.ServeHTTP(first, logoutRequest("Bearer "+tokenStr))
	if first.Code != http.StatusNoContent {
		t.Fatalf("first logout got status code = %v, want = %v", first.Code, http.StatusNoContent)
	}

	second := httptest.NewRecorder()
	mux.ServeHTTP(second, logoutRequest("Bearer "+tokenStr))
	if second.Code != http.StatusUnauthorized {
		t.Errorf("reusing a logged-out token got status code = %v, want = %v", second.Code, http.StatusUnauthorized)
	}
}

func TestUserController_logout_LeavesOtherSessionsAlone(t *testing.T) {
	mux, tokenService := logoutTestServer(t)

	userId := uuid.NewString()
	phone, _, err := tokenService.Generate(userId)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	laptop, _, err := tokenService.Generate(userId)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, logoutRequest("Bearer "+phone))
	if w.Code != http.StatusNoContent {
		t.Fatalf("logout got status code = %v, want = %v", w.Code, http.StatusNoContent)
	}

	// The laptop session should still be able to log itself out.
	other := httptest.NewRecorder()
	mux.ServeHTTP(other, logoutRequest("Bearer "+laptop))
	if other.Code != http.StatusNoContent {
		t.Errorf("the user's other session got status code = %v, want = %v", other.Code, http.StatusNoContent)
	}
}

func TestUserController_logout_Unauthenticated(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
	}{
		{name: "no authorization header", authHeader: ""},
		{name: "malformed authorization header", authHeader: "some-token"},
		{name: "invalid token", authHeader: "Bearer not-a-real-token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, _ := logoutTestServer(t)

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, logoutRequest(tt.authHeader))

			if w.Code != http.StatusUnauthorized {
				t.Errorf("logout() got status code = %v, want = %v", w.Code, http.StatusUnauthorized)
			}
		})
	}
}

// brokenRevokedTokenRepository stands in for a database that rejects writes.
type brokenRevokedTokenRepository struct{}

func (brokenRevokedTokenRepository) Revoke(context.Context, *token.RevokedToken) error {
	return apperror.Internal(errors.New("database unavailable"))
}

func (brokenRevokedTokenRepository) IsRevoked(context.Context, string) (bool, error) {
	return false, nil
}

func (brokenRevokedTokenRepository) DeleteExpired(context.Context, time.Time) error {
	return apperror.Internal(errors.New("database unavailable"))
}

// If the revocation cannot be written, the client must not be told it was.
func TestUserController_logout_RevocationFailureIsNotReportedAsSuccess(t *testing.T) {
	tokenService := services.NewTokenService([]byte("f81d4fae-7dec-11d0-a765-00a0c91e6bf6"), 24*time.Hour)
	sessionService := services.NewSessionService(brokenRevokedTokenRepository{})
	service := services.NewUserService(&memory.InMemoryUserRepository{}, tokenService)
	controller := NewUserController(service, sessionService, testLoginThrottle(), interfaces.NewAuthenticationMiddleware(tokenService, sessionService))

	mux := http.NewServeMux()
	controller.AddUserHandlersToMux(mux)

	tokenStr, _, err := tokenService.Generate(uuid.NewString())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, logoutRequest("Bearer "+tokenStr))

	if w.Code == http.StatusNoContent {
		t.Fatal("logout reported success even though the revocation was not written")
	}
	if w.Code != http.StatusInternalServerError {
		t.Errorf("logout() got status code = %v, want = %v", w.Code, http.StatusInternalServerError)
	}
}

// Logging out ends a session, not the account.
func TestUserController_logout_UserCanLogInAgain(t *testing.T) {
	ctx := test.CreateTestContext(t)
	userRepo := &memory.InMemoryUserRepository{}
	tokenService := services.NewTokenService([]byte("f81d4fae-7dec-11d0-a765-00a0c91e6bf6"), 24*time.Hour)
	sessionService := services.NewSessionService(&memory.InMemoryRevokedTokenRepository{})
	service := services.NewUserService(userRepo, tokenService)
	controller := NewUserController(service, sessionService, testLoginThrottle(), interfaces.NewAuthenticationMiddleware(tokenService, sessionService))

	mux := http.NewServeMux()
	controller.AddUserHandlersToMux(mux)

	// bcrypt hash of "MDP!@#111111111"
	userToAdd := user.CreateUser(
		uuid.NewString(),
		"Edgouille",
		"$2a$14$/IhjU2PxRypamw1kLypfIeB28u32sgVtTL2EvCl8Ar.sUlPk77drO",
		"email@email.com",
		time.Now())
	if err := userRepo.Create(ctx, userToAdd); err != nil {
		t.Fatal(err.Error())
	}

	login := func() *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, test.CreateRequest(
			http.MethodPost,
			"/login",
			bytes.NewBufferString(`{"Email":"email@email.com","Password":"MDP!@#111111111"}`),
			"application/json",
		))
		return w
	}

	first := login()
	if first.Code != http.StatusOK {
		t.Fatalf("first login got status code = %v, want = %v", first.Code, http.StatusOK)
	}
	firstToken := extractToken(t, first.Body.String())

	out := httptest.NewRecorder()
	mux.ServeHTTP(out, logoutRequest("Bearer "+firstToken))
	if out.Code != http.StatusNoContent {
		t.Fatalf("logout got status code = %v, want = %v", out.Code, http.StatusNoContent)
	}

	second := login()
	if second.Code != http.StatusOK {
		t.Fatalf("login after logout got status code = %v, want = %v", second.Code, http.StatusOK)
	}

	// The fresh token must work; the old one must stay dead.
	secondToken := extractToken(t, second.Body.String())
	if secondToken == firstToken {
		t.Fatal("expected a new token after logging back in")
	}

	reuse := httptest.NewRecorder()
	mux.ServeHTTP(reuse, logoutRequest("Bearer "+firstToken))
	if reuse.Code != http.StatusUnauthorized {
		t.Errorf("the revoked token got status code = %v, want = %v", reuse.Code, http.StatusUnauthorized)
	}

	fresh := httptest.NewRecorder()
	mux.ServeHTTP(fresh, logoutRequest("Bearer "+secondToken))
	if fresh.Code != http.StatusNoContent {
		t.Errorf("the new token got status code = %v, want = %v", fresh.Code, http.StatusNoContent)
	}
}

func extractToken(t *testing.T, loginBody string) string {
	t.Helper()

	var response struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal([]byte(loginBody), &response); err != nil {
		t.Fatalf("could not read login response: %v", err)
	}
	if response.Token == "" {
		t.Fatal("login response carried no token")
	}
	return response.Token
}

// testLoginThrottle builds a throttle with the production budgets, generous
// enough that the tests which are not about throttling never trip it.
func testLoginThrottle() *services.LoginThrottleService {
	return services.NewLoginThrottleService(
		ratelimit.NewInMemoryAttemptStore(ratelimit.DefaultMaxEntries),
		services.Policy{Limit: 10, Window: 15 * time.Minute},
		services.Policy{Limit: 30, Window: 15 * time.Minute},
	)
}

// throttleTestServer wires a login-capable mux with a tight username budget, so
// the throttling tests trip it in a few attempts instead of ten.
func throttleTestServer(t *testing.T, usernameLimit int) *http.ServeMux {
	t.Helper()

	ctx := test.CreateTestContext(t)
	userRepo := &memory.InMemoryUserRepository{}
	tokenService := services.NewTokenService([]byte("f81d4fae-7dec-11d0-a765-00a0c91e6bf6"), 24*time.Hour)
	sessionService := services.NewSessionService(&memory.InMemoryRevokedTokenRepository{})
	throttle := services.NewLoginThrottleService(
		ratelimit.NewInMemoryAttemptStore(ratelimit.DefaultMaxEntries),
		services.Policy{Limit: usernameLimit, Window: 15 * time.Minute},
		services.Policy{Limit: 1000, Window: 15 * time.Minute},
	)
	service := services.NewUserService(userRepo, tokenService)
	controller := NewUserController(service, sessionService, throttle, interfaces.NewAuthenticationMiddleware(tokenService, sessionService))

	// bcrypt hash of "MDP!@#111111111"
	userToAdd := user.CreateUser(
		uuid.NewString(),
		"Edgouille",
		"$2a$14$/IhjU2PxRypamw1kLypfIeB28u32sgVtTL2EvCl8Ar.sUlPk77drO",
		"email@email.com",
		time.Now())
	if err := userRepo.Create(ctx, userToAdd); err != nil {
		t.Fatal(err.Error())
	}

	mux := http.NewServeMux()
	controller.AddUserHandlersToMux(mux)

	return mux
}

func attemptLogin(mux *http.ServeMux, email string, password string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, test.CreateRequest(
		http.MethodPost,
		"/login",
		bytes.NewBufferString(fmt.Sprintf(`{"Email":%q,"Password":%q}`, email, password)),
		"application/json",
	))
	return w
}

func TestUserController_login_BlocksAfterTooManyFailures(t *testing.T) {
	mux := throttleTestServer(t, 3)

	for i := 1; i <= 3; i++ {
		w := attemptLogin(mux, "email@email.com", "wrongpassword1")
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d got status code = %v, want = %v", i, w.Code, http.StatusUnauthorized)
		}
	}

	blocked := attemptLogin(mux, "email@email.com", "wrongpassword1")
	if blocked.Code != http.StatusTooManyRequests {
		t.Fatalf("the attempt past the limit got status code = %v, want = %v", blocked.Code, http.StatusTooManyRequests)
	}

	retryAfter := blocked.Header().Get("Retry-After")
	seconds, err := strconv.Atoi(retryAfter)
	if err != nil {
		t.Fatalf("Retry-After = %q, want whole seconds", retryAfter)
	}
	if seconds < 1 || seconds > 60 {
		t.Errorf("Retry-After = %d, want the first offense's one-minute block", seconds)
	}
}

// The throttle has to run before the password is checked, or it does nothing
// about the ~607ms bcrypt cost that makes login a denial-of-service vector.
func TestUserController_login_BlockedEvenWithTheCorrectPassword(t *testing.T) {
	mux := throttleTestServer(t, 3)

	for i := 0; i < 3; i++ {
		attemptLogin(mux, "email@email.com", "wrongpassword1")
	}

	w := attemptLogin(mux, "email@email.com", "MDP!@#111111111")

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("a blocked user with the right password got status code = %v, want = %v", w.Code, http.StatusTooManyRequests)
	}
}

func TestUserController_login_SuccessClearsTheBudget(t *testing.T) {
	mux := throttleTestServer(t, 3)

	// Two failures, then a success, then two more failures must not trip a
	// limit of three.
	for i := 0; i < 2; i++ {
		attemptLogin(mux, "email@email.com", "wrongpassword1")
	}

	ok := attemptLogin(mux, "email@email.com", "MDP!@#111111111")
	if ok.Code != http.StatusOK {
		t.Fatalf("login got status code = %v, want = %v", ok.Code, http.StatusOK)
	}

	for i := 1; i <= 2; i++ {
		w := attemptLogin(mux, "email@email.com", "wrongpassword1")
		if w.Code != http.StatusUnauthorized {
			t.Errorf("failure %d after a success got status code = %v, want = %v", i, w.Code, http.StatusUnauthorized)
		}
	}
}

func TestUserController_login_OtherAccountsAreUnaffected(t *testing.T) {
	mux := throttleTestServer(t, 3)

	for i := 0; i < 4; i++ {
		attemptLogin(mux, "email@email.com", "wrongpassword1")
	}

	// A different username from the same address still gets a real answer,
	// since the IP budget is nowhere near spent.
	w := attemptLogin(mux, "bystander@example.com", "wrongpassword1")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("a different account got status code = %v, want = %v", w.Code, http.StatusUnauthorized)
	}
}
