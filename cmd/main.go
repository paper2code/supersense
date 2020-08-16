package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/minskylab/supersense"
	"github.com/minskylab/supersense/persistence"
	"github.com/minskylab/supersense/server"
	"github.com/minskylab/supersense/sources"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)
	_ = godotenv.Load() // loading .env vars

	port := os.Getenv("PORT")

	ctx := context.TODO()

	dummySource, err := sources.NewDummy(5*time.Minute, "Hello World")
	if err != nil {
		log.Panic(err)
	}

	authToken := os.Getenv("GITHUB_TOKEN")

	githubSource, err := sources.NewGithub([]string{"minskylab/supersense"}, &authToken)
	if err != nil {
		log.Panic(err)
	}

	twitterSource, err := sources.NewTwitter(sources.TwitterClientProps{
		ConsumerKey:    os.Getenv("CONSUMER_KEY"),
		ConsumerSecret: os.Getenv("CONSUMER_SECRET"),
		AccessToken:    os.Getenv("ACCESS_TOKEN"),
		AccessSecret:   os.Getenv("ACCESS_SECRET"),
		QueryToTrack:   []string{"#peru"},
	})
	if err != nil {
		log.Panic(err)
	}

	mux, err := supersense.NewMux(ctx, twitterSource, dummySource, githubSource)
	if err != nil {
		log.Panic(err)
	}

	go func() {
		for event := range mux.Events() {
			maxLength := 32
			cutMessage := event.Message
			if len(event.Message) > maxLength {
				cutMessage = cutMessage[:maxLength]
				cutMessage = strings.Replace(cutMessage, "\n", " ", -1)
				cutMessage = strings.Trim(cutMessage, "\n ")
			}
			log.Infof(
				"[%s] %s: %s | by: %s",
				event.SourceName,
				event.Title,
				cutMessage,
				event.Actor.Name,
			)
		}
	}()

	if err := mux.RunAllSources(ctx); err != nil {
		log.Panic(err)
	}

	per, err := persistence.New("./supersense.db", []byte("example"))
	if err != nil {
		log.Panic(err)
	}

	instance, err := server.New(mux, per)
	if err != nil {
		log.Panic(err)
	}

	instance.LaunchServer(port)
}
