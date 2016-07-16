package dashing

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/GeertJohan/go.rice"
)

// An Event contains the widget ID, a body of data,
// and an optional target (only "dashboard" for now).
type Event struct {
	ID     string
	Body   map[string]interface{}
	Target string
}

func NewEvent(id string, data map[string]interface{}, target string) *Event {
	data["id"] = id
	data["updatedAt"] = int32(time.Now().Unix())
	return &Event{
		ID:     id,
		Body:   data,
		Target: target,
	}
}

// Dashing struct definition.
type Dashing struct {
	started bool
	Broker  *Broker
	Worker  *Worker
	Server  *Server
	Router  http.Handler
}

// ServeHTTP implements the HTTP Handler.
func (d *Dashing) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !d.started {
		panic("dashing.Start() has not been called")
	}
	d.Router.ServeHTTP(w, r)
}

// Start actives the broker and workers.
func (d *Dashing) Start() *Dashing {
	if !d.started {
		d.initFolders()

		if d.Router == nil {
			d.Router = d.Server.NewRouter()
		}
		d.Broker.Start()
		d.Worker.Start()
		d.started = true
	}
	return d
}

func (d *Dashing) initFolders() {

	// Si dashboards n'existe pas
	if ok, _ := exists(d.Worker.webroot + "dashboards"); !ok {
		os.MkdirAll(d.Worker.webroot+"dashboards", 0777)
		d1, _ := rice.MustFindBox("assets/dashboards").String("sample.gerb")
		ioutil.WriteFile(d.Worker.webroot+"dashboards"+string(filepath.Separator)+"sample.gerb", []byte(d1), 0644)
		d2, _ := rice.MustFindBox("assets/dashboards").String("layout.gerb")
		ioutil.WriteFile(d.Worker.webroot+"dashboards"+string(filepath.Separator)+"layout.gerb", []byte(d2), 0644)
	}

	// Si jobs n'existe pas
	if ok, _ := exists(d.Worker.webroot + "jobs"); !ok {
		os.MkdirAll(d.Worker.webroot+"jobs", 0777)
		jobbox := rice.MustFindBox("assets/jobs")
		jobbox.Walk("", func(path string, f os.FileInfo, err error) error {
			content, _ := jobbox.String(path)
			ioutil.WriteFile(d.Worker.webroot+"jobs"+string(filepath.Separator)+filepath.Base(path), []byte(content), 0644)
			return nil
		})
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// NewDashing sets up the event broker, workers and webservice.
func NewDashing(root string, port string, token string) *Dashing {
	broker := NewBroker()
	worker := NewWorker(broker)
	server := NewServer(broker)

	server.webroot = root
	worker.webroot = root
	worker.url = "http://127.0.0.1:" + port
	worker.token = token

	if os.Getenv("DEV") != "" {
		server.dev = true
	}

	server.dev = true
	return &Dashing{
		started: false,
		Broker:  broker,
		Worker:  worker,
		Server:  server,
	}
}
