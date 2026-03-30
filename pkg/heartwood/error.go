package heartwood

import (
	"encoding/json"
	"fmt"
	"io"
)

type HeartwoodError struct {
	StatusCode int   `json:"status_code"`
	Err        error `json:"error"`
}

func Error(code int, err error) *HeartwoodError {
	return &HeartwoodError{StatusCode: code, Err: err}
}

func (e *HeartwoodError) Error() string {
	return e.Err.Error()
}

func (e *HeartwoodError) Unwrap() error {
	return e.Err
}

func (e *HeartwoodError) Serialize(w io.Writer) error {
	return json.NewEncoder(w).Encode(struct {
		StatusCode int    `json:"status_code"`
		Err        string `json:"error"`
	}{
		StatusCode: e.StatusCode,
		Err:        e.Error(),
	})
}

type ClientError struct {
	StatusCode int    `json:"status_code"`
	Err        string `json:"error"`
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("[%d] %s", e.StatusCode, e.Err)
}

func (e *ClientError) Deserialize(r io.Reader) error {
	if err := json.NewDecoder(r).Decode(e); err != nil {
		return err
	}
	return nil
}

func (e *ClientError) Validate() error { return nil }
