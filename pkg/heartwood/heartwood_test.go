package heartwood_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	hw "github.com/bbsify-landed/heartwood/pkg/heartwood"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimple(t *testing.T) {
	app := SimpleApp()

	err, r := marshalReader(&Foo{Bar: "alice"})
	require.Nil(t, err, err)

	w := bytes.NewBuffer(nil)

	ctx := t.Context()
	err = hw.Handle(app, ctx, "POST", "/health", r, w)
	require.Nil(t, err, err)

	d := json.NewDecoder(w)
	var b Baz
	err = d.Decode(&b)
	require.Nil(t, err, err)
	assert.Equal(t, b.Ble, "bob")
}

func TestSimpleError(t *testing.T) {
	app := SimpleApp()

	err, r := marshalReader(&Foo{Bar: "bob"})
	require.Nil(t, err, err)

	w := bytes.NewBuffer(nil)

	ctx := t.Context()
	err = hw.Handle(app, ctx, "POST", "/health", r, w)

	var hwError *hw.HeartwoodError
	require.True(t, errors.As(err, &hwError))
	assert.Equal(t, hwError.StatusCode, 400)
}
