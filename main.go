package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/tucnak/telebot.v2"
)

var configPath string

func main() {
	var err error

	flag.StringVar(&configPath, "c", "./config/config.json", "Caminho para o arquivo de configuração. Padrão: ./config/config.json")

	err = InitializeConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	b, err := telebot.NewBot(telebot.Settings{
		Token:  appConfig.TelegramApiKey,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	for _, camera := range appConfig.Cameras {
		b.Handle(fmt.Sprintf("/%s", camera.Command), func(m *telebot.Message) {
			_, userFound := appConfig.Users[m.Sender.ID]

			if !userFound {
				log.Printf("Unknown user: %d (%s %s)\n", m.Sender.ID, m.Sender.FirstName, m.Sender.LastName)
				return
			}

			lCamera := camera
			lCamera.SendPhotoTo(b, m.Sender)
		})
	}

	b.Handle("/tudo", func(m *telebot.Message) {
		_, userFound := appConfig.Users[m.Sender.ID]

		if !userFound {
			log.Printf("Unknown user: %d (%s %s)\n", m.Sender.ID, m.Sender.FirstName, m.Sender.LastName)
			return
		}

		for _, camera := range appConfig.Cameras {
			camera.SendPhotoTo(b, m.Sender)
		}
	})

	b.Handle("/who", func(m *telebot.Message) {
		log.Printf("%d (%s %s)\n", m.Sender.ID, m.Sender.FirstName, m.Sender.LastName)
	})

	b.Handle("/start", func(m *telebot.Message) {
		_, userFound := appConfig.Users[m.Sender.ID]

		if !userFound {
			log.Printf("Unknown user: %d (%s %s)\n", m.Sender.ID, m.Sender.FirstName, m.Sender.LastName)
			return
		}

		b.Send(m.Sender, "oi")
	})

	b.Handle("/temp", func(m *telebot.Message) {
		_, userFound := appConfig.Users[m.Sender.ID]

		if !userFound {
			log.Printf("Unknown user: %d (%s %s)\n", m.Sender.ID, m.Sender.FirstName, m.Sender.LastName)
			return
		}

		fi, err := os.Stat("/sys/class/thermal/thermal_zone0/temp")

		if err != nil {
			log.Println(err)
			b.Send(m.Sender, err.Error())
			return
		}

		if fi.IsDir() {
			log.Println("Error: Dir")
			b.Send(m.Sender, "Error: Dir")
			return
		}

		file, err := ioutil.ReadFile("/sys/class/thermal/thermal_zone0/temp")

		if err != nil {
			log.Println(err)
			b.Send(m.Sender, err.Error())
			return
		}

		// Convert []byte to string and print to screen
		temp, err := strconv.Atoi(strings.TrimSpace(string(file)))

		if err != nil {
			log.Println(err)
			b.Send(m.Sender, err.Error())
			return
		}

		b.Send(m.Sender, fmt.Sprintf("%.2f°C", float64(temp)/1000))
	})

	b.Handle("/ip", func(m *telebot.Message) {
		_, userFound := appConfig.Users[m.Sender.ID]

		if !userFound {
			log.Printf("Unknown user: %d (%s %s)\n", m.Sender.ID, m.Sender.FirstName, m.Sender.LastName)
			return
		}

		var client = &http.Client{
			Timeout: time.Second * 10,
		}

		r, err := client.Get("https://httpbin.org/ip")
		if err != nil {
			b.Send(m.Sender, err.Error())
			return
		}

		defer r.Body.Close()

		ipData := struct {
			Origin string `json:"origin"`
		}{}

		err = json.NewDecoder(r.Body).Decode(&ipData)

		if err != nil {
			b.Send(m.Sender, err.Error())
			return
		}

		sb := strings.Builder{}

		sb.WriteString(fmt.Sprintf("%s\n", ipData.Origin))
		sb.WriteString("-----------------------------\n")
		sb.WriteString("http://blamoo.dynu.net:5952/zzz\n")
		sb.WriteString("http://192.168.0.52/zzz\n")
		sb.WriteString(fmt.Sprintf("http://%s:5952/zzz\n", ipData.Origin))
		sb.WriteString("-----------------------------\n")
		sb.WriteString("https://blamoo.dynu.net:9798\n")
		sb.WriteString("https://192.168.0.52:9798\n")
		sb.WriteString(fmt.Sprintf("https://%s:9798\n", ipData.Origin))

		b.Send(m.Sender, sb.String())
	})

	b.Handle("/cfg", func(m *telebot.Message) {
		_, userFound := appConfig.Users[m.Sender.ID]

		if !userFound {
			log.Printf("Unknown user: %d (%s %s)\n", m.Sender.ID, m.Sender.FirstName, m.Sender.LastName)
			return
		}

		sb := strings.Builder{}

		sb.WriteString("start - Start\n")
		sb.WriteString("who - Who\n")
		sb.WriteString("temp - Temperatura\n")
		sb.WriteString("ip - Ip\n")
		sb.WriteString("cfg - Configuração\n")
		for _, camera := range appConfig.Cameras {
			sb.WriteString(fmt.Sprintf("%s - %s\n", camera.Command, camera.Name))
		}
		sb.WriteString("tudo - tudo")

		b.Send(m.Sender, sb.String())
	})

	b.Start()
}
