package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/julienschmidt/httprouter"
)

const (
	// VERSION is the Application Version
	VERSION = "0.1.0"
	// EmailRegexp is a regular expression to validate email addresses
	EmailRegexp = `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`
)

// ApplicationConfig represents the configuration that is filled from the config file
type ApplicationConfig struct {
	Port             string            `json:"port"`
	SMTPHost         string            `json:"smtpHost"`
	SMTPPort         string            `json:"smtpPort"`
	SMTPAuthUser     string            `json:"smtpAuthUser"`
	SMTPAuthPassword string            `json:"smtpAuthPassword"`
	RecipientMap     map[string]string `json:"recipients"`
	Lifetime         int               `json:"lifetime"`
	CleanupInterval  int               `json:"cleanupInterval"`
	TarpitInterval   int               `json:"tarpitInterval"`
}

// validateConfig validates the configuration, currently simply validates the email addresses
func (c *ApplicationConfig) validateConfig() error {
	Re := regexp.MustCompile(EmailRegexp)
	for _, v := range c.RecipientMap {
		if !Re.MatchString(v) {
			return fmt.Errorf("config Error: not a email address: %v", v)
		}
	}
	return nil
}

// loadConfig loads the configuration from the provided config file
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

// printVersion prints out version number, and commit id if a commit file is found. Then it exits with 0
func printVersion() {
	// try to read a commit file, ignore errors if we do not have one
	raw, _ := ioutil.ReadFile("COMMIT")
	var commit string
	if len(raw) > 0 {
		commit = fmt.Sprintf(" (commit %s)", string(raw))
	}
	log.Printf("Mailbridge Version %s%s", VERSION, commit)
	os.Exit(0)
}

// main starts the application
func main() {

	var configFile = flag.String("configFile", "config.json", "Configuration File")
	var versionAndExit = flag.Bool("version", false, "print application version and exit")
	flag.Parse()

	// print only version and exit
	if *versionAndExit {
		printVersion()
	}

	// try to get config file
	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Could not read Configuration: %v", err)
	}

	// ok, start the router
	router := httprouter.New()

	// initialize mail server and map of active tokens
	mailServer := InitMailServer(config)
	activeTokens := InitActiveTokens(config)
	tarpit := InitTarpit(config)

	// initialize the Controller
	c := InitController(mailServer, activeTokens, tarpit)

	// now set up the router
	router.GET("/api/token", c.GetToken)
	router.POST("/api/send", c.SendMail)
	log.Fatal(http.ListenAndServe(":"+config.Port, router))
}
