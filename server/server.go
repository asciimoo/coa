package server

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/asciimoo/coa/config"
	"github.com/asciimoo/coa/notification"
)

func Listen(c *config.Config) {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "### CONFIGURATION ###\n\n%v", c.String())
	})

	http.HandleFunc("/api/add", func(w http.ResponseWriter, r *http.Request) {
		path := strings.Join(r.URL.Query()["path"], "")
		if path == "" {
			http.Error(w, "Missing path", http.StatusBadRequest)
			return
		}
		cwd := r.URL.Query()["cwd"][0]

		p, err := c.AddProject(filepath.Join(cwd, path))
		if err != nil {
			http.Error(w, "Cannot add project: "+err.Error(), http.StatusBadRequest)
			return
		}
		go p.Start()
		fmt.Fprint(w, "OK")
	})

	http.HandleFunc("/api/reload", func(w http.ResponseWriter, r *http.Request) {
		c2, err := config.Load(c.ConfigFolder)
		if err != nil {
			http.Error(w, "Cannot load configuration: "+err.Error(), http.StatusBadRequest)
			return
		}
		if c2.Notifiers == nil {
			http.Error(w, "Configuration error! Please specify at least one notification backend in your Coa settings", http.StatusBadRequest)
			return
		}
		for _, p := range c.Projects {
			p.Stop()
		}
		for _, p := range c2.Projects {
			go p.Start()
		}
		c = c2
		if err := notification.Initialize(c.Notifiers); err != nil {
			http.Error(w, "Server reload error! Failed to reload notifiers: "+err.Error(), http.StatusBadRequest)
			return
		}
	})

	for _, p := range c.Projects {
		go p.Start()
	}

	log.Println("Listening on", c.ServerAddress)
	log.Fatal(http.ListenAndServe(c.ServerAddress, nil))
}

func Call(u string, params map[string]string) error {
	cwd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}

	if strings.Index("http:", u) != 0 {
		u = "http://" + u
	}

	if params == nil {
		params = make(map[string]string)
	}

	params["cwd"] = cwd

	pu, err := url.Parse(u)
	if err != nil {
		return err
	}

	if pu.Scheme == "" {
		pu.Scheme = "http://"
	}

	q := pu.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	pu.RawQuery = q.Encode()

	resp, err := http.Get(pu.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(body)
	}

	if resp.StatusCode != 200 {
		return errors.New("Error received from server: " + string(body))
	}

	return nil
}
