package jobs

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/fsnotify.v1"

	"github.com/carlescere/scheduler"
	"github.com/streamrail/concurrent-map"
	"github.com/vjeantet/goDash"
)

type execJob struct {
	tasks cmap.ConcurrentMap
	send  chan *dashing.Event
	url   string
	token string
}

type task struct {
	name     string
	executor int
	interval int
	path     string
	widgetID string
	job      *scheduler.Job
	url      string
	token    string
}

const (
	XPHP = iota + 1
	XBIN
)

func (j *execJob) Work(send chan *dashing.Event, webroot string, url string, token string) {
	j.tasks = cmap.New()
	j.send = send
	j.url = url
	j.token = token
	j.readDir(webroot + "jobs/")

	j.watchChanges(webroot + "jobs/")
}

func (j *execJob) readDir(jobspath string) {
	files, _ := filepath.Glob(jobspath + "*")

	// Add new tasks
	for _, file := range files {

		var filename = filepath.Base(file)
		var extension = filepath.Ext(filename)
		var name = filename[0 : len(filename)-len(extension)]

		//Is file ?
		fileInfo, err := os.Stat(file)
		if err != nil || fileInfo.IsDir() {
			continue
		}

		// Already registered
		if j.tasks.Has(filename) {
			continue
		}

		s := strings.SplitN(name, "_", 2)
		if len(s) != 2 {
			log.Printf("ExecJob : ignoring file jobs/%s", name)
			continue
		}

		interval, err := strconv.Atoi(s[0])
		if err != nil {
			interval = 300
		}

		t := &task{}
		t.path = file
		t.interval = interval
		t.widgetID = s[1]
		t.name = filename
		t.url = j.url
		t.token = j.token
		switch extension {
		case ".php":
			t.executor = XPHP
		default:
			t.executor = XBIN
		}

		j.tasks.Set(t.name, t)
		t.start(j.send)
	}
}

func (t *task) start(send chan *dashing.Event) {

	var err error
	t.job, err = scheduler.Every(t.interval).Seconds().Run(func() {

		var command string
		var args []string

		switch t.executor {
		case XPHP:
			command = "php"
			args = append(args, t.path)
		case XBIN:
			command = t.path
		}

		args = append(args, t.url)
		args = append(args, t.token)

		data, err := doExec(command, args...)
		if err != nil {
			log.Printf("ExecJob - %s - error executing %s %s : %s", command, args, err.Error(), data)
			return
		}

		var j map[string]interface{}
		if err := json.Unmarshal(data, &j); err != nil {
			log.Printf("ExecJob - %s - output error  %s - '%s'", t.name, err.Error(), data)
			return
		}

		send <- dashing.NewEvent(t.widgetID, j, "")

		//log.Printf("JOB - %s - run - %s", t.name, data)
	})

	if err != nil {
		log.Printf("ExecJob - %s - scheduler error %s : %s", t.name, t.path, err.Error)
		return
	}
	log.Printf("ExecJob - %s - scheduled every %ds", t.name, t.interval)
}

func doExec(command string, args ...string) (data []byte, err error) {
	var (
		buferr bytes.Buffer
		raw    []byte
		cmd    *exec.Cmd
	)
	cmd = exec.Command(command, args...)
	cmd.Stderr = &buferr
	if raw, err = cmd.Output(); err != nil {
		return
	}
	data = raw
	if buferr.Len() > 0 {
		err = errors.New(buferr.String())
	}
	return
}

func (j *execJob) Remove(key string) {
	if t, ok := j.tasks.Get(key); ok {
		t.(*task).job.Quit <- true
		j.tasks.Remove(key)
	}
}

func (j *execJob) watchChanges(jobspath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					j.readDir(jobspath)
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					j.readDir(jobspath)
				}
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					j.Remove(filepath.Base(event.Name))
				}
				if event.Op&fsnotify.Rename == fsnotify.Rename {
					j.Remove(filepath.Base(event.Name))
				}
			case err := <-watcher.Errors:
				log.Printf("ExecJob error: %s", err)
			}
		}
	}()

	err = watcher.Add(jobspath)
	if err != nil {
		log.Println(err)
	}
	<-done

}

func init() {
	dashing.Register(&execJob{})
}
