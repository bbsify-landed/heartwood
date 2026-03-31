package basic_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	hw "github.com/bbsify-landed/heartwood/pkg/heartwood"

	"github.com/bbsify-landed/heartwood/cmd/hwgen/testdata/basic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type RequestTest[R hw.Serializable, T any, D hw.Deserializable[T]] struct {
	method     string
	target     string
	req        R
	statusCode int
	res        D
}

func (r *RequestTest[R, T, D]) Name() string {
	return fmt.Sprintf("%d [%s] %s", r.statusCode, r.method, r.target)
}

func (r *RequestTest[R, T, D]) CreateRequest(t *testing.T) *http.Request {
	b := bytes.NewBuffer(nil)
	err := r.req.Serialize(b)
	require.Nil(t, err, err)

	return httptest.NewRequest(r.method, r.target, b)
}

func (r *RequestTest[R, T, D]) AssertResponse(t *testing.T, res *http.Response) {
	assert.Equal(t, r.statusCode, res.StatusCode)

	resVal := D(new(T))
	err := resVal.Deserialize(res.Body)
	require.Nil(t, err, err)
	assert.Equal(t, r.res, resVal)
}

type ReqTest interface {
	Name() string
	CreateRequest(*testing.T) *http.Request
	AssertResponse(*testing.T, *http.Response)
}

func testApp() *hw.App {
	app := hw.New()

	basic.RegisterHealthCheck(app, func(ctx context.Context, req *basic.HealthCheckRequest) (*basic.HealthCheckResponse, error) {
		if req.AreYouHealthy != "are you healthy?" {
			return nil, hw.Error(400, errors.New("expected 'are you healthy?' in 'are_you_healthy' field"))
		}
		return &basic.HealthCheckResponse{
			Healthy: "healthy",
		}, nil
	})

	basic.RegisterCreateUser(app, func(ctx context.Context, req *basic.CreateUserRequest) (*basic.CreateUserResponse, error) {
		return &basic.CreateUserResponse{
			Id:    "usr_123",
			Name:  req.Name,
			Email: req.Email,
			Age:   req.Age,
		}, nil
	})

	return app
}

func TestHTTPRequests(t *testing.T) {
	app := testApp()
	ctx := t.Context()
	mu := hw.NewServeMux(app, ctx)

	testCases := []ReqTest{
		// Successful health check
		&RequestTest[*basic.HealthCheckRequest, basic.HealthCheckResponse, *basic.HealthCheckResponse]{
			method: "POST",
			target: "/health",
			req: &basic.HealthCheckRequest{
				AreYouHealthy: "are you healthy?",
			},
			statusCode: 200,
			res: &basic.HealthCheckResponse{
				Healthy: "healthy",
			},
		},
		// Wrong method
		&RequestTest[*basic.HealthCheckRequest, hw.ClientError, *hw.ClientError]{
			method: "GET",
			target: "/health",
			req: &basic.HealthCheckRequest{
				AreYouHealthy: "are you healthy?",
			},
			statusCode: 405,
			res: &hw.ClientError{
				StatusCode: 405,
				Err:        "method not allowed",
			},
		},
		// Validation failure — empty required field
		&RequestTest[*basic.HealthCheckRequest, hw.ClientError, *hw.ClientError]{
			method: "POST",
			target: "/health",
			req:    &basic.HealthCheckRequest{},
			statusCode: 400,
			res: &hw.ClientError{
				StatusCode: 400,
				Err:        "validation failed: are_you_healthy is required",
			},
		},
		// Successful user creation
		&RequestTest[*basic.CreateUserRequest, basic.CreateUserResponse, *basic.CreateUserResponse]{
			method: "POST",
			target: "/users",
			req: &basic.CreateUserRequest{
				Name:  "Alice",
				Email: "alice@example.com",
				Age:   30,
			},
			statusCode: 200,
			res: &basic.CreateUserResponse{
				Id:    "usr_123",
				Name:  "Alice",
				Email: "alice@example.com",
				Age:   30,
			},
		},
		// Validation failure — missing name and email
		&RequestTest[*basic.CreateUserRequest, hw.ClientError, *hw.ClientError]{
			method: "POST",
			target: "/users",
			req:    &basic.CreateUserRequest{},
			statusCode: 400,
			res: &hw.ClientError{
				StatusCode: 400,
				Err:        "validation failed: name is required; name must be at least 1 characters; email is required",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name(), func(t *testing.T) {
			r := testCase.CreateRequest(t)
			rr := httptest.NewRecorder()
			mu.ServeHTTP(rr, r)
			res := rr.Result()
			testCase.AssertResponse(t, res)
		})
	}
}
