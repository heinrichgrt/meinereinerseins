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
	zookeeperHost = "localhost"
	zookeeperPort = "2181"
)

var (
	DefaultTopicSettings map[string]string
)

//Resultset  Json for output
type Resultset struct {
	OSCommand            string    `json:"os.cmd"`
	Stdout               string    `json:"stdout"`
	Stderr               string    `json:"stderr,omitempty"`
	CmdStarttime         time.Time `json:"cmd.starttime"`
	CMDStoptime          time.Time `json:"=md.stoptime"`
	DurationSecounds     int       `json:"duration.seconds"`
	SuccessfullExecution bool      `json:"succesful"`
	ErrorStr             string    `json:"errorstr,omitempty"`
}
type Topic struct {
	Name              string
	NumberPartitions  int
	ReplicationFactor int
	Retentionms       int
}

var topicURL = "https://raw.githubusercontent.com/heinrichgrt/meinereinerseins/master/TopicClient/topic.yml"

func selfInit() {
	// set the topic defaults
	DefaultTopicSettings = make(map[string]string)
	DefaultTopicSettings["noofreplicas"] = "1"
	DefaultTopicSettings["noofpartitions"] = "1"
	DefaultTopicSettings["retiontionms"] = "2592000000"
}

type Topics struct {
	Tops []Topic `topics`
}

func readTopicData() *Topics {
	var t Topics

	return &t
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

func execCommand(execmd string, args *[]string) (string, string, string, bool) {
	//func execCommand(extcmd string, args []string) error, string, string{
	cmd := exec.Command(execmd, *args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	ret := true
	err := cmd.Run()
	errstring := ""
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if err != nil {
		//log.Fatalf("cmd.Run() failed with %s\n", err)
		log.Println("error: %v", err)
		ret = false
		errstring = err.Error()

	}

	fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)

	return outStr, errStr, errstring, ret
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
	out.Stdout, out.Stderr, out.ErrorStr, out.Result = execCommand(cmdstr, &args)
	out.Stoptime = time.Now()
	out.Secounds = int(out.Stoptime.Unix() - out.Starttime.Unix())
	json.NewEncoder(w).Encode(out)
	//fmt.Printf("out: %s, err: %s, sys: %v">q!, so, se, e)

	//log.Println("ende")

}

func main() {
	selfInit()
	r := mux.NewRouter()

	r.HandleFunc("/topics/create/{name}", createTopic)
	r.HandleFunc("/topics/acls/{name}/{user}/", setAclsTopic)
	//	r.HandleFunc("/create/topic/{name}/repliction/{norep}/partitions/{nopart}/retensms/{noms}", fullTopicHandler)
	http.ListenAndServe(":8088", r)
}
