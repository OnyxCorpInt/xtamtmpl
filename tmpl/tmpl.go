package tmpl

import (
	"strings"
	"xtamtmpl/api"
	"fmt"
)

// TemplateContext: intended to be used as the template object
type TemplateContext struct {
	xtam         *api.RestApi
	secrets      map[string]int
	certificates map[string]int
}

// NewContext creates a new template context that will fetch records from the given folder.
func NewContext(folderId string, xtam *api.RestApi) (*TemplateContext, error) {
	entries, err := xtam.ListFolder(folderId)
	if err != nil {
		return nil, err
	}

	ctx := &TemplateContext{
		xtam:         xtam,
		secrets:      make(map[string]int),
		certificates: make(map[string]int),
	}

	for _, entry := range entries {
		lcName := strings.ToLower(entry.Name)
		switch entry.RecordType.Name {
		case api.RecordTypeSecret:
			ctx.secrets[lcName] = entry.Id
		case api.RecordTypeCertificate:
			ctx.certificates[lcName] = entry.Id
		}
	}

	return ctx, nil
}

// CertPEM: intended for use inside templates to insert a PEM-encoded certificate
// e.g. {{.CertPEM('my-certificate')}}
func (ctx *TemplateContext) CertPEM(name string) (string, error) {
	name = strings.ToLower(name)
	id, ok := ctx.certificates[name]
	if !ok {
		return "", fmt.Errorf("certificate '%s' not found; known certificates include: %v", name, names(ctx.certificates))
	}

	return ctx.xtam.UnlockCertificate(id)
}

// Secret: intended for use inside templates to insert a secret value
// e.g. {{.Secret('my-second')}}
func (ctx *TemplateContext) Secret(name string) (string, error) {
	name = strings.ToLower(name)
	id, ok := ctx.secrets[name]
	if !ok {
		return "", fmt.Errorf("secret '%s' not found; known secrets include: %v", name, names(ctx.secrets))
	}

	return ctx.xtam.UnlockSecret(id)
}

func names(records map[string]int) []string {
	names := make([]string, 0, len(records))
	for name := range records {
		names = append(names, name)
	}
	return names
}

