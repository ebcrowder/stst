package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

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

func main() {
	CredentialsFile := os.Getenv("CREDENTIALS_FILE") // TODO set as default
	SerialNumber := os.Getenv("SERIAL_NUMBER")       // TODO read from aws config or credentials file
	Region := os.Getenv("REGION")                    // TODO read from aws config file
	var DurationSeconds int32 = 900                  // TODO handle default somehow

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(Region))
	if err != nil {
		panic("Configuration error, " + err.Error())
	}

	client := sts.NewFromConfig(cfg)

	// get TokenCode from stdin
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter your 2FA Code: ")
	TokenCode, _ := reader.ReadString('\n')
	TokenCode = strings.TrimSuffix(TokenCode, "\n")

	response, err := GetSessionToken(context.TODO(), client, &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int32(DurationSeconds),
		SerialNumber:    aws.String(SerialNumber),
		TokenCode:       aws.String(TokenCode),
	})
	if err != nil {
		panic("Could not get session token:" + err.Error())
	}

	file, err := os.Open(CredentialsFile)
	if err != nil {
		panic("Could not open aws credentials file:" + err.Error())
	}
	defer file.Close()

	input, err := ioutil.ReadAll(file)
	if err != nil {
		panic("Could not open aws credentials file:" + err.Error())
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, "[development]") {
			lines[i] = "[development]\n" + "aws_access_key_id" + " = " + *response.Credentials.AccessKeyId + "\n" + "aws_secret_access_key" + " = " + *response.Credentials.SecretAccessKey + "\n" + "aws_security_token" + " = " + *response.Credentials.SessionToken + "\n"
		} else {
			if i == len(lines)-1 {
				lines[i] = "[development]\n" + "aws_access_key_id" + " = " + *response.Credentials.AccessKeyId + "\n" + "aws_secret_access_key" + " = " + *response.Credentials.SecretAccessKey + "\n" + "aws_security_token" + " = " + *response.Credentials.SessionToken + "\n"
			}
		}
	}

	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(CredentialsFile, []byte(output), 0644)
	if err != nil {
		panic("Could not write to aws credentials file:" + err.Error())
	}
}
