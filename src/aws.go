package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

func newSession(cfg *config) (*sts.Credentials, error) {
	credsCache, err := loadCredentials(cfg.profile)
	if err != nil {
		return nil, err
	}

	if credsCache != nil {
		return credsCache, nil
	}

	token := os.Getenv("AWS_MFA_TOKEN")
	if token == "" {
		fmt.Print("Token: ")
		fmt.Scanln(&token)
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewSharedCredentials("", cfg.profile),
	})
	if err != nil {
		return nil, err
	}

	svc := sts.New(sess)
	creds, err := svc.GetSessionToken(&sts.GetSessionTokenInput{
		DurationSeconds: aws.Int64(cfg.duration),
		SerialNumber:    aws.String(cfg.mfaSerial),
		TokenCode:       aws.String(token),
	})
	if err != nil {
		return nil, err
	}

	return creds.Credentials, saveCredentials(cfg.profile, creds.Credentials)
}

func assumeRole(creds *sts.Credentials, cfg *config) (*sts.Credentials, error) {
	credsCache, err := loadCredentials(cfg.roleProfile)
	if err != nil {
		return nil, err
	}

	if credsCache != nil {
		return credsCache, nil
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			*creds.AccessKeyId,
			*creds.SecretAccessKey,
			*creds.SessionToken,
		),
	})
	if err != nil {
		return nil, err
	}

	svc := sts.New(sess)
	roleCreds, err := svc.AssumeRole(&sts.AssumeRoleInput{
		DurationSeconds: aws.Int64(cfg.duration),
		RoleSessionName: aws.String(fmt.Sprintf("aws-mfa-%s", cfg.roleProfile)),
		RoleArn:         aws.String(cfg.roleArn),
	})
	if err != nil {
		return nil, err
	}

	return roleCreds.Credentials, saveCredentials(cfg.roleProfile, roleCreds.Credentials)
}

func saveCredentials(profile string, creds *sts.Credentials) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dir := fmt.Sprintf("%s/.aws-mfa/credentials", home)

	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	bytes, err := json.Marshal(creds)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fmt.Sprintf("%s/%s.json", dir, profile), bytes, 0600)
}

func loadCredentials(profile string) (*sts.Credentials, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("%s/.aws-mfa/credentials/%s.json", home, profile)
	bytes, err := ioutil.ReadFile(path)
	if err == nil {
		var creds *sts.Credentials
		err = json.Unmarshal(bytes, &creds)
		if err != nil {
			return nil, err
		}

		return creds, nil
	}

	return nil, nil
}

func run(creds *sts.Credentials) error {
	path, err := exec.LookPath(os.Args[1])
	if err != nil {
		return err
	}

	env := os.Environ()
	env = append(
		env,
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", *creds.AccessKeyId),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", *creds.SecretAccessKey),
		fmt.Sprintf("AWS_SESSION_TOKEN=%s", *creds.SessionToken),
	)
	return syscall.Exec(path, os.Args[1:], env)
}
