package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	yaml "gopkg.in/yaml.v2"
)

type Topic struct {
	Name        string
	Partitions  int
	Replication int
	Retentionms int
}

type Topics struct {
	Tops []Topic `topics`
}

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

func sendData(t *Topic) int {
	res, err := http.Get("http://localhost:8088/topics/create/" + t.Name)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	robots, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", robots)
	return 0
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
		sendData(&topix.Tops[t])

	}
	fmt.Printf("--- config:\n%v\n\n", topix)
}
