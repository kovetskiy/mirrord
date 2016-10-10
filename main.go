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
  -l --listen <addr>      Listen specified address for HTTP connections. [default: :80]
  -s --listen-ssl <addr>  Listen specified address for HTTPS connections. [default: :443]
  -g --git-daemon <port>  Address of git daemon. [default: localhost:9418] 
  -k --key <path>         SSL certificate public part. [default: /etc/mirrord/ssl.key]
  -c --crt <path>         SSL certificate private part. [default: /etc/mirrord/ssl.crt]
  -h --help               Show this screen.
  --version               Show version.
`
)

func main() {
	args := godocs.MustParse(usage, version, godocs.UsePager)

	var (
		root      = args["--directory"].(string)
		listen    = (args["--listen"].(string))
		listenSSL = (args["--listen-ssl"].(string))
		daemon    = args["--git-daemon"].(string)
		sslKey    = args["--key"].(string)
		sslCrt    = args["--crt"].(string)
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
					`content="%s git git://%s/%s"`+
					`/></head></html>`,
				module, daemon, module,
			)
		},
	)

	go func() {
		err := http.ListenAndServeTLS(listenSSL, sslCrt, sslKey, nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		err := http.ListenAndServe(listen, nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("listen on %s and %s", listen, listenSSL)

	select {}
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
		cmd := exec.Command(
			"git", "clone", "--mirror", "--quiet",
			prefix+module, filepath.Join(root, module),
		)

		_, _, err := executil.Run(cmd)
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
