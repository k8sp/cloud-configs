package main

import (
    "fmt"
    "log"
    "net/http"
    "io/ioutil"
    "time"
    "github.com/gorilla/mux"
    "golang.org/x/net/context"
    "github.com/coreos/etcd/client"
    "text/template"
    "gopkg.in/yaml.v2"
    tp "template-server/template"
)

var etcd_template_key = "/unisound/template_server/template"
var etcd_config_key = "/unisound/template_server/config"
var template_url = "https://raw.githubusercontent.com/k8sp/cloud-configs/jiameng-template-server/template/cloud-config.template"
var config_url = "https://raw.githubusercontent.com/k8sp/cloud-configs/jiameng-template-server/template/build_config.yml"

var kapi client.KeysAPI

func init() {
    cfg := client.Config{
        Endpoints:               []string{"http://127.0.0.1:2379"},
        Transport:               client.DefaultTransport,
        // set timeout per request to fail fast when the target endpoint is unavailable
        HeaderTimeoutPerRequest: time.Second * 2,
    }
    c, err := client.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    kapi = client.NewKeysAPI(c)
}

func main() {
    router := mux.NewRouter().StrictSlash(true)
    router.HandleFunc("/cloud-config/{mac}", HttpHandler)
    log.Fatal(http.ListenAndServe(":8080", router))

    ticker := time.NewTicker(time.Minute * 10)
    go func() {
        for _ = range ticker.C {
            template, config, err := RetriveFromGithub(30 * time.Second)
            if err != nil {
                continue
            }
            CacheToEtcd(template, config)
        }
    }()
}

func HttpHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    mac_addr := vars["mac"]

    templ, config, err := RetriveFromGithub(10 * time.Second)
    if err != nil {
        templ, config, err = RetrieveFromEtcd()
        if err != nil {
            return
        }
    } else {
        CacheToEtcd(templ, config)
    }
    tpl := template.Must(template.New("template").Parse(templ))
    cfg := &tp.Config{}
    err = yaml.Unmarshal([]byte(config), &cfg)
    tp.Execute(tpl, cfg, mac_addr, w)
}

func RetriveFromGithub(timeout time.Duration) (template string, config string, err error){
    template, err = httpGet(template_url, timeout)
    if err != nil {
        log.Fatal(err)
        return "", "", err
    }
    config, err = httpGet(config_url, timeout)
    if err != nil {
        log.Fatal(err)
        return "", "", err
    }
    return template, config, nil 
}

func RetrieveFromEtcd() (template string, config string, err error){
    resp, err := kapi.Get(context.Background(), etcd_template_key, nil)
    if err != nil {
        log.Fatal(err)
        return "", "", err
    } else {
        template = resp.Node.Value
    }
    resp, err = kapi.Get(context.Background(), etcd_config_key, nil)
    if err != nil {
        log.Fatal(err)
        return "", "", err
    } else {
        config = resp.Node.Value
    }
    return template, config, nil
}

func CacheToEtcd(template string, config string){
    fmt.Printf("%#v\n", etcd_template_key)
    resp, err := kapi.Set(context.Background(), etcd_template_key, template, nil)
    if err != nil {
        log.Fatal(err)
    } else {
        // print common key info
        log.Printf("Set is done. Metadata is %q\n", resp)
    }
    fmt.Printf("%#v\n", etcd_config_key)
    resp, err = kapi.Set(context.Background(), etcd_config_key, config, nil)
    if err != nil {
        log.Fatal(err)
    } else {
        // print common key info
        log.Printf("Set is done. Metadata is %q\n", resp)
    }
}

func httpGet(url string, timeout time.Duration) (string, error) {
    client := http.Client{
        Timeout: timeout,
    }
    resp, err := client.Get(url)
    if err != nil {
        log.Fatal(err)
        return "", err
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
        return "", err
    }
    return string(body), nil
}

