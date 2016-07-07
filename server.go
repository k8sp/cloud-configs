import (
    "fmt"
    "log"
    "net/http"
    "time"
    "github.com/gorilla/mux"
    "golang.org/x/net/context"
    "github.com/coreos/etcd/client"
)

var kapi NewKeysAPI
var etcd_template_key = "/unisound/template_server/template"
var etcd_config_key = "/unisound/template_server/config"
var template_url = ""
var config_url = ""



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
    kapi := client.NewKeysAPI(c)
}

func main() {
    router := mux.NewRouter().StrictSlash(true)
    router.HandleFunc("/cloud-config/{mac}", HttpHandler)
    log.Fatal(http.ListenAndServe(":8080", router))
}

func HttpHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    mac_addr := vars["mac"]

  template, config, timeout := RetriveFromGithub(1 * time.Second)
  if !timeout {
    CacheToEtcd(template, config)
  } else {
    template, config, ok := RetrieveFromEtcd()
    if !ok {
          return error
    }
  }
  return Execute(template, config[mac])
}



func UpdateEtcdCache() {
  for {
    time.Sleep(10 * time.Minute)
    template, config := RetriveFromGithub(30 * time.Second)
        WriteToEtcd(template, config)
  }
}


func RetriveFromGithub(timeout) {

}

func RetrieveFromEtcd(){


}

func CacheToEtcd(template string, config string){


}


func UpdateEtcdCache() {
    for {
        time.Sleep(10 * time.Minute)
        template, config := RetriveFromGithub(30 * time.Second)
        WriteToEtcd(template, config)
    }
}
