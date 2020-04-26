// Inspired from https://github.com/kelseyhightower/inspector/blob/master/logger.go

package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vasu1124/introspect/pkg/cookie"
	"github.com/vasu1124/introspect/pkg/dynconfig"
	"github.com/vasu1124/introspect/pkg/election"
	"github.com/vasu1124/introspect/pkg/environ"
	"github.com/vasu1124/introspect/pkg/guestbook"
	"github.com/vasu1124/introspect/pkg/healthz"
	"github.com/vasu1124/introspect/pkg/logger"
	"github.com/vasu1124/introspect/pkg/mandelbrot"
	"github.com/vasu1124/introspect/pkg/operator"
	"github.com/vasu1124/introspect/pkg/validate"
	"github.com/vasu1124/introspect/pkg/version"
)

func main() {

	log.Printf("[introspect] Version = %s/%s/%s", version.Version, version.Commit, version.Branch)

	// index.html
	http.HandleFunc("/", serveHTTP)

	http.Handle("/environ", logger.NewRequestLoggerHandler(environ.New()))
	log.Println("[introspect] registered /environ")
	http.Handle("/mandelbrot", logger.NewRequestLoggerHandler(mandelbrot.New()))
	log.Println("[introspect] registered /mandelbrot")
	http.Handle("/dynconfig", logger.NewRequestLoggerHandler(dynconfig.New()))
	log.Println("[introspect] registered /dynconfig")
	http.Handle("/cookie", logger.NewRequestLoggerHandler(cookie.New()))
	log.Println("[introspect] registered /cookie")
	http.Handle("/metrics", promhttp.Handler())
	log.Println("[introspect] registered /metrics")
	http.Handle("/validate", logger.NewRequestLoggerHandler(validate.New()))
	log.Println("[introspect] registered /validate")

	http.Handle("/healthz", logger.NewRequestLoggerHandler(healthz.New()))
	http.Handle("/healthzr", logger.NewRequestLoggerHandler(healthz.New()))
	log.Println("[introspect] registered /healthz|r")

	http.HandleFunc("/favicon.ico", http.NotFound)
	http.Handle("/css/", logger.NewRequestLoggerHandler(http.StripPrefix("/css/", http.FileServer(http.Dir("css")))))

	//register in background due to possible timeouts in dependant backend services
	go func() {
		http.Handle("/guestbook", logger.NewRequestLoggerHandler(guestbook.New()))
		log.Println("[introspect] registered /guestbook")

		http.Handle("/election", logger.NewRequestLoggerHandler(election.New()))
		log.Println("[introspect] registered /election")

		o := operator.New()
		if o != nil {
			http.Handle("/operator", logger.NewRequestLoggerHandler(o))
			http.HandleFunc("/operatorws", func(w http.ResponseWriter, r *http.Request) {
				o.Melody.HandleRequest(w, r)
			})
			log.Println("[introspect] registered /operator")
		}
	}()

	log.Printf("[introspect] serving HTTP  on port %d\n", *version.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *version.Port), nil))
}

func serveHTTP(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("index").Parse(`
		<!DOCTYPE html>
		<html>
			<head>
				<link rel="stylesheet" href="css/bootstrap.css">
				<style>
				{{if eq .Version "v1.0" }}
				body { background-color: #F0FFF0; }
				{{end}}
				{{if eq .Version "v2.0" }}
				body { background-color: #F0F0FF; }
				{{end}}
				</style>
			</head>
			<div class="container">
				<body>
				<h1>Introspection-{{.Version}}</h1>
				<table class="table table-bordered">
					<tr><td><a href="/environ">/environ</a></td><td>List Environment of runtime & HTTP Request</td></tr>
					<tr><td><a href="/guestbook">/guestbook</a></td><td>Guestbook Demo with MongoDB</td></tr>
					<tr><td><a href="/mandelbrot?xmin=-1.8&ymin=-1.5&xmax=1.2&ymax=1.5">/mandelbrot</a></td><td>Mandelbrot Demo, default window with xmin=-1.8 ymin=-1.5 xmax=1.2 ymax=1.5</td></tr>
					<tr><td><a href="/mandelbrot?steps=10&xfmin=-1.110&yfmin=0.228&xfmax=-1.106&yfmax=0.232">/mandelbrot</a></td><td>Mandelbrot animated gif Demo, with steps=10 and default zoom window with xfmin=-1.110 yfmin=0.228 xfmax=-1.106 yfmax=0.232</td></tr>
					<tr><td><a href="/cookie">/cookie</a></td><td>Cookie Demo</td></tr>
					<tr><td><a href="/dynconfig">/dynconfig</a></td><td>Dynamic Configuration Demo</td></tr>
					<tr><td><a href="/election">/election</a></td><td>Leadership Election</td></tr>
					<tr><td><a href="/operator">/operator</a></td><td>Useless Machine Operator Demo</td></tr>
					<tr><td><a href="/validate?ui">/validate?ui</a></td><td>ValidatingWebhook Demo</td></tr>
					<tr><td><a href="/metrics">/metrics</a></td><td>Metrics endpoint for Prometheus</td></tr>
					<tr><td><a href="/healthz">/healthz</a></td><td>Liveliness endpoint for Kubernetes</td></tr>
					<tr><td><a href="/healthzr">/healthzr</a></td><td> Readyness endpoint for Kubernetes</td></tr>
				</table>
				</body>
			</div>
			</html>
  `)
	if err != nil {
		log.Println("[introspect] parse template:", err)
		fmt.Fprint(w, "[introspect] parse template: ", err)
		return
	}

	type EnvData struct {
		Version string
	}
	data := EnvData{version.Version}

	err = t.Execute(w, data)
	if err != nil {
		log.Println("[introspect] executing template:", err)
		fmt.Fprint(w, "[introspect] executing template: ", err)
	}
}
