package command

import (
	"bytes"
	"fmt"
	"io"
	"text/template"

	"github.com/lunarway/release-manager/internal/intent"
)

// templateOutput parses the template text as a Go template. The empty interface
// is available as the root object in the template.
//
// Some utility functions are available for data manipulation.
func templateOutput(destination io.Writer, name, text string, data interface{}) error {
	t := template.New(name)
	t.Funcs(template.FuncMap{
		"rightPad":    tmplRightPad,
		"printIntent": tmplPrintIntent,
	})
	t, err := t.Parse(text)
	if err != nil {
		return fmt.Errorf("invalid template: %v", err)
	}
	return t.Execute(destination, data)
}

func templateToString(name, text string, data interface{}) (string, error) {
	buf := bytes.NewBuffer(nil)
	err := templateOutput(buf, name, text, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func tmplRightPad(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(template, s)
}

func tmplPrintIntent(i intent.Intent) string {
	switch i.Type {
	case intent.TypeAutoRelease:
		return "autorelease"
	case intent.TypePromote:
		return fmt.Sprintf("promotion from %s", i.Promote.FromEnvironment)
	case intent.TypeReleaseArtifact:
		return "artifact release"
	case intent.TypeReleaseBranch:
		return fmt.Sprintf("%s branch release", i.ReleaseBranch.Branch)
	case intent.TypeRollback:
		return fmt.Sprintf("rollback of %s", i.Rollback.PreviousArtifactID)
	default:
		return fmt.Sprintf("unknown intent type '%s'", i.Type)
	}
}
