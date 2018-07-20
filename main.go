package main

import (
	"fmt"
	"github.com/namsral/flag"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"
	"xtamtmpl/api"
	"xtamtmpl/tmpl"
)

func main() {
	xtamURL := flag.String("xtam-host", "", "XTAM Base URL (required)")
	xtamCASURL := flag.String("xtam-cas-host", "", "XTAM CAS URL (required)")
	xtamUsername := flag.String("xtam-username", "", "XTAM authentication string (required)")
	xtamPassword := flag.String("xtam-password", "", "XTAM authentication string (required)")
	xtamFolderID := flag.String("xtam-folder-id", "", "XTAM folder ID (required)")
	templatePath := flag.String("template-path", "/mnt/templates", "Directory from which to read templates")
	flag.String(flag.DefaultConfigFlagname, "", "path to config file")
	outputPath := flag.String("output-path", "/etc/config", "Directory to which filled templates will be written")

	flag.Parse()
	requireCliFlags("xtam-username", "xtam-password", "xtam-folder-id", "xtam-host", "xtam-cas-host")

	abortUnlessDirectory(*templatePath)
	abortUnlessDirectory(*outputPath)

	fmt.Printf("Auth: %s\nXTAM Folder: %s\nRead from: %s\nWrite to: %s\n", *xtamUsername, *xtamFolderID, *templatePath, *outputPath)

	xtamClient := &api.RestApi{
		URL: *xtamURL,
		Authenticator: &api.CasAuth{
			BaseURL:  *xtamURL,
			CasURL:   *xtamCASURL,
			User:     *xtamUsername,
			Password: *xtamPassword,
		},
	}

	tmplCtx, err := tmpl.NewContext(*xtamFolderID, xtamClient)
	abortOnError("unable to create template context", err)

	tmplFiles, err := ioutil.ReadDir(*templatePath)
	abortOnError("cannot read templates", err)

	for _, tmplFile := range tmplFiles {
		if strings.HasSuffix(tmplFile.Name(), ".template") {
			targetName := strings.TrimSuffix(tmplFile.Name(), ".template")
			targetPath := path.Join(*outputPath, targetName)

			tmplContent, err := ioutil.ReadFile(path.Join(*templatePath, tmplFile.Name()))
			abortOnError("failed to read template: "+tmplFile.Name(), err)

			parsedTmpl, err := template.New(tmplFile.Name()).Parse(string(tmplContent))
			abortOnError("failed to parse template "+tmplFile.Name(), err)

			outputFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
			abortOnError("cannot open output file for writing", err)
			println(outputFile.Name())

			err = parsedTmpl.Execute(outputFile, tmplCtx)
			abortOnError("failed to process template: "+tmplFile.Name(), err)
			outputFile.Close()
		}
	}
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

func abortOnError(reason string, err error) {
	if err != nil {
		abortWithCause("Error: %s: %s", reason, err.Error())
	}
}

func abortUnlessDirectory(path string) {
	fi, err := os.Stat(path)
	abortOnError("cannot read template directory", err)
	if !fi.IsDir() {
		abortWithCause("%s must be a directory", fi.Name())
	}
}

func abortWithCause(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
	fmt.Println()
	os.Exit(3)
}
