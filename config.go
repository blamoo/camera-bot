package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"gopkg.in/tucnak/telebot.v2"
	tb "gopkg.in/tucnak/telebot.v2"
)

var appConfig Config

type Camera struct {
	Name    string
	Url     string
	Command string
}

type Config struct {
	Debug          bool
	TelegramApiKey string
	Users          map[int]string
	Cameras        []Camera
}

func (this *Camera) SendPhotoTo(b *telebot.Bot, to *telebot.User) {
	client, err := http.Get(this.Url)

	if err != nil {
		b.Send(to, err.Error())
		return
	}

	defer client.Body.Close()

	photo := &tb.Photo{
		File: tb.File{
			FileReader: client.Body,
		},
	}

	_, err = b.Send(to, photo)

	if err != nil {
		b.Send(to, err.Error())
		return
	}
}

func InitializeConfig() error {
	var err error
	configFile, err := os.Open(configPath)

	if err != nil {
		return err
	}

	configBytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(configBytes, &appConfig)
	if err != nil {
		return err
	}

	return nil
}
