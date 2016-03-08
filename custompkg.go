package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"text/template"
)

var (
	importPrefix string
	pathPrefix   string
	vcsPrefix    string
	vcs          string
	port         string
)

func init() {
	flag.StringVar(&importPrefix, "import", "jhaven.me/go/", "Import prefix")
	flag.StringVar(&vcsPrefix, "vcs-prefix", "https://github.com/jacobhaven/", "VCS prefix")
	flag.StringVar(&vcs, "vcs", "git", "VCS")
	flag.StringVar(&port, "port", "8880", "")
}

var tmpl = template.Must(template.New("main").Parse(`<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta name="go-import" content="{{.ImportPrefix}}{{.PkgRoot}} {{.VCS}} {{.VCSPrefix}}{{.PkgRoot}}">
<meta http-equiv="refresh" content="0; url=https://godoc.org/{{.ImportPrefix}}{{.PkgRoot}}{{.Suffix}}">
</head>
</html>
`))

type data struct {
	ImportPrefix string
	VCSPrefix    string
	VCS          string
	PkgRoot      string
	Suffix       string
}

func servePkg(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Serving %s\n", r.URL.String())
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	pkg := strings.TrimPrefix(r.URL.Path, pathPrefix)
	if pkg == r.URL.Path || pkg == "" {
		http.NotFound(w, r)
		return
	}
	log.Printf("Serving traffic for pkg: %s\n", pkg)
	splitPkg := strings.SplitN(pkg, "/", 2)
	d := data{
		ImportPrefix: importPrefix,
		VCSPrefix:    vcsPrefix,
		VCS:          vcs,
		PkgRoot:      splitPkg[0],
	}
	if len(splitPkg) > 1 {
		d.Suffix = "/" + splitPkg[1]
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(buf.Bytes())
}

func main() {
	flag.Parse()
	prefix := strings.SplitN(importPrefix, "/", 2)
	if len(prefix) == 2 {
		pathPrefix = "/" + prefix[1]
	}
	http.HandleFunc(pathPrefix, servePkg)

	addr := net.JoinHostPort("", port)
	log.Printf("Listening at %s%s", addr, pathPrefix)
	log.Fatal(http.ListenAndServe(addr, nil))
}
