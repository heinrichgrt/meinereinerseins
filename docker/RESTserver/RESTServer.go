package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

const (
	zookeeperHost             string = "localhost"
	zookeeperPort             string = "2181"
	DefaultNumberOfPartitions int    = 10
	DefaultReplicationFactor  int    = 1
	DefaultConfigRetentionMs  int    = 2628000000
	EnvironMentVarPrefix      string = "TCRS_"
)

var (
	AppConfig   map[string]string
	ConfigItems = [5]string{
		"NumberOfPartitions",
		"ReplicationFactor",
		"ConfigRetentionMs",
		"topicURL",
		"aclsURL",
	}
)

//Resultset  Json for output
type Resultset struct {
	OSCommand            string    `json:"os.cmd"`
	CommandArgs          []string  `json:"cmd.args"`
	Stdout               string    `json:"stdout"`
	Stderr               string    `json:"stderr,omitempty"`
	CmdStarttime         time.Time `json:"cmd.starttime"`
	CMDStoptime          time.Time `json:"=md.stoptime"`
	DurationSecounds     int       `json:"duration.seconds"`
	SuccessfullExecution bool      `json:"succesful"`
	ErrorStr             string    `json:"errorstr,omitempty"`
}

type Topics struct {
	Tops []Topic `topics`
}

type Topic struct {
	TopicName          string `yaml:"topic.name"`
	NumberOfPartitions int    `yaml:"number.of.partitions"`
	ReplicationFactor  int    `yaml:"replication.factor"`
	ConfigRetentionMs  int    `yaml:"config.retention.ms"`
}
type Acls struct {
	Rules []Acl `acls`
}
type Acl struct {
	TopicName string `yaml:"topic.name"`
	User      string `yaml:"user"`
	Action    string `yaml:"action"`
	Role      string `yaml:"role"`
}

func selfInit() {
	// set the topic defaults
	AppConfig = make(map[string]string)
	AppConfig["NumberOfPartitions"] = strconv.Itoa(DefaultNumberOfPartitions)
	AppConfig["ReplicationFactor"] = strconv.Itoa(DefaultNumberOfPartitions)
	AppConfig["ConfigRetentionMs"] = strconv.Itoa(DefaultConfigRetentionMs)
	AppConfig["topicURL"] = "https://raw.githubusercontent.com/heinrichgrt/meinereinerseins/master/TopicClient/topic.yml"
	AppConfig["aclsURL"] = "https://raw.githubusercontent.com/heinrichgrt/meinereinerseins/master/RESTServer/ACLs.yml"
	for i := range ConfigItems {
		AppConfig[ConfigItems[i]] = setDefaultFromEnvironment(ConfigItems[i], AppConfig[ConfigItems[i]])
	}

}

func setDefaultFromEnvironment(envVar string, defaultValue string) string {
	// todo: make this uppercase and prefix it.
	if testVar := os.Getenv(strings.ToUpper(EnvironMentVarPrefix + envVar)); testVar != "" {
		return testVar
	} else {
		return defaultValue
	}
}

func setTopicParams(fromTopic int, fromDefault int) string {
	if fromTopic != 0 {
		log.Debugf("value taken from yml", fromTopic)
		return strconv.Itoa(fromTopic)
	} else if fromDefault != 0 {
		log.Debugf("value taken from default", fromDefault)
		return strconv.Itoa(fromDefault)
	}
	log.Debug("There is no vaulue or default for %v")
	return ""
}

func getYamlData(configKey string) []byte {
	res, err := http.Get(AppConfig[configKey])
	if err != nil {
		log.Errorf("error in appconfig: %v", err)
		return nil
	}
	result, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Errorf("Can not get Topic YAML")
		return nil
	}

	return result

}

