package static

import "embed"

//go:embed *.html fav templates
var Files embed.FS
