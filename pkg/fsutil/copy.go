package fsutil

import "github.com/otiai10/copy"

func CopyDir(src, dest string) error {
	return copy.Copy(src, dest)
}
