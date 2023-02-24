//go:build go1.20

package binary

import "unsafe"

//go:inline
func unsafeString(s []byte) string {
	return unsafe.String(&s[0], len(s))
}
