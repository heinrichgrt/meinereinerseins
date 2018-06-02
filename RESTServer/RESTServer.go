package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	zookeeperhost = "localhost"
	zookeeperport = "2181"
)

var (
	topic map[string]string
)

//Resultset  Json for output
type Resultset struct {
	Command   string    `json:"cmd,omitempty"`
	Stdout    string    `json:"stdout,omitempty"`
	Stderr    string    `json:"stderr,omitempty"`
	Starttime time.Time `json:"starttime,omitempty"`
	Stoptime  time.Time `json:"stoptime,omitempty"`
	Secounds  int       `json:"seconds,omitempty"`
	Result    bool      `json:"succesful,omitempty"`
}

func selfInit() {
	// set the topic defaults
	topic = make(map[string]string)
	topic["noofreplicas"] = "1"
	topic["noofpartitions"] = "1"
	topic["retiontionms"] = "2592000000"
}


func setDefaultTopicValue(r *http.Request, v string) string {
	keys, ok := r.URL.Query()[v]
	if ok && len(keys) > 0 {
		log.Println("Url Param 'key' is present")
		return keys[0]
	}
	value, ok := topic[v]
	if ok {
		return value
	}
	// this is bad no info via url and no default
	return ""

}

func execCommand(execmd string, args *[]string) (string, string, bool) {
	//func execCommand(extcmd string, args []string) error, string, string{
	cmd := exec.Command(execmd, *args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	ret := true
	err := cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if err != nil {
		//log.Fatalf("cmd.Run() failed with %s\n", err)
		ret = false
	}

	fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)

	return outStr, errStr, ret
}

func setAclsTopic(w http.ResponseWriter, r *http.Request) {

}
func createTopic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	name := vars["name"]
	for key := range vars {
		fmt.Printf("k: %v v: %v\n", key, vars[key])
	}
	noofreplicas := setDefaultTopicValue(r, "noofreplicas")
	noofpartitions := setDefaultTopicValue(r, "noofpartitions")
	retiontionms := setDefaultTopicValue(r, "retiontionms")
	out := new(Resultset)

	// check invorenment for zk var
	//kafka-topics --create --zookeeper localhost:2181 --topic lall.fasel.x --replication-factor 1 --partitions 10 --config retention.ms=2592000000 --if-not-exists
	args := []string{"--create", "--topic", name, "--zookeeper", zookeeperhost + ":" + zookeeperport, "--replication-factor", noofreplicas,
		"--partitions", noofpartitions, "--config", "retention.ms=" + retiontionms, "--if-not-exists"}
	cmdstr := "/usr/local/bin/kafka-topics"

	out.Command = cmdstr + " " + strings.Join(args, " ")
	out.Starttime = time.Now()
	out.Stdout, out.Stderr, out.Result = execCommand(cmdstr, &args)
	out.Stoptime = time.Now()
	out.Secounds = int(out.Stoptime.Unix() - out.Starttime.Unix())
	json.NewEncoder(w).Encode(out)
	//fmt.Printf("out: %s, err: %s, sys: %v">q!, so, se, e)

	//log.Println("ende")

}

/*
func getValue(v *map[string][string], parm string) {

}
*/
func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

func main() {
	selfInit()
	r := mux.NewRouter()
	r.HandleFunc("//{title}/page/{page}", homePage)
	r.HandleFunc("/test/{title}/p/{page}", getBooks)
	r.HandleFunc("/topics/create/{name}", createTopic)
	r.HandleFunc("/topics/acls/{name}/{user}/"), setAclsTopic)
	//	r.HandleFunc("/create/topic/{name}/repliction/{norep}/partitions/{nopart}/retensms/{noms}", fullTopicHandler)

	http.ListenAndServe(":8092", r)
}
