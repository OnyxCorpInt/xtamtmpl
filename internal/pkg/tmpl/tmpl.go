package tmpl

import (
	"fmt"
	"strings"
	"xtamtmpl/pkg/client"
)

// TemplateContext is intended to be used as the template object in a text/template
type TemplateContext struct {
	xtam         *client.RestAPI
	secrets      map[string]int
	certificates map[string]int
}

// NewContext creates a new template context that will fetch records from the given container.
func NewContext(containerID string, xtam *client.RestAPI) (*TemplateContext, error) {
	entries, err := xtam.ListContainer(containerID)
	if err != nil {
		return nil, err
	}

	ctx := &TemplateContext{
		xtam:         xtam,
		secrets:      make(map[string]int),
		certificates: make(map[string]int),
	}

	all := make(map[string]string)
	for _, entry := range entries {
		lcName := strings.ToLower(entry.Name)

		// abort if more than one record has the same name, as it is likely accidental and could lead to unpredictable builds
		if recordType, ok := all[lcName]; ok {
			return nil, fmt.Errorf("Container has repeated record with name %s (first type %s, second type %s)", entry.Name, recordType, entry.RecordType.Name)
		}
		all[lcName] = entry.RecordType.Name

		switch entry.RecordType.Name {
		case client.RecordTypeSecret:
			ctx.secrets[lcName] = entry.ID
		case client.RecordTypeCertificate:
			ctx.certificates[lcName] = entry.ID
		}
	}

	return ctx, nil
}

// CertPEM is a templates function to insert a PEM-encoded certificate
// e.g. {{.CertPEM('my-certificate')}}
func (ctx *TemplateContext) CertPEM(name string) (string, error) {
	name = strings.ToLower(name)
	id, ok := ctx.certificates[name]
	if !ok {
		return "", fmt.Errorf("certificate '%s' not found; known certificates include: %v", name, names(ctx.certificates))
	}

	return ctx.xtam.UnlockCertificate(id)
}

// Secret is a templates function to insert a secret value
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
