//go:build !dev

package web

import "embed"

//go:embed dist
var Files embed.FS
