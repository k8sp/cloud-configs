package template

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
	"time"

	"gopkg.in/yaml.v2"
)

type PerNodeConfig struct {
	IP       string `yaml:"ip"`
	EtcdRole string `yaml:"etcd_role"`
	Hostname string `yaml:"hostname"`
	NicName  string `yaml:"nic_name"`
}

type GlobalConfig struct {
	SSHAuthorizedKeys string `yaml:"ssh_authorized_keys"`
	SSHPrivateKey     string `yaml:"ssh_private_key"`
}

type ExecutionConfig struct {
	PerNodeConfig
	GlobalConfig
}

type Config struct {
	Nodes  map[string]PerNodeConfig
	Global GlobalConfig
}

// Execute returns the executed cloud-config template for a node with
// given MAC address.
func Execute(tmpl, cfg []byte, mac string, w io.Writer) error {
	t, e := template.New("cloud-config").Parse(string(tmpl))
	if e != nil {
		return e
	}

	c := &Config{}
	e = yaml.Unmarshal(cfg, &c)
	if e != nil {
		return e
	}

	ec := ExecutionConfig{
		PerNodeConfig: c.Nodes[mac],
		GlobalConfig:  c.Global,
	}
	return t.Execute(w, ec)
}

func httpGet(url string, timeout time.Duration) []byte {
	c := http.Client{
		Timeout: timeout,
		// TODO: Able to retrieve via https:// from private Github repos.
	}

	resp, e := c.Get(url)
	if e != nil {
		panic(e)
	}
	defer resp.Body.Close()

	b, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		panic(e)
	}

	return b
}

func Retrieve(tmplUrl, cfgUrl string, timeout time.Duration) (tmpl, cfg []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	tmpl = httpGet(tmplUrl, timeout)
	cfg = httpGet(cfgUrl, timeout)
	return tmpl, cfg, err
}

func SetToEtcd(tmpl, cfg []byte, etcdEndpoints string) error {
	// TODO(jiameng): Implementation to be done.
	return nil
}

func GetFromEtcd(etcdEndpoints string) (tmpl, cfg []byte, err error) {
	// TODO(jiameng): Implementation to be done.
	return nil, nil, nil
}

var (
	FLAG_TemplateUrl   = flag.String("template-url", "", "URL to template file")
	FLAG_ConfigUrl     = flag.String("config-url", "", "URL to config file")
	FLAG_EtcdEndpoints = flag.String("etcd", "", "etcd endpoints")
)

func Handler(w http.ResponseWriter, r *http.Request) {
	mac := r.Form["mac"]
	if mac == nil || len(mac) < 1 {
		log.Printf("Error: %v doesn't have mac paramter", r.URL)
	}

	tmpl, cfg, e := Retrieve(*FLAG_TemplateUrl, *FLAG_ConfigUrl, time.Second)
	if e == nil {
		SetToEtcd(tmpl, cfg, *FLAG_EtcdEndpoints)
	} else {
		tmpl, cfg, e = GetFromEtcd(*FLAG_EtcdEndpoints)
		if e != nil {
			log.Printf("Error: %v", e)
			http.Error(w, fmt.Sprint(e), http.StatusInternalServerError)
			return
		}
	}

	if e := Execute(tmpl, cfg, mac[0], w); e != nil {
		log.Printf("Error: %v", e)
		http.Error(w, fmt.Sprint(e), http.StatusInternalServerError)
	}
}
