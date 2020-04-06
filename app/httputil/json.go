package httputil

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/golang/gddo/httputil/header"
)

type JSONDecoder struct {
	MaxRequestSize int64
}

func (j *JSONDecoder) DecodeRequest(w http.ResponseWriter, r *http.Request, v interface{}) error {
	// Check content type.
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			return errors.New("Content-Type header is not application/json")
		}
	}

	// Impose limit on request size.
	body := http.MaxBytesReader(w, r.Body, j.MaxRequestSize)

	// Decode JSON.
	d := json.NewDecoder(body)
	d.DisallowUnknownFields()

	if err := d.Decode(v); err != nil {
		return err
	}

	// Should not have trailing data.
	if d.More() {
		return errors.New("unexpected extra data after JSON")
	}

	return nil
}

type JSONEncoder struct {
	Debug bool
}

func (j *JSONEncoder) EncodeResponse(w http.ResponseWriter, v interface{}) error {
	// Set content type header.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Build encoder. Enable indentation in debug mode.
	e := json.NewEncoder(w)

	if j.Debug {
		e.SetIndent("", "\t")
	}

	return e.Encode(v)
}
