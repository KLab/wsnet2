//go:build go1.20

package binary

import "unsafe"

//go:inline
func unsafeString(s []byte) string {
	if len(s) == 0 {
		return ""
	}
	return unsafe.String(&s[0], len(s))
}
