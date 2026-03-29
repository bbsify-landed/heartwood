package heartwood_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	hw "github.com/bbsify-landed/heartwood/pkg/heartwood"
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

	return httptest.NewRequest(
		r.method,
		r.target,
		b,
	)
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

func TestHTTPRequests(t *testing.T) {
	app := SimpleApp()
	ctx := t.Context()
	mu := hw.NewServeMux(app, ctx)

	testCases := []ReqTest{
		&RequestTest[*Foo, Baz, *Baz]{
			method: "POST",
			target: "/health",
			req: &Foo{
				Bar: "alice",
			},
			statusCode: 200,
			res: &Baz{
				Ble: "bob",
			},
		},
		&RequestTest[*Foo, hw.ClientError, *hw.ClientError]{
			method: "GET",
			target: "/health",
			req: &Foo{
				Bar: "alice",
			},
			statusCode: 405,
			res: &hw.ClientError{
				StatusCode: 405,
				Err:        "method not allowed",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.Name(),
			func(t *testing.T) {
				r := testCase.CreateRequest(t)
				rr := httptest.NewRecorder()

				mu.ServeHTTP(rr, r)

				res := rr.Result()
				testCase.AssertResponse(t, res)
			},
		)
	}
}
