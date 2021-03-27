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

func GetSessionToken(c context.Context, api STSGetSessionTokenAPI, input *sts.GetSessionTokenInput) (*sts.GetSessionTokenOutput, error) {
	return api.GetSessionToken(c, input)
}

func openAndReadFile(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		panic("Could not open file:" + err.Error())
	}
	defer file.Close()
	input, err := ioutil.ReadAll(file)
	if err != nil {
		panic("Could not read file:" + err.Error())
	}
	return input
}

func generateCredentialsText(path, temporaryCredentials string) []string {
	input := openAndReadFile(path)
	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, "[temp]") {
			lines[i] = temporaryCredentials
		}
	}
	return lines
}

func generateTemporaryCredentials(response *sts.GetSessionTokenOutput) string {
	temporaryCredentials := "[temp]\n" +
		"aws_access_key_id" + " = " + *response.Credentials.AccessKeyId + "\n" +
		"aws_secret_access_key" + " = " + *response.Credentials.SecretAccessKey + "\n" +
		"aws_security_token" + " = " + *response.Credentials.SessionToken + "\n" +
		"aws_token_expiration" + " = " + response.Credentials.Expiration.Format(time.RFC3339)
	return temporaryCredentials
}

func findConfigValue(path, value string) string {
	var match string
	input := openAndReadFile(path)
	lines := strings.Split(string(input), "\n")

	for _, line := range lines {
		if strings.Contains(line, value) {
			s := strings.SplitAfter(line, value+" = ")
			match = s[1]
		}
	}

	if len(match) == 0 {
		panic("Could not locate value in aws config file:" + value)
	}

	return match
}

func getTokenCodeFromStdIn() string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Please enter your 2FA Code: ")

	TokenCode, _ := reader.ReadString('\n')
	TokenCode = strings.TrimSuffix(TokenCode, "\n")
	return TokenCode
}

func main() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic("User home directory could not be determined:, " + err.Error())
	}
	awsDir := userHomeDir + "/.aws/"
	var credentialsFile = flag.String("credentials", awsDir+"credentials", "Path to aws credentials file")
	var configFile = flag.String("config", awsDir+"config", "Path to aws config file")
	var durationSeconds = flag.Int("duration", 900, "Duration in seconds that temporary credentials should remain valid")
	flag.Parse()

	SerialNumber := findConfigValue(*configFile, "mfa_serial")
	Region := findConfigValue(*configFile, "region")

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(Region))
	if err != nil {
		panic("Configuration error, " + err.Error())
	}

	client := sts.NewFromConfig(cfg)

	TokenCode := getTokenCodeFromStdIn()

	response, err := GetSessionToken(context.Background(), client, &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int32(int32(*durationSeconds)),
		SerialNumber:    aws.String(SerialNumber),
		TokenCode:       aws.String(TokenCode),
	})
	if err != nil {
		panic("Could not get session token:" + err.Error())
	}

	temporaryCredentials := generateTemporaryCredentials(response)
	lines := generateCredentialsText(*credentialsFile, temporaryCredentials)
	output := strings.Join(lines, "\n")

	err = ioutil.WriteFile(*credentialsFile, []byte(output), 0644)
	if err != nil {
		panic("Could not write to aws credentials file:" + err.Error())
	}
}
