package template

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/topicai/candy"
	"gopkg.in/yaml.v2"
)

func TestExecute(t *testing.T) {
	candy.WithOpened("build_config.yml", func(r io.Reader) interface{} {
		cfg, e := ioutil.ReadAll(r)
		candy.Must(e)

		c := &Config{}
		assert.Nil(t, yaml.Unmarshal(cfg, &c))

		candy.WithOpened("cloud-config.template", func(r io.Reader) interface{} {
			tmpl, e := ioutil.ReadAll(r)
			candy.Must(e)

			Execute(tmpl, cfg, "00:25:90:c0:f7:62", os.Stdout)
			return nil
		})

		return nil
	})
}

func TestRetrieve(t *testing.T) {
	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "config")
	})
	http.HandleFunc("/template", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "template")
	})

	ln, e := net.Listen("tcp", ":0") // Allocate an unused port.
	candy.Must(e)
	go http.Serve(ln, nil)

	tmpl, cfg, e := Retrieve(
		"http://"+ln.Addr().String()+"/template",
		"http://"+ln.Addr().String()+"/config",
		time.Second)
	assert.Nil(t, e)
	assert.Equal(t, "config", string(cfg))
	assert.Equal(t, "template", string(tmpl))
}
