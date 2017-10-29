package main

import (
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"encoding/json"
	"flag"
	"io/ioutil"
	"fmt"
	"regexp"
	"errors"
)

const VERSION = "0.0.1"

type ApplicationConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
	AuthUser string `json:"authUser"`
	AuthPassword string `json:"authPassword"`
	RecipientMap map[string]string `json:"recipients"`
	Lifetime int `json:"lifetime"`
	CleanupInterval int `json:"cleanupInterval"`
	TarpitInterval int `json:"tarpitInterval"`
}

func (c *ApplicationConfig) validateConfig() error {
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	for _, v := range c.RecipientMap {
		if ! Re.MatchString(v) {
			return errors.New(fmt.Sprintf("Config Error: not a email address: %v", v))
		}
	}
	return nil
}

func loadConfig(fileName *string) (*ApplicationConfig, error) {
	//filename is the path to the json config file
	var config ApplicationConfig
	raw, err := ioutil.ReadFile(*fileName)
	if err != nil {
		return &config, err
	}

	err = json.Unmarshal(raw, &config)
	if err != nil {
		return &config, err
	}

	err = config.validateConfig()
	if err != nil {
		return &config, err
	}

	return &config, nil
}

func main() {

	var configFile = flag.String("configFile", "config.json", "Configuration File")
	flag.Parse()

	config, err := loadConfig( configFile )
	if err != nil {
		log.Fatalf("Could not read Configuration: %v", err)
	}

	router := httprouter.New()

	// initialize mail server and map of active tokens
	mailServer := InitMailServer(config)
	activeTokens := InitActiveTokens(config)
	tarpit := InitTarpit(config)

	// initialize the controller
	c := InitController(mailServer, activeTokens, tarpit)

	// now set up the router
	router.GET("/api/token", c.GetToken)
	router.POST("/api/send", c.SendMail)
	log.Fatal(http.ListenAndServe(":8081", router))
}
