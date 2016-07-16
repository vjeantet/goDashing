package dashing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"path/filepath"
	"regexp"
	"strings"

	"github.com/GeertJohan/go.rice"
	"github.com/GeertJohan/go.rice/embedded"
	"github.com/clbanning/mxj"
	"github.com/husobee/vestigo"
	"gopkg.in/karlseguin/gerb.v0"
)

// A Server contains webservice parameters and middlewares.
type Server struct {
	dev     bool
	webroot string
	broker  *Broker
}

func param(r *http.Request, name string) string {
	// vars := mux.Vars(r)
	// value := vars[name]
	// return value
	return r.FormValue(fmt.Sprintf(":%s", name))
}

// EventsHandler opens a keepalive connection and pushes events to the client.
func (s *Server) EventsHandler(w http.ResponseWriter, r *http.Request) {

	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	c, ok := w.(http.CloseNotifier)
	if !ok {
		http.Error(w, "Close notification unsupported!", http.StatusInternalServerError)
		return
	}

	// Create a new channel, over which the broker can
	// send this client events.
	events := make(chan *Event)

	// Add this client to the map of those that should
	// receive updates
	s.broker.newClients <- events

	// Remove this client from the map of attached clients
	// when the handler exits.
	defer func() {
		s.broker.defunctClients <- events
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	closer := c.CloseNotify()

	for {
		select {
		case event := <-events:
			json, err := mxj.Map(event.Body).Json()
			if err != nil {
				continue
			}
			if event.Target != "" {
				fmt.Fprintf(w, "event: %s\n", event.Target)
			}
			fmt.Fprintf(w, "data: %s\n\n", json)
			f.Flush()
		case <-closer:
			// log.Println("Closing connection")
			return
		}
	}

}

const (
	locationBOX = iota + 1
	locationFS
)

func (s *Server) fileGetContent(path string, boxName string) (string, int, error) {
	var fileContent string
	var err error
	var box *rice.Box

	fullpath := filepath.Join(s.webroot, boxName, path)

	_, err = os.Stat(fullpath)
	if err == nil {
		var fileContentByte []byte
		fileContentByte, err = ioutil.ReadFile(fullpath)
		if err != nil {
			return "", locationFS, fmt.Errorf("Error while reading file %s: %s\n", fullpath, err.Error())
		}
		// log.Printf("(FS) %s", fullpath)
		fileContent = string(fileContentByte)
		return fileContent, locationFS, nil
	}

	switch boxName {
	case "public":
		box, err = rice.FindBox("assets/public")
	case "widgets":
		box, err = rice.FindBox("assets/widgets")
	case "dashboards":
		box, err = rice.FindBox("assets/dashboards")
	default:
		panic("Unknow BOX")
	}

	if err != nil {
		return "", locationBOX, fmt.Errorf("Box %s not found : %s\n", boxName, err.Error())
	}

	fileContent, err = box.String(path)
	if err != nil {
		return "", locationBOX, fmt.Errorf("file %s not found in box : %s\n", path, boxName, err.Error())
	}

	return fileContent, locationBOX, err
}

// DashboardEventHandler accepts dashboard events.
func (s *Server) DashboardEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	var data map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	s.broker.events <- NewEvent(param(r, "id"), data, "dashboards")

	w.WriteHeader(http.StatusNoContent)
}

// WidgetEventHandler accepts widget data.
func (s *Server) WidgetEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	var data map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("%v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	s.broker.events <- NewEvent(param(r, "id"), data, "")

	w.WriteHeader(http.StatusNoContent)
}

var camelingRegex = regexp.MustCompile("[0-9A-Za-z]+")

func CamelCase(src string) string {
	byteSrc := []byte(src)
	chunks := camelingRegex.FindAll(byteSrc, -1)
	for idx, val := range chunks {

		chunks[idx] = bytes.Title(val)

	}
	return string(bytes.Join(chunks, nil))
}

// WidgetHandler serves widget templates.
func (s *Server) WidgetHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	widget := param(r, "widget")
	widget = widget[0 : len(widget)-5]

	tplWidget, FSTYPE, err := s.fileGetContent(fmt.Sprintf("%s/%s.html", widget, widget), "widgets")

	if err != nil {
		// log.Printf("%v", err)
		widget = CamelCase(widget)
		tplWidget, FSTYPE, err = s.fileGetContent(fmt.Sprintf("%s/%s.html", widget, widget), "widgets")
		if err != nil {
			log.Printf("404 - %s - %s\n", "widgets", fmt.Sprintf("%s/%s.html", widget, widget))
		}

	}

	template, err := gerb.ParseString(true, tplWidget)

	if err != nil {
		log.Printf("500 - %s - %s\n", r.URL.Path, err.Error())
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	if FSTYPE == locationBOX {
		w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 120))
	}

	template.Render(w, nil)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// WidgetsJSHandler serves widget templates.
func (s *Server) WidgetsJSHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/javascript; charset=UTF-8")

	//Disk
	files, _ := filepath.Glob(s.webroot + "widgets/*/*.js")
	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Printf(`Error while reading "%s" [%s]`, file, err)
			continue
		}
		// log.Printf("(FS) %s", file)
		w.Write(content)
		w.Write([]byte("\n\n\n"))
	}

	//BOX
	for k, v := range embedded.EmbeddedBoxes["assets/widgets"].Files {
		if filepath.Ext(k) == ".js" && !stringInSlice(s.webroot+"widgets/"+k, files) {
			w.Write([]byte(v.Content))
			w.Write([]byte("\n\n\n"))
		}
	}
}

