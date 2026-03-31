package heartwood_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"

	hw "github.com/bbsify-landed/heartwood/pkg/heartwood"
)

type Foo struct {
	Bar string `json:"bar"`
}

func (f *Foo) Serialize(w io.Writer) error {
	e := json.NewEncoder(w)
	if err := e.Encode(f); err != nil {
		return err
	}

	return nil
}

func (f *Foo) Deserialize(r io.Reader) error {
	d := json.NewDecoder(r)
	if err := d.Decode(f); err != nil {
		return err
	}

	return nil
}

func (f *Foo) Validate() error { return nil }

type Baz struct {
	Ble string `json:"ble"`
}

func (b *Baz) Serialize(w io.Writer) error {
	e := json.NewEncoder(w)
	if err := e.Encode(b); err != nil {
		return err
	}

	return nil
}

func (b *Baz) Deserialize(r io.Reader) error {
	d := json.NewDecoder(r)
	if err := d.Decode(b); err != nil {
		return err
	}

	return nil
}

func (b *Baz) Validate() error { return nil }

func marshalReader[T any](v T) (io.Reader, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

func SimpleApp() *hw.App {
	app := hw.New()
	hw.Use(
		app,
		"POST",
		"/health",
		func(ctx context.Context, req *Foo) (*Baz, error) {
			if req.Bar != "alice" {
				return nil, hw.Error(400, errors.New("expected 'alice' in 'bar' field"))
			}

			return &Baz{
				Ble: "bob",
			}, nil
		},
	)
	return app
}
