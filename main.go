package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"os"
	"strings"
)

var Region string       // TODO read from ~/.aws/config
var SerialNumber string // TODO read from ~/.aws/credentials

type STSGetSessionTokenAPI interface {
	GetSessionToken(ctx context.Context,
		params *sts.GetSessionTokenInput,
		optFns ...func(*sts.Options)) (*sts.GetSessionTokenOutput, error)
}

func GetSessionToken(c context.Context, api STSGetSessionTokenAPI, input *sts.GetSessionTokenInput) (*sts.GetSessionTokenOutput, error) {
	return api.GetSessionToken(c, input)
}

func main() {
	SerialNumber := os.Getenv("SERIAL_NUMBER")
	Region := os.Getenv("REGION")

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(Region))
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := sts.NewFromConfig(cfg)

	// get TokenCode from stdin
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter your 2FA Code: ")
	TokenCode, _ := reader.ReadString('\n')
	TokenCode = strings.TrimSuffix(TokenCode, "\n")

	sessionToken, err := GetSessionToken(context.TODO(), client, &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int32(900),
		SerialNumber:    aws.String(SerialNumber),
		TokenCode:       aws.String(TokenCode),
	})
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	fmt.Println(sessionToken)
}
