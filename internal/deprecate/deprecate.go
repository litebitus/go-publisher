// Package deprecate provides simple functions to standardize the output
// of deprecation notices on go-publisher
package deprecate

import (
	"bytes"
	"text/template"

	"github.com/caarlos0/log"
	"github.com/litebitus/go-publisher/internal/logext"
	"github.com/litebitus/go-publisher/pkg/context"
)

// Notice warns the user about the deprecation of the given property.
func Notice(ctx *context.Context, property string) {
	NoticeCustom(ctx, property, "{{ .Property }} should not be used anymore")
}

// NoticeCustom warns the user about the deprecation of the given property.
func NoticeCustom(ctx *context.Context, property, tmpl string) {
	if _, ok := ctx.NotifiedDeprecations[property]; ok {
		// already notified this one
		return
	}
	ctx.NotifiedDeprecations[property] = struct{}{}

	var out bytes.Buffer
	if err := template.
		Must(template.New("deprecation").Parse("DEPRECATED: "+tmpl)).
		Execute(&out, templateData{
			Property: logext.Keyword(property),
		}); err != nil {
		panic(err) // this should never happen
	}

	ctx.Deprecated = true
	log.Warn(logext.Warning(out.String()))
}

type templateData struct {
	Property string
}
