// Package middleware — JSON encoding helper shared by auth and audit.
package middleware

import "encoding/json"

// encodeJSON encodes v to w as JSON.
func encodeJSON(w interface{ Write([]byte) (int, error) }, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}
