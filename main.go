package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type STSGetSessionTokenAPI interface {
	GetSessionToken(ctx context.Context,
		params *sts.GetSessionTokenInput,
		optFns ...func(*sts.Options)) (*sts.GetSessionTokenOutput, error)
}

type AWSConfig struct {
	mfaSerial string
	region    string
}

func (c *AWSConfig) initConfigValues(awsConfigLines []string) {
	mfaSerial := findValueInFile(awsConfigLines, "mfa_serial")
	region := findValueInFile(awsConfigLines, "region")
	c.mfaSerial = mfaSerial
	c.region = region
	if len(c.mfaSerial) == 0 || len(c.region) == 0 {
		panic("Could not locate mfa_serial and/or region in aws config file:")
	}
}

func GetSessionToken(c context.Context, api STSGetSessionTokenAPI, input *sts.GetSessionTokenInput) (*sts.GetSessionTokenOutput, error) {
	return api.GetSessionToken(c, input)
}

func openAndReadFile(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		panic("Could not open file:" + err.Error())
	}
	defer file.Close()
	input, err := ioutil.ReadAll(file)
	if err != nil {
		panic("Could not read file:" + err.Error())
	}
	return strings.Split(string(input), "\n")
}

func generateCredentialsText(profile string, lines []string, response *sts.GetSessionTokenOutput) []string {
	updatedLines := make([]string, 0)
	existingCredentialsIndex := 0
	temporaryCredentials := generateTemporaryCredentials(profile, response)

	// check for existing credentials
	for i, line := range lines {
		if strings.Contains(line, profile) {
			existingCredentialsIndex = i
		}
	}
	// if the credentials exist, replace them entirely
	if existingCredentialsIndex != 0 {
		for i, line := range lines {
			if i <= existingCredentialsIndex {
				updatedLines = append(updatedLines, line)
			}
		}
		updatedLines[existingCredentialsIndex] = temporaryCredentials
	} else {
		// otherwise set the credentials
		updatedLines = append(updatedLines, lines...)
		updatedLines = append(updatedLines, "")
		updatedLines = append(updatedLines, temporaryCredentials)
	}
	return updatedLines
}

func generateTemporaryCredentials(profile string, response *sts.GetSessionTokenOutput) string {
	temporaryCredentials := "[" + profile + "]\n" +
		"aws_access_key_id" + " = " + *response.Credentials.AccessKeyId + "\n" +
		"aws_secret_access_key" + " = " + *response.Credentials.SecretAccessKey + "\n" +
		"aws_security_token" + " = " + *response.Credentials.SessionToken + "\n" +
		"aws_token_expiration" + " = " + response.Credentials.Expiration.Format(time.RFC3339) + "\n"
	return temporaryCredentials
}

func findValueInFile(lines []string, value string) string {
	var match string

	for _, line := range lines {
		if strings.Contains(line, value) {
			s := strings.SplitAfter(line, value+" = ")
			match = s[1]
		}
	}
	return match
}

func getTokenCodeFromStdIn() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter your MFA Code: ")
	tokenCode, _ := reader.ReadString('\n')
	tokenCode = strings.TrimSuffix(tokenCode, "\n")
	return tokenCode
}

func main() {
	// parse command line args and set defaults
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic("User home directory could not be determined:, " + err.Error())
	}
	awsDir := userHomeDir + "/.aws/"
	credentialsFile := flag.String("credentials", awsDir+"credentials", "Path to aws credentials file.")
	configFile := flag.String("config", awsDir+"config", "Path to aws config file.")
	profile := flag.String("profile", "temp", "Profile name used to associate temporary credentials.")
	durationSeconds := flag.Int("duration", 900, "Duration in seconds that temporary credentials should remain valid.")
	flag.Parse()

	// read AWS credentials file
	awsCredentialsLines := openAndReadFile(*credentialsFile)

	// obtain required values from AWS config file
	awsConfigLines := openAndReadFile(*configFile)
	awsConfig := AWSConfig{}
	awsConfig.initConfigValues(awsConfigLines)

	// check for existing expiration value
	// if temporary credentials have expired, fetch MFA code from stdin
	// otherwise, print that the credentials have not expired and exit
	var tokenCode string
	expiration := findValueInFile(awsCredentialsLines, "aws_token_expiration")
	if len(expiration) > 0 {
		now := time.Now()
		parsedExpiration, err := time.Parse(time.RFC3339, expiration)
		if err != nil {
			panic("Error parsing expiration:, " + err.Error())
		}
		if now.After(parsedExpiration) {
			tokenCode = getTokenCodeFromStdIn()
		} else {
			fmt.Print("Temporary credentials have not expired and remain valid.\n")
			return
		}
	} else {
		tokenCode = getTokenCodeFromStdIn()
	}

	// init AWS SDK config and sts client
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(awsConfig.region))
	if err != nil {
		panic("AWS SDK configuration error:, " + err.Error())
	}
	client := sts.NewFromConfig(cfg)

	// get session token and update the credentials file
	// with the temporary credentials
	response, err := GetSessionToken(context.Background(), client, &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int32(int32(*durationSeconds)),
		SerialNumber:    aws.String(awsConfig.mfaSerial),
		TokenCode:       aws.String(tokenCode),
	})
	if err != nil {
		panic("Could not get session token:" + err.Error())
	}

	credentialsText := generateCredentialsText(*profile, awsCredentialsLines, response)
	output := strings.Join(credentialsText, "\n")

	err = ioutil.WriteFile(*credentialsFile, []byte(output), 0644)
	if err != nil {
		panic("Could not write to aws credentials file:" + err.Error())
	}
	fmt.Print("Successfully authenticated with AWS STS and updated the AWS credentials file at: ", awsDir+"credentials\n")
}
