package main

import (
	"fmt"
	"github.com/namsral/flag"
	"os"
	"xtamtmpl/api"
)

func main() {
	xtamUrl := flag.String("xtam-host", "", "XTAM Base URL (required)")
	xtamCasUrl := flag.String("xtam-cas-host", "", "XTAM CAS URL (required)")
	xtamUsername := flag.String("xtam-username", "", "XTAM authentication string (required)")
	xtamPassword := flag.String("xtam-password", "", "XTAM authentication string (required)")
	xtamFolderId := flag.String("xtam-folder-id", "", "XTAM folder ID (required)")
	templatePath := flag.String("template-path", "/mnt/templates", "Directory from which to read templates")
	outputPath := flag.String("output-path", "/etc/config", "Directory to which filled templates will be written")

	flag.Parse()
	requireCliFlags("xtam-username", "xtam-password", "xtam-folder-id", "xtam-host", "xtam-cas-host")

	fmt.Printf("Auth: %s\nXTAM Folder: %s\nRead from: %s\nWrite to: %s\n", *xtamUsername, *xtamFolderId, *templatePath, *outputPath)

	xtamClient := api.RestApi{
		URL: *xtamUrl,
		Authenticator: &api.CasAuth{
			BaseURL:  *xtamUrl,
			CasURL:   *xtamCasUrl,
			User:     *xtamUsername,
			Password: *xtamPassword,
		},
	}

	folders, err := xtamClient.ListFolder(*xtamFolderId)
	abortOnError("Failed to fetch contents of folder with ID "+*xtamFolderId, err)

	fmt.Printf("\n\n%v", folders)
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

func abortWithCause(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
	os.Exit(3)
}
