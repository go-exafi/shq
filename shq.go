package shq

import "fmt"

type Arg []byte

// return the arg as an unescaped string as it is expected to be
// represented in bash.  This is different from the original strings
// if the original string had a NUL in it.
func (a Arg) Unescaped() string {
	for i := 0; i < len(a); i++ {
		if a[i] == 0 {
			return string(a[:i])
		}
	}
	return string(a)
}

// return true if the arg can be represented in bash.
//
// returns false if there is a NUL in the string anywhere.
func (a Arg) Valid() bool {
	for i := 0; i < len(a); i++ {
		if a[i] == 0 {
			return false
		}
	}
	return true
}

// return the sh escaped string.
func (a Arg) String() string {
	retlen := 2 // start with the enclosing single quotes
	for i := 0; i < len(a); i++ {
		if a[i] == '\'' {
			// every time we see a single quote, we must wrap it in '"___"'
			// making '"'"' which will give us 'isn'"'"'t the weather lovely?'
			// this creates three strings which sh concatenates:
			// "isn"  "'"  "t the weather lovely?"
			retlen += 5
		} else if a[i] == 0 {
			break // bash strings stop at NUL
		} else {
			retlen += 1
		}
	}
	ret := make([]byte, retlen)
	retptr := ret
	retptr[0] = '\''
	retptr = retptr[1:]
	for i := 0; i < len(a); i++ {
		if a[i] == '\'' {
			retptr[0] = '\''
			retptr[1] = '"'
			retptr[2] = '\''
			retptr[3] = '"'
			retptr[4] = '\''
			retptr = retptr[5:]
		} else if a[i] == 0 {
			break
		} else {
			retptr[0] = a[i]
			retptr = retptr[1:]
		}
	}
	retptr[0] = '\''
	return string(ret)
}

// return a string which describes the input and escaped string
// for use with `fmt`'s `%#v`.
func (a Arg) GoString() string {
	return fmt.Sprintf("Arg(%#v -> %#v)", string(a), a.String())
}
