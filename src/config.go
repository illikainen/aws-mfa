package main

import (
	"fmt"
	"os"

	"gopkg.in/ini.v1"
)

type config struct {
	duration    int64
	mfaSerial   string
	profile     string
	roleProfile string
	roleArn     string
}

func readConfig() (*config, error) {
	config := config{
		duration: 900,
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	aws, err := ini.Load(fmt.Sprintf("%s/.aws/credentials", home))
	if err != nil {
		return nil, err
	}

	config.roleProfile = os.Getenv("AWS_PROFILE")
	if config.roleProfile == "" {
		return nil, fmt.Errorf("AWS_PROFILE must be set")
	}

	config.roleArn = aws.Section(config.roleProfile).Key("role_arn").MustString("")
	if config.roleArn != "" {
		config.profile = aws.Section(config.roleProfile).
			Key("source_profile").MustString("")
		if config.profile == "" {
			return nil, fmt.Errorf("%s: missing source_profile", config.roleProfile)
		}
	} else {
		config.profile = config.roleProfile
	}

	config.mfaSerial = aws.Section(config.profile).Key("mfa_serial").MustString("")
	if config.mfaSerial == "" {
		return nil, fmt.Errorf("%s: missing mfa_serial", config.profile)
	}

	mfa, err := ini.Load(fmt.Sprintf("%s/.aws-mfa/config", home))
	if err == nil {
		config.duration = mfa.Section("").Key("duration").RangeInt64(900, 900, 129600)
	}

	return &config, nil
}
