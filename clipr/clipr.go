package clipr

import (
	"fmt"
	"net/http"
)

type IndexHandler struct {
	Addr string
}

type FileHander struct {
	Path string
}

func (h IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, responseBody, h.Addr, h.Addr)
}

func (h FileHander) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.Path)
}

func Configure(server *http.Server, addr, osxPath, linux64Path string) {
	mux := http.NewServeMux()
	mux.Handle("/list", IndexHandler{Addr: addr})
	mux.Handle("/bin/osx/echo", FileHander{Path: osxPath})
	mux.Handle("/bin/linux64/echo", FileHander{Path: linux64Path})
	server.Handler = mux
}

var responseBody = `{"plugins": [
  {
    "name":"echo",
    "description":"echo repeats input back to the terminal",
    "version":"0.1.4",
    "date":"0001-01-01T00:00:00Z",
    "company":"",
    "author":"",
    "contact":"feedback@email.com",
    "homepage":"https://github.com/johndoe/plugin-repo",
    "binaries": [
      {
        "platform":"osx",
        "url":"%s/bin/osx/echo",
        "checksum":"ce4964c30c3eecdf5f218ae2a7572304602f23cb"
      },
      {
        "platform":"linux64",
        "url":"%s/bin/linux64/echo",
        "checksum":"434542420336614e23b2ed91a5aab87c6325d433"
      },
      {
        "platform":"win64",
        "url":"%s/bin/windows64/echo.exe",
        "checksum":"3062d690bc2991b93c29b823771c19257a7f42f5"
      }
    ]
  }
]}`
