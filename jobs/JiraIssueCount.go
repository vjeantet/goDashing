package jobs

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/fsnotify.v1"

	"github.com/BurntSushi/toml"
	"github.com/PuerkitoBio/goquery"
	"github.com/andygrunwald/go-jira"
	"github.com/streamrail/concurrent-map"
	"github.com/vjeantet/goDashing"
)

type jiraIssueCount struct {
	config *jiraIssueConfig
}

type jiraIssueConfig struct {
	Url        string
	Username   string
	Password   string
	Interval   int
	Indicators cmap.ConcurrentMap
}

type JiraIssurConfigIndicator struct {
	Jql          string
	WarningOver  int
	DangerOver   int
	Interval     int
	WarningUnder int
	DangerUnder  int
}

func (j *jiraIssueCount) Work(send chan *dashing.Event, webroot string, url string, token string) {
	j.config = &jiraIssueConfig{
		Interval:   60,
		Indicators: cmap.New(),
	}

	// Lire le fichier de conf
	if _, err := toml.DecodeFile(webroot+"conf/jiraissuecount.ini", &j.config); err != nil {
		log.Printf("JiraJob : can not read config file %s", "conf/jiraissuecount.ini")
		return
	}

	if j.config.Url == "" {
		log.Println("JiraJob : not started (no configuration)")
		return
	}

	// Capture indicators from dashbords
	j.readIndicators(webroot + "dashboards/")
	j.pushData(send)

	go j.watchChanges(webroot + "dashboards/")

	ticker := time.NewTicker(time.Duration(j.config.Interval) * time.Second)
	for {
		select {
		case <-ticker.C:
			j.pushData(send)
		}
	}
}

func (j *jiraIssueCount) pushData(send chan *dashing.Event) {
	for WID, indicator := range j.config.Indicators.Items() {
		count, err := j.getNumberOfIssues(indicator.(JiraIssurConfigIndicator).Jql)
		if err != nil {
			log.Printf("JiraJob : error jira search : %s", err)
			continue
		}

		status, _ := j.getIndicatorStatus(count, indicator.(JiraIssurConfigIndicator))

		send <- dashing.NewEvent(
			WID,
			map[string]interface{}{
				"current": count,
				"status":  status,
			},
			"")

	}
}

func (j *jiraIssueCount) getIndicatorStatus(count int, indicator JiraIssurConfigIndicator) (string, error) {
	if indicator.DangerOver > 0 && count > indicator.DangerOver {
		return "warning", nil
	}

	if indicator.WarningOver > 0 && count > indicator.WarningOver {
		return "danger", nil
	}

	if indicator.DangerUnder > 0 && count < indicator.DangerUnder {
		return "warning", nil
	}

	if indicator.WarningUnder > 0 && count < indicator.WarningUnder {
		return "danger", nil
	}

	return "normal", nil
}

func (j *jiraIssueCount) getNumberOfIssues(jql string) (int, error) {

	jiraClient, err := jira.NewClient(nil, j.config.Url)
	if err != nil {
		return 0, fmt.Errorf("JiraJob : error jira connect : %s", err)
	}

	if j.config.Username != "" && jiraClient.Authentication.Authenticated() == false {
		res, err := jiraClient.Authentication.AcquireSessionCookie(j.config.Username, j.config.Password)
		if err != nil || res == false {
			fmt.Printf("JiraJob : Authentification error : %v\n", res)
			return 0, err
		}
	}

	options := jira.SearchOptions{
		StartAt:    0,
		MaxResults: 1,
	}

	_, body, err := jiraClient.Issue.Search(jql, &options)
	if err != nil {
		return 0, fmt.Errorf("JiraJob : search error : %s, %s", err, jql)
	}

	return body.Total, nil

}

