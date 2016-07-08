package main

import (
//    "fmt"
    "log"
    "net/http"
    "io/ioutil"
    "time"
    "github.com/gorilla/mux"
    "golang.org/x/net/context"
    "github.com/coreos/etcd/client"
    tp "./template"
)

var etcd_template_key = "/unisound/template_server/template"
var etcd_config_key = "/unisound/template_server/config"
var template_url = "https://raw.githubusercontent.com/k8sp/cloud-configs/template-server/template/cloud-config.template"
var config_url = "https://raw.githubusercontent.com/k8sp/cloud-configs/template-server/template/build_config.yml"

var kapi client.KeysAPI

func init() {
    cfg := client.Config{
        Endpoints:               []string{"http://127.0.0.1:2379"},
        Transport:               client.DefaultTransport,
        // set timeout per request to fail fast when the target endpoint is unavailable
        HeaderTimeoutPerRequest: time.Second,
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
            CacheToEtcd(template, config)
        }
    }()
}

func HttpHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    mac_addr := vars["mac"]

    templ, config, err := RetriveFromGithub(1 * time.Second)
    if err != nil {
        CacheToEtcd(templ, config)
    } else {
        templ, config, err = RetrieveFromEtcd()
        if err != nil {
            return
        }
    }
    tpl := template.Parse(tmpl)
    var cfg Config
    err := yaml.Unmarshal([]byte(config), &config)
    return tp.Execute(tpl, cfg, mac_addr, w)
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
    return template, config, err
}

func RetrieveFromEtcd() (template string, config string, err error){
    resp, err := kapi.Get(context.Background(), etcd_template_key, nil)
    if err != nil {
        log.Fatal(err)
        return "", "", err
    } else {
        template := resp.Node.Value
    }
    resp, err = kapi.Get(context.Background(), etcd_config_key, nil)
    if err != nil {
        log.Fatal(err)
        return "", "", err
    } else {
        config := resp.Node.Value
    }
    return template, config, nil
}

func CacheToEtcd(template string, config string){
    resp, err := kapi.Set(context.Background(), etcd_template_key, template, nil)
    if err != nil {
        log.Fatal(err)
    } else {
        // print common key info
        log.Printf("Set is done. Metadata is %q\n", resp)
    }
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
        return "", err
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    return string(body), nil
}

