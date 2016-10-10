package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kovetskiy/godocs"
	"github.com/reconquest/executil-go"
)

var (
	version = "[manual build]"
	usage   = "mirrord " + version + `


Usage:
    mirrord [options]
    mirrord -h | --help
    mirrord --version

Options:
  -d --directory <path>   Root directory to place repositories. [default: /var/mirrord/]
  -g --git-daemon <port>  Port of git daemon. [default: 9419] 
  -h --help               Show this screen.
  --version               Show version.
`
)

func main() {
	args := godocs.MustParse(usage, version, godocs.UsePager)

	var (
		root   = args["--directory"].(string)
		daemon = args["--git-daemon"].(string)
	)

	http.HandleFunc(
		"/",
		func(response http.ResponseWriter, request *http.Request) {
			module := request.Host + request.URL.Path

			err := sync(module, root)
			if err != nil {
				log.Println(err)
			}

			fmt.Fprintf(
				response,
				`<html><head><meta name="go-import" `+
					`content="%s git git://localhost:%s/%s"`+
					`/></head></html>`,
				module, daemon, module,
			)
		},
	)
}

func sync(module, root string) error {
	_, err := os.Lstat(filepath.Join(root, module, "refs/heads"))
	if os.IsNotExist(err) {
		return clone(module, root)
	}

	return update(module, root)
}

func clone(module, root string) error {
	prefixes := []string{
		"https://",
		"git+ssh://",
		"git://",
	}

	log.Printf("clone: %s", module)

	for _, prefix := range prefixes {
		_, _, err := executil.Run(
			exec.Command(
				"git", "clone", "--mirror",
				prefix+module, filepath.Join(root, module),
			),
		)
		if err != nil {
			log.Println(err)
			continue
		}

		return nil
	}

	return fmt.Errorf("can't clone repository: %s", module)
}

func update(module, root string) error {
	cmd := exec.Command("git", "remote", "update")
	cmd.Dir = filepath.Join(root, module)

	log.Printf("sync: %s", module)

	_, _, err := executil.Run(cmd)
	if err != nil {
		return fmt.Errorf("can't update repository: %s", module)
	}

	return nil
}