func (j *jiraIssueCount) readIndicators(dashroot string) {
	//init empty Indicators
	for k := range j.config.Indicators.Items() {
		j.config.Indicators.Remove(k)
	}

	// open each gerb
	files, _ := filepath.Glob(dashroot + "**/*.gerb")
	files2, _ := filepath.Glob(dashroot + "*.gerb")
	files = append(files, files2...)
	for _, file := range files {
		reader, err := os.Open(file)
		if err != nil {
			log.Println("JiraJob : error openning file : " + err.Error())
			continue
		}

		doc, err := goquery.NewDocumentFromReader(reader)
		reader.Close()
		if err != nil {
			log.Println("JiraJob : error goquery file : " + err.Error())
			continue
		}

		// find job="jira-count-filter"
		doc.Find("div[jira-count-filter]").Each(func(i int, s *goquery.Selection) {
			var jobInterval, dangerOver, warningOver, dangerUnder, warningUnder int

			jobInterval = j.config.Interval
			dangerOver = 0
			warningOver = 0

			// For each item found, get the properties
			widgetID, _ := s.Attr("data-id")
			jql, _ := s.Attr("jira-count-filter")
			jobIntervalString, _ := s.Attr("jira-interval")
			dangerOverString, _ := s.Attr("jira-danger-over")
			warningOverString, _ := s.Attr("jira-warning-over")
			dangerUnderString, _ := s.Attr("jira-danger-under")
			warningUnderString, _ := s.Attr("jira-warning-under")

			jobInterval, _ = strconv.Atoi(jobIntervalString)
			dangerOver, _ = strconv.Atoi(dangerOverString)
			warningOver, _ = strconv.Atoi(warningOverString)
			dangerUnder, _ = strconv.Atoi(dangerUnderString)
			warningUnder, _ = strconv.Atoi(warningUnderString)

			// register indicator
			j.config.Indicators.Set(widgetID,
				JiraIssurConfigIndicator{
					Jql:          "filter=" + jql,
					Interval:     jobInterval,
					DangerOver:   dangerOver,
					WarningOver:  warningOver,
					DangerUnder:  dangerUnder,
					WarningUnder: warningUnder,
				})

		})

		// find job="jira-count-jql"
		doc.Find("div[jira-count-jql]").Each(func(i int, s *goquery.Selection) {
			var jobInterval, dangerOver, warningOver, dangerUnder, warningUnder int
			jobInterval = j.config.Interval
			dangerOver = 0
			warningOver = 0

			// For each item found, get the properties
			widgetID, _ := s.Attr("data-id")
			jql, _ := s.Attr("jira-count-jql")
			jobIntervalString, _ := s.Attr("jira-interval")
			dangerOverString, _ := s.Attr("jira-danger-over")
			warningOverString, _ := s.Attr("jira-warning-over")
			dangerUnderString, _ := s.Attr("jira-danger-under")
			warningUnderString, _ := s.Attr("jira-warning-under")

			jobInterval, _ = strconv.Atoi(jobIntervalString)
			dangerOver, _ = strconv.Atoi(dangerOverString)
			warningOver, _ = strconv.Atoi(warningOverString)
			dangerUnder, _ = strconv.Atoi(dangerUnderString)
			warningUnder, _ = strconv.Atoi(warningUnderString)

			// register indicator
			j.config.Indicators.Set(widgetID,
				JiraIssurConfigIndicator{
					Jql:          jql,
					Interval:     jobInterval,
					DangerOver:   dangerOver,
					WarningOver:  warningOver,
					DangerUnder:  dangerUnder,
					WarningUnder: warningUnder,
				})
		})
	}
}

func (j *jiraIssueCount) watchChanges(dashroot string) {
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
				if event.Op&fsnotify.Write == fsnotify.Write {
					j.readIndicators(dashroot)
				}

				if event.Op&fsnotify.Create == fsnotify.Create {
					f, _ := os.Stat(event.Name)
					if f.IsDir() {
						err = watcher.Add(dashroot + f.Name())
						if err != nil {
							log.Println(err)
						}
					}
				}

			case err := <-watcher.Errors:
				log.Println("JiraJob : error:", err)
			}
		}
	}()

	err = watcher.Add(dashroot)
	if err != nil {
		log.Println(err)
	}
	files, _ := ioutil.ReadDir(dashroot)
	for _, f := range files {
		if f.IsDir() {
			err = watcher.Add(dashroot + f.Name())
			if err != nil {
				log.Println(err)
			}
		}

	}

	<-done

}

func init() {
	dashing.Register(&jiraIssueCount{})
}
