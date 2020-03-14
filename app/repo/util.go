package repo

import "encoding/hex"

// isgitsha reports whether s has the right format to be a git sha; that is,
// whether it is 40 hex characters.
func isgitsha(s string) bool {
	b, err := hex.DecodeString(s)
	return err == nil && len(b) == 20
}
