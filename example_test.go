package heartwood_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"

	hw "github.com/bbsify-landed/heartwood"
)

// GreetRequest is the request type for the greeting endpoint.
type GreetRequest struct {
	Name string `json:"name"`
}

func (r *GreetRequest) Serialize(w io.Writer) error  { return json.NewEncoder(w).Encode(r) }
func (r *GreetRequest) Deserialize(rd io.Reader) error { return json.NewDecoder(rd).Decode(r) }
func (r *GreetRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// GreetResponse is the response type for the greeting endpoint.
type GreetResponse struct {
	Message string `json:"message"`
}

func (r *GreetResponse) Serialize(w io.Writer) error    { return json.NewEncoder(w).Encode(r) }
func (r *GreetResponse) Deserialize(rd io.Reader) error { return json.NewDecoder(rd).Decode(r) }
func (r *GreetResponse) Validate() error                { return nil }

func Example() {
	app := hw.New()

	hw.Use(app, "POST", "/greet", func(ctx context.Context, req *GreetRequest) (*GreetResponse, error) {
		return &GreetResponse{Message: "Hello, " + req.Name + "!"}, nil
	})

	// Dispatch a request directly (useful for testing without HTTP).
	body, _ := json.Marshal(GreetRequest{Name: "World"})
	var out bytes.Buffer
	_ = hw.Handle(app, context.Background(), "POST", "/greet", bytes.NewReader(body), &out)

	var res GreetResponse
	_ = json.NewDecoder(&out).Decode(&res)
	fmt.Println(res.Message)
	// Output: Hello, World!
}

func Example_httpServer() {
	app := hw.New()

	hw.Use(app, "POST", "/greet", func(ctx context.Context, req *GreetRequest) (*GreetResponse, error) {
		return &GreetResponse{Message: "Hello, " + req.Name + "!"}, nil
	})

	// Build an http.ServeMux and use httptest for a full round-trip.
	mux := hw.NewServeMux(app, context.Background())
	srv := httptest.NewServer(mux)
	defer srv.Close()

	body, _ := json.Marshal(GreetRequest{Name: "Heartwood"})
	resp, err := srv.Client().Post(srv.URL+"/greet", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Println("request failed:", err)
		return
	}
	defer resp.Body.Close()

	var res GreetResponse
	_ = json.NewDecoder(resp.Body).Decode(&res)
	fmt.Println(res.Message)
	// Output: Hello, Heartwood!
}

func Example_errorHandling() {
	app := hw.New()

	hw.Use(app, "POST", "/greet", func(ctx context.Context, req *GreetRequest) (*GreetResponse, error) {
		if req.Name == "nobody" {
			return nil, hw.Error(422, errors.New("we don't greet nobody"))
		}
		return &GreetResponse{Message: "Hello, " + req.Name + "!"}, nil
	})

	body, _ := json.Marshal(GreetRequest{Name: "nobody"})
	err := hw.Handle(app, context.Background(), "POST", "/greet", bytes.NewReader(body), &bytes.Buffer{})

	var hwErr *hw.HeartwoodError
	if errors.As(err, &hwErr) {
		fmt.Printf("status %d: %s\n", hwErr.StatusCode, hwErr.Error())
	}
	// Output: status 422: we don't greet nobody
}