func execCommand(CommandOutPut *Resultset) {
	//func execCommand(extcmd string, args []string) error, string, string{

	cmd := exec.Command(CommandOutPut.OSCommand, CommandOutPut.CommandArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	CommandOutPut.SuccessfullExecution = true
	CommandOutPut.CmdStarttime = time.Now()

	err := cmd.Run()
	CommandOutPut.Stdout = string(stdout.Bytes())
	CommandOutPut.Stderr = string(stderr.Bytes())
	CommandOutPut.CMDStoptime = time.Now()
	CommandOutPut.DurationSecounds = int(CommandOutPut.CMDStoptime.Unix() - CommandOutPut.CmdStarttime.Unix())
	if err != nil {
		//log.Fatalf("cmd.Run() failed with %s\n", err)
		log.Errorln("Command failed %v err: ", err)
		CommandOutPut.SuccessfullExecution = false
		CommandOutPut.ErrorStr = err.Error()

	}

	//	fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)

	return
}

func setAclsTopic(w http.ResponseWriter, r *http.Request) {

}

func createOneTopic(Topic Topic) *Resultset {

	CommandResult := new(Resultset)

	if Topic.TopicName == "" {

		//cleanup to error result
		log.Errorln("Topic name is mandatory")
		CommandResult.ErrorStr = "Topic name is mandatory, skipping item"
		return CommandResult

	}

	replicationFactor := setTopicParams(Topic.ReplicationFactor, DefaultReplicationFactor)
	noofpartitions := setTopicParams(Topic.NumberOfPartitions, DefaultNumberOfPartitions)
	retiontionms := setTopicParams(Topic.ConfigRetentionMs, DefaultConfigRetentionMs)
	CommandResult.OSCommand = "/usr/bin/kafka-topics"
	CommandResult.CommandArgs = []string{"--create", "--topic", Topic.TopicName, "--zookeeper",
		zookeeperHost + ":" + zookeeperPort, "--replication-factor", replicationFactor,
		"--partitions", noofpartitions, "--config", "retention.ms=" + retiontionms,
		"--if-not-exists"}

	execCommand(CommandResult)
	return CommandResult
}
func setOneACL(acl Acl) *Resultset {
	CommandResult := new(Resultset)

	if acl.TopicName == "" {

		CommandResult.ErrorStr = "Cmd not executed. Topic name is mandatory"
		return CommandResult

	}
	if acl.Role != "producer" && acl.Role != "consumer" && acl.Role != "both" {
		CommandResult.ErrorStr = "Role must be producer or consumer "
	}
	if acl.User == "" {
		CommandResult.ErrorStr = "user must be set"
	}
	CommandResult.OSCommand = "/usr/bin/kafka-acls"
	CommandResult.CommandArgs = []string{"--" + acl.Action,
		"--topic", acl.TopicName, "--" + acl.Role,
		"--authorizer-properties", "zookeeper.connect=" + zookeeperHost + ":" + zookeeperPort,
		"--allow-principal", "User:" + acl.User, "--group", "'*'"}

	execCommand(CommandResult)
	log.Infof("acl set was %v", CommandResult)
	return CommandResult
}

func createTopics(w http.ResponseWriter, r *http.Request) {
	// add check for no internet
	yamlData := getYamlData("topicURL")
	ResultList := make([]*Resultset, 0)
	if yamlData == nil {
		fmt.Println("gna")
		CommandResult := new(Resultset)
		CommandResult.ErrorStr = "could not fetch yaml file for topics"
		ResultList = append(ResultList, CommandResult)
	}

	var topics Topics

	err := yaml.Unmarshal(yamlData, &topics)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	for singleTopic := range topics.Tops {
		createResult := createOneTopic(topics.Tops[singleTopic])
		fmt.Printf("lall: %v", createResult)
		ResultList = append(ResultList, createResult)

	}
	json.NewEncoder(w).Encode(ResultList)
}
func setACLs(w http.ResponseWriter, r *http.Request) {

	// hier geht es weiter
	yamlData := getYamlData("aclsURL")
	ResultList := make([]*Resultset, 0)
	if yamlData == nil {
		fmt.Println("gna")
		CommandResult := new(Resultset)
		CommandResult.ErrorStr = "could not fetch yaml file for acls"
		ResultList = append(ResultList, CommandResult)
	}
	var acls Acls

	err := yaml.Unmarshal(yamlData, &acls)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("yaml: %v \n acls %v", string(yamlData), acls)
	for singleACL := range acls.Rules {
		//createResult := createOneTopic(topics.Tops[singleTopic])
		fmt.Printf("acls: %v", acls.Rules[singleACL])
		//	createResult := setOneACL(acls.Rules[singleACL])
		createResult := setOneACL(acls.Rules[singleACL])
		ResultList = append(ResultList, createResult)
	}
	json.NewEncoder(w).Encode(ResultList)
}

func showhealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "I am good")
}

func main() {
	selfInit()
	restRouter := mux.NewRouter()

	restRouter.HandleFunc("/topics/create", createTopics)
	restRouter.HandleFunc("/acls/create", setACLs)
	restRouter.HandleFunc("/health", showhealth)
	http.ListenAndServe(":8088", restRouter)
}
