package environ

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/vasu1124/introspect/pkg/network"
	"github.com/vasu1124/introspect/pkg/osinfo"
	"github.com/vasu1124/introspect/pkg/version"

	"github.com/prometheus/client_golang/prometheus"
)

var count uint32

func init() {
	// Register the summary and the histogram with Prometheus's default registry.
	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(requestDuration)
}

//Handler .
type Handler struct{}

// New .
func New() *Handler {
	var h Handler
	return &h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	atomic.AddUint32(&count, 1)

	start := time.Now()
	serveEnviron(w, r)
	duration := time.Since(start).Seconds() * 1e3 //time.Now().Sub(start).Seconds() * 1e3

	proto := strconv.Itoa(r.ProtoMajor)
	proto = proto + "." + strconv.Itoa(r.ProtoMinor)

	requestCount.WithLabelValues(proto).Inc()
	requestDuration.WithLabelValues(proto).Observe(duration)

}

func serveEnviron(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("tmpl/environ.html")
	if err != nil {
		log.Println("[environ] template parse error: ", err)
		return
	}
	err = r.ParseForm()
	if err != nil {
		log.Println("[environ] ParseForm error:", err)
	}

	type EnvData struct {
		Version     string
		Flag        bool
		Environment map[string]string
		Header      map[string][]string
		Form        map[string][]string
		Request     map[string][]string
		Process     map[string]string
		OS          map[string]string
		Server      map[string]string
		Counter     uint32
		Network     network.Data
	}

	envMap := make(map[string]string)
	for _, env := range os.Environ() {
		kv := strings.Split(env, "=")
		envMap[kv[0]] = kv[1]
	}

	requestMap := make(map[string][]string)
	{
		requestMap["ContentLength"] = []string{fmt.Sprintf("%d", r.ContentLength)}
		requestMap["Host"] = []string{r.Host}
		requestMap["RemoteAddr"] = []string{r.RemoteAddr}
		//	requestMap["URL"] = []string{r.URL.String()}
		requestMap["Proto"] = []string{r.Proto}
		requestMap["Method"] = []string{r.Method}
		requestMap["Referer"] = []string{r.Referer()}
		requestMap["RequestURI"] = []string{r.RequestURI}
		requestMap["TransferEncoding"] = r.TransferEncoding
	}

	processMap := make(map[string]string)
	{
		processMap["GO ARCH"] = runtime.GOARCH
		processMap["GO OS"] = runtime.GOOS
		processMap["GO Version"] = runtime.Version()
		processMap["GO NumCPU"] = fmt.Sprintf("%d", runtime.NumCPU())
		processMap["GO NumGoroutine"] = fmt.Sprintf("%d", runtime.NumGoroutine())
		processMap["Introspect Version"] = version.Get().GitVersion
		processMap["Introspect TreeState"] = version.Get().GitTreeState
		processMap["Introspect Commit"] = version.Get().GitCommit
		processMap["Introspect BuildDate"] = version.Get().BuildDate
	}

	serverMap := make(map[string]string)
	{
		serverMap["Machine Architecture"] = osinfo.Utsname.Machine
		serverMap["Nodename"] = osinfo.Utsname.Nodename

	}

	data := EnvData{version.Get().GitVersion, version.GetPatchVersion()%2 == 0, envMap, r.Header, r.Form, requestMap, processMap, osinfo.OSrelease, serverMap, count, network.NetworkData}

	err = t.Execute(w, data)
	if err != nil {
		log.Println("[environ] executing template: ", err)
		fmt.Fprint(w, "[environ] executing template: ", err)
	}

}
