package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type Resultset struct {
	Command   string    `json:"cmd"`
	Stdout    string    `json:"stdout"`
	Stderr    string    `json:"stderr,omitempty"`
	Starttime time.Time `json:"starttime"`
	Stoptime  time.Time `json:"stoptime"`
	Secounds  int       `json:"seconds"`
	Result    bool      `json:"succesful"`
	ErrorStr  string    `json:"errorstr,omitempty"`
}

type Topic struct {
	Name        string
	Partitions  int
	Replication int
	Retentionms int
}

type Topics struct {
	Tops []Topic `topics`
}

var Config struct {
	targetHost string
	targetPort int
}

/*
var data = `
topics:
  - topic:
    name: "lall.fasel.03"
    partitions: 10
    replication: 3
    retentionms: 234000000

  - topic:
    name: "branch.lall.fasel.04"
    partitions: 5
    replication: 2
    retentionms: 234000232

`
*/
func getData() (int, string) {
	res, err := http.Get("https://raw.githubusercontent.com/heinrichgrt/meinereinerseins/master/TopicClient/topic.yml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	result, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
		return 1, ""
	}

	return 0, string(result)

}
func sendData(t *Topic) (int, *Resultset) {
	res, err := http.Get("http://localhost:8088/topics/create/" + t.Name)
	if err != nil {
		log.Fatalf("error: %v", err)
	} /*else {
		log.Debug("")
	}
	*/
	result, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
		return 1, nil
	}
	response := Resultset{}
	json.Unmarshal([]byte(result), &response)
	fmt.Printf("%s", result)
	return 0, &response
}
func main() {

	var topix Topics

	/*
		 filename := os.Args[1]
		 source, err := ioutil.ReadFile(filename)
		 if err != nil {
			 panic(err)
		 }
	*/
	_, data := getData()
	source := []byte(data)

	err := yaml.Unmarshal(source, &topix)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	//range config.Cfgs[id].Bar
	for t := range topix.Tops {
		//	fmt.Printf("topic %v\n", topix.Tops[t])
		fmt.Printf("topic: %s, parts: %d, replicas: %v; ret days %v\n",
			topix.Tops[t].Name, topix.Tops[t].Partitions, topix.Tops[t].Replication, topix.Tops[t].Retentionms)
		// var r int
		//R := Resultset{}
		r, R := sendData(&topix.Tops[t])
		fmt.Printf("%v is %v", r, R)
	}
	fmt.Printf("--- config:\n%v\n\n", topix)
}