func (s *Server) WidgetsCSSHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=UTF-8")

	//Disk
	files, _ := filepath.Glob(s.webroot + "widgets/*/*.css")
	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Printf(`Error while reading "%s" [%s]`, file, err)
			continue
		}
		log.Printf("(FS) %s", file)
		w.Write(content)
		w.Write([]byte("\n\n\n"))
	}

	//BOX
	for k, v := range embedded.EmbeddedBoxes["assets/widgets"].Files {
		if filepath.Ext(k) == ".css" && !stringInSlice(s.webroot+"widgets/"+k, files) {
			w.Write([]byte(v.Content))
			w.Write([]byte("\n\n\n"))
		}
	}

}

func (s *Server) StaticHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	content, FSTYPE, err := s.fileGetContent(r.URL.Path[8:], "public")
	if err != nil {
		log.Printf("404 - %s - %s\n", "public", r.URL.Path[8:])
		http.NotFound(w, r)
		return
	}

	switch path.Ext(r.URL.Path) {
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=UTF-8")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript; charset=UTF-8")
	case ".ttf":
		w.Header().Set("Content-Type", "application/x-font-ttf")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	if FSTYPE == locationBOX {
		w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 120))
	}

	w.Write([]byte(content))
}

// DashboardHandler serves the dashboard layout template.
func (s *Server) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var tplDashboard string
	var tplLayout string

	dashboard := param(r, "dashboard")
	folder := param(r, "sub")

	if dashboard == "" {
		dashboard = fmt.Sprintf("events%s", param(r, "suffix"))
	}

	if folder != "" {
		folder, dashboard = dashboard, folder
		folder = folder + "/"
	}
	dashboardpath := folder + dashboard

	tplDashboard, _, err = s.fileGetContent(fmt.Sprintf("%s.gerb", dashboardpath), "dashboards")
	if err != nil {
		fileInfo, err := os.Stat(s.webroot + "dashboards/" + dashboardpath)
		if err != nil || fileInfo.IsDir() {
			http.Redirect(w, r, fmt.Sprintf("/%s/", dashboardpath), http.StatusTemporaryRedirect)
			return
		}
		log.Printf("404 - %s - %s\n", "dashboards", fmt.Sprintf("%s.gerb", dashboardpath))
		http.NotFound(w, r)
		return
	}

	tplLayout, _, err = s.fileGetContent(folder+"layout.gerb", "dashboards")
	if err != nil {
		tplLayout, _, err = s.fileGetContent(s.webroot+"layout.gerb", "dashboards")
		if err != nil {
			log.Printf("404 - %s - %s\n", "dashboards", "layout.gerb")
			http.NotFound(w, r)
			return
		}
	}

	template, err := gerb.ParseString(true, tplDashboard, tplLayout)

	if err != nil {
		log.Printf("500 - %s - %s\n", r.URL.Path, err.Error())
		http.NotFound(w, r)
		return
	}

	dashboardNames := s.getDashboardNames(dashboardpath)

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")

	template.Render(w, map[string]interface{}{
		"dashboard":      dashboard,
		"development":    s.dev,
		"request":        r,
		"dashboardnames": dashboardNames,
	})
}

func (s *Server) getDashboardNames(path string) string {
	pathBlock := strings.SplitN(path, "/", 2)
	basePath := ""
	currentdb := pathBlock[0]

	if len(pathBlock) > 1 {
		basePath = pathBlock[0] + "/"
		currentdb = pathBlock[1]
	}

	bdnames := ""

	files, _ := filepath.Glob(s.webroot + "dashboards/" + basePath + "*.gerb")
	for _, file := range files {
		name := filepath.Base(file)
		name = name[:len(name)-5]
		if name != "layout" && name != currentdb {
			bdnames = bdnames + basePath + name + " "
		}
	}

	if bdnames != "" {
		bdnames = basePath + currentdb + " " + bdnames
	}

	return bdnames
}

// IndexHandler redirects to the default dashboard.
func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	path := ""
	if dashboard := param(r, "dashboard"); dashboard != "" {
		path = path + dashboard + "/"
	}

	files, _ := filepath.Glob(s.webroot + "dashboards/" + path + "*.gerb")

	for _, file := range files {
		dashboard := file[len(s.webroot+"dashboards/"+path) : len(file)-5]
		if dashboard != "layout" {
			http.Redirect(w, r, fmt.Sprintf("/%s", path+dashboard), http.StatusTemporaryRedirect)
			return
		}
	}

	http.NotFound(w, r)
}

// NewRouter creates a router with defaults.
func (s *Server) NewRouter() http.Handler {
	r := vestigo.NewRouter()

	r.Get("/", s.IndexHandler)
	r.Get("/widgets.js", s.WidgetsJSHandler)
	r.Get("/widgets.css", s.WidgetsCSSHandler)

	r.Get("/events", s.EventsHandler)
	r.Get("/events:suffix", s.DashboardHandler) // workaround for router edge case

	r.Post("/dashboards/:id", s.DashboardEventHandler)

	r.Get("/views/:widget", s.WidgetHandler)
	r.Post("/widgets/:id", s.WidgetEventHandler)

	r.Get("/public/*", s.StaticHandler)

	r.Get("/:dashboard", s.DashboardHandler)
	r.Get("/:dashboard/", s.IndexHandler)
	r.Get("/:dashboard/:sub", s.DashboardHandler)
	return r
}

// NewServer creates a Server instance.
func NewServer(b *Broker) *Server {
	return &Server{
		dev:     false,
		webroot: "",
		broker:  b,
	}
}
