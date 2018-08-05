package xtamtmpl

import (
	"fmt"
	"github.com/namsral/flag"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"
	"xtamtmpl/pkg/client"
	"xtamtmpl/internal/pkg/tmpl"
)

// RunCLI parses CLI flags and fills templates. Application will exit on error.
func RunCLI() {
	flag.String(flag.DefaultConfigFlagname, "", "path to config file")
	outputPath := flag.String("output-path", "/etc/config", "Directory to which filled templates will be written")
	templatePath := flag.String("template-path", "/mnt/templates", "Directory from which to read templates")
	xtamCASURL := flag.String("xtam-cas-host", "", "XTAM CAS URL (required)")
	xtamURL := flag.String("xtam-host", "", "XTAM Base URL (required)")
	xtamUsername := flag.String("xtam-username", "", "XTAM authentication string (required)")
	xtamPassword := flag.String("xtam-password", "", "XTAM authentication string (required)")
	xtamFolderID := flag.String("xtam-folder-id", "", "XTAM folder ID (required)")

	flag.Parse()

	requireCliFlags("xtam-username", "xtam-password", "xtam-folder-id", "xtam-host", "xtam-cas-host")
	requireDir(*templatePath)
	requireDir(*outputPath)

	fmt.Printf("Auth: %s\nXTAM Folder: %s\nRead from: %s\nWrite to: %s\n", *xtamUsername, *xtamFolderID, *templatePath, *outputPath)

	parsedTmpls := mustParseTemplates(*templatePath)
	if len(parsedTmpls) == 0 {
		abortWithCause("no templates found in %s, aborting", *templatePath)
	}

	xtamClient := &client.RestAPI{
		URL: *xtamURL,
		Authenticator: &client.CASAuth{
			BaseURL:  *xtamURL,
			CASURL:   *xtamCASURL,
			User:     *xtamUsername,
			Password: *xtamPassword,
		},
	}

	tmplCtx, err := tmpl.NewContext(*xtamFolderID, xtamClient)
	abortOnError("unable to create template context", err)

	for _, parsedTmpl := range parsedTmpls {
		targetPath := path.Join(*outputPath, parsedTmpl.Name())

		outputFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
		abortOnError("cannot open output file for writing", err)
		println(outputFile.Name())

		err = parsedTmpl.Execute(outputFile, tmplCtx)
		abortOnError("failed to process template: "+parsedTmpl.Name(), err)
		outputFile.Close()
	}
}

func mustParseTemplates(dir string) []*template.Template {
	tmplFiles, err := ioutil.ReadDir(dir)
	abortOnError("cannot read templates", err)

	var templates []*template.Template
	for _, tmplFile := range tmplFiles {
		if strings.HasSuffix(tmplFile.Name(), ".template") {
			targetName := strings.TrimSuffix(tmplFile.Name(), ".template")

			tmplContent, err := ioutil.ReadFile(path.Join(dir, tmplFile.Name()))
			abortOnError("failed to read template: "+tmplFile.Name(), err)

			parsedTmpl, err := template.New(targetName).Parse(string(tmplContent))
			abortOnError("failed to parse template "+tmplFile.Name(), err)

			templates = append(templates, parsedTmpl)
		}
	}

	return templates
}

func requireCliFlags(names ...string) {
	for _, name := range names {
		f := flag.CommandLine.Lookup(name)
		if f.Value.String() == f.DefValue {
			fmt.Println("missing required flag:", f.Name)
			flag.Usage()
			os.Exit(2)
		}
	}
}

func requireDir(path string) {
	fi, err := os.Stat(path)
	abortOnError("cannot read template directory", err)
	if !fi.IsDir() {
		abortWithCause("%s must be a directory", fi.Name())
	}
}

func abortOnError(reason string, err error) {
	if err != nil {
		abortWithCause("Error: %s: %s", reason, err.Error())
	}
}

func abortWithCause(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
	fmt.Println()
	os.Exit(3)
}
