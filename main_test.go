package main

import "testing"

func TestMain_ValidateApplicationConfig(t *testing.T) {
	t.Parallel()
	config := &ApplicationConfig{
		RecipientMap: make(map[string]string),
	}
	config.RecipientMap["test1"] = "test1@example.com"
	config.RecipientMap["test2"] = "test2@example.com"
	if err := config.validateConfig(); err != nil {
		t.Errorf("Error in config validation but should not %v", err)
	}
	config.RecipientMap["wrong"] = "wrong_example.com"
	if err := config.validateConfig(); err == nil {
		t.Errorf("Error in config validation: should return error but does not")
	}
}

func TestMain_LoadConfigFile(t *testing.T) {
	t.Parallel()
	file := "sample_config.json"
	config, err := loadConfig(&file)
	if err != nil {
		t.Errorf("Error in config file loading but should not %v", err)
	}
	if config.SMTPHost != "mail.example.com" {
		t.Errorf("Error in loadConfig: %v is %v but should be %v", "SMTPHost", config.SMTPHost, "mail.example.com")
	}
	if config.SMTPPort != "25" {
		t.Errorf("Error in loadConfig: %v is %v but should be %v", "SMTPPort", config.SMTPPort, "25")
	}
	if config.Port != "8081" {
		t.Errorf("Error in loadConfig: %v is %v but should be %v", "Port", config.SMTPPort, "8081")
	}
	if config.SMTPAuthUser != "SMTP_USER" {
		t.Errorf("Error in loadConfig: %v is %v but should be %v", "AuthUser", config.SMTPAuthUser, "SMTP_USER")
	}
	if config.SMTPAuthPassword != "SMTP_PASSWORD" {
		t.Errorf("Error in loadConfig: %v is %v but should be %v", "AuthPassword", config.SMTPAuthPassword, "SMTP_PASSWORD")
	}
	if config.Lifetime != 60 {
		t.Errorf("Error in loadConfig: %v is %v but should be %v", "Lifetime", config.Lifetime, 60)
	}
	if config.CleanupInterval != 10 {
		t.Errorf("Error in loadConfig: %v is %v but should be %v", "CleanupInterval", config.CleanupInterval, 10)
	}
	if config.TarpitInterval != 10 {
		t.Errorf("Error in loadConfig: %v is %v but should be %v", "TarpitInterval", config.TarpitInterval, 10)
	}
	if len(config.RecipientMap) != 2 {
		t.Errorf("Error in loadConfig: %v is %v but should be %v", "length of recipient map", len(config.RecipientMap), 2)
	}
}
