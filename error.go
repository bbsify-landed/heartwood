package heartwood

import (
	"encoding/json"
	"fmt"
	"io"
)

// HeartwoodError is a server-side error that carries an HTTP status code.
// Return one from a [Handler] (via [Error]) to control the status code sent
// to the client. Unrecognized errors are mapped to 500 by [NewServeMux].
type HeartwoodError struct {
	StatusCode int   `json:"status_code"`
	Err        error `json:"error"`
}

// Error creates a [*HeartwoodError] with the given HTTP status code and
// underlying error.
func Error(code int, err error) *HeartwoodError {
	return &HeartwoodError{StatusCode: code, Err: err}
}

func (e *HeartwoodError) Error() string {
	return e.Err.Error()
}

// Unwrap returns the underlying error, supporting [errors.Is] and [errors.As].
func (e *HeartwoodError) Unwrap() error {
	return e.Err
}

// Serialize writes the error as a JSON object with status_code and error fields.
func (e *HeartwoodError) Serialize(w io.Writer) error {
	return json.NewEncoder(w).Encode(struct {
		StatusCode int    `json:"status_code"`
		Err        string `json:"error"`
	}{
		StatusCode: e.StatusCode,
		Err:        e.Error(),
	})
}

// ClientError is the client-side representation of a server error response.
// It implements [Deserializable] so it can be decoded from a JSON error body.
type ClientError struct {
	StatusCode int    `json:"status_code"`
	Err        string `json:"error"`
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("[%d] %s", e.StatusCode, e.Err)
}

// Deserialize decodes a JSON error response into the ClientError.
func (e *ClientError) Deserialize(r io.Reader) error {
	if err := json.NewDecoder(r).Decode(e); err != nil {
		return err
	}
	return nil
}

// Validate is a no-op; ClientError is always valid.
func (e *ClientError) Validate() error { return nil }
