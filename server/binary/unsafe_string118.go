//go:build !go1.20

package binary

//go:inline
func unsafeString(s []byte) string {
	return string(s)
}
