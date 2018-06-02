package main

import (
	"fmt"
	"log"

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
    partitions: "10"
    replication: 3
    retentionms: 234000000
   
  - topic: 
    name: "branch.lall.fasel.04"
    partitions: 5
    replication: 2
    retentionms: 234000232
  
`

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

	fmt.Printf("--- config:\n%v\n\n", topix)
}
