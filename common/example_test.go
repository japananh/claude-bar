package common_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/soi/backend/pkg/common"
)

// repoGetUser is a stand-in for a real repository call.
func repoGetUser(id string) (any, error) {
	if id == "missing" {
		return nil, fmt.Errorf("repo lookup: %w", common.ErrRecordNotFound)
	}
	return nil, errors.New("simulated database failure")
}

func ExampleWriteJSON_notFound() {
	handler := common.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := repoGetUser("missing")
		if errors.Is(err, common.ErrRecordNotFound) {
			common.WriteJSON(w, r, common.ErrEntityNotFound("user", err))
			return
		}
		if err != nil {
			common.WriteJSON(w, r, common.ErrDB(err))
			return
		}
	}), common.RequestIDMiddleware)

	req := httptest.NewRequest(http.MethodGet, "/users/missing", nil)
	req.Header.Set(common.RequestIDHeader, "demo-rid")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Body contains: {"code":"ENTITY_NOT_FOUND","message":"user not found","details":{"entity":"user"},"request_id":"demo-rid"}
	fmt.Println(strings.Contains(w.Body.String(), `"code":"ENTITY_NOT_FOUND"`))
	fmt.Println(strings.Contains(w.Body.String(), `"request_id":"demo-rid"`))
	// Output:
	// 404
	// true
	// true
}

func ExampleErrValidation() {
	err := common.ErrValidation([]common.FieldError{
		{Field: "email", Message: "must be a valid email"},
		{Field: "password", Message: "must be at least 8 characters", Code: "WEAK_PASSWORD"},
	}, nil)

	fmt.Println(err.StatusCode, err.Code)
	// Output: 422 VALIDATION_FAILED
}
