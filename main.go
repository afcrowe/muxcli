package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

// main.go keeps only the CLI wiring. Helpers and commands live in other files.

func usage() {
	fmt.Println("Usage: muxcli <command> [flags]\nCommands:\n  create           Create a direct upload, upload a file, and wait for asset to be ready\n  delete           Delete an asset by ID\n  get              Get asset details by ID\n  create-rendition Create a static rendition for an asset\n  list-renditions  List static renditions for an asset\n  get-master       Enable master access and print master MP4 URL (temporary)")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	cmd := os.Args[1]

	// common flags
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	key := fs.String("key-id", "", "Mux Token ID")
	secret := fs.String("secret-key", "", "Mux Token Secret")
	input := fs.String("input-file", "", "Local file to upload (for create)")
	assetID := fs.String("asset-id", "", "Asset ID for commands that require it")
	resolution := fs.String("resolution", "highest", "Resolution for static rendition (create-rendition)")
	wait := fs.Bool("wait", false, "For get-master: poll until master URL is available")
	timeout := fs.Int("timeout", 300, "Timeout in seconds when --wait is used (default 300)")
	// parse subcommand flags
	fs.Parse(os.Args[2:])

	if *key == "" || *secret == "" {
		log.Fatalf("--key-id and --secret-key are required")
	}

	var err error
	switch cmd {
	case "create":
		if *input == "" {
			log.Fatalf("create requires --input-file")
		}
		err = cmdCreate(*key, *secret, *input)
	case "delete":
		if *assetID == "" {
			log.Fatalf("delete requires --asset-id")
		}
		err = cmdDelete(*key, *secret, *assetID)
	case "get":
		if *assetID == "" {
			log.Fatalf("get requires --asset-id")
		}
		err = cmdGet(*key, *secret, *assetID)
	case "create-rendition":
		if *assetID == "" {
			log.Fatalf("create-rendition requires --asset-id")
		}
		err = cmdCreateRendition(*key, *secret, *assetID, *resolution)
	case "list-renditions":
		if *assetID == "" {
			log.Fatalf("list-renditions requires --asset-id")
		}
		err = cmdListRenditions(*key, *secret, *assetID)
	case "get-master":
		if *assetID == "" {
			log.Fatalf("get-master requires --asset-id")
		}
		err = cmdGetMasterDownload(*key, *secret, *assetID, *wait, *timeout)
	default:
		usage()
		os.Exit(2)
	}

	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
