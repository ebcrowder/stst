package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
)

type STSGetSessionTokenImpl struct{}

func (dt STSGetSessionTokenImpl) GetSessionToken(ctx context.Context,
	params *sts.GetSessionTokenInput,
	optFns ...func(*sts.Options)) (*sts.GetSessionTokenOutput, error) {

	credentials := types.Credentials{
		AccessKeyId:     aws.String("accessKeyId"),
		SecretAccessKey: aws.String("secretAccessKey"),
		SessionToken:    aws.String("sessionToken"),
		Expiration:      aws.Time(time.Date(2020, 01, 01, 00, 00, 00, 00, time.UTC)),
	}

	output := &sts.GetSessionTokenOutput{
		Credentials: &credentials,
	}

	return output, nil
}

func callMockSTS() (*sts.GetSessionTokenOutput, error) {
	api := &STSGetSessionTokenImpl{}

	input := &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int32(int32(900)),
		SerialNumber:    aws.String("serialNumber"),
		TokenCode:       aws.String("123456"),
	}

	response, err := GetSessionToken(context.Background(), api, input)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func TestMockGetSessionToken(t *testing.T) {
	response, err := callMockSTS()
	if err != nil {
		t.Error(err)
	}

	t.Log("SessionToken:      " + *response.Credentials.SessionToken)
	t.Log("SecretAccessKey:  " + *response.Credentials.SecretAccessKey)
	t.Log("AccessKeyId:  " + *response.Credentials.AccessKeyId)
	t.Log("Expiration:  " + response.Credentials.Expiration.Format(time.RFC3339))
}

func TestGenerateCredentialsText(t *testing.T) {
	response, err := callMockSTS()
	if err != nil {
		t.Error(err)
	}

	lines := []string{
		"[default]",
		"aws_access_key_id = accessKeyId",
		"aws_secret_access_key = secretAccessKey",
	}

	wantedLines := []string{
		"[default]",
		"aws_access_key_id = accessKeyId",
		"aws_secret_access_key = secretAccessKey",
		"",
		"[temp]",
		"aws_access_key_id = accessKeyId",
		"aws_secret_access_key = secretAccessKey",
		"aws_security_token = sessionToken",
		"aws_token_expiration = 2020-01-01T00:00:00Z",
		"",
	}

	credentialsText := generateCredentialsText(lines, response)

	got := strings.Join(credentialsText, "\n")
	want := strings.Join(wantedLines, "\n")

	if got != want {
		t.Errorf("got %s; want %s", got, want)
	}
}
