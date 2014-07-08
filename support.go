package tablestruct

import (
	"io"
	"text/template"
)

// GenSupport generates the support Go code for all tablestruct mappers.
func GenSupport(w io.Writer, pkg string) {
	tmpl, err := template.New("support").Parse(supportTemplate)
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(w, struct{ Package string }{pkg}); err != nil {
		panic(err)
	}
}
