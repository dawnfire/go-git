//go:build js
// +build js

package ioutil

import "github.com/go-git/go-git/v5/bufpipe"

func Pipe() (PipeReader, PipeWriter) {
	return bufpipe.New(nil)
}
