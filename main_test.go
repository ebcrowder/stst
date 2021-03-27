package main

import (
	"context"
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
		Expiration:      aws.Time(time.Now()),
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

func TestGenerateTemporaryCredentials(t *testing.T) {
	response, err := callMockSTS()
	if err != nil {
		t.Error(err)
	}

	got := generateTemporaryCredentials(response)
	want := "[temp]\n" +
		"aws_access_key_id" + " = " + *response.Credentials.AccessKeyId + "\n" +
		"aws_secret_access_key" + " = " + *response.Credentials.SecretAccessKey + "\n" +
		"aws_security_token" + " = " + *response.Credentials.SessionToken + "\n" +
		"aws_token_expiration" + " = " + response.Credentials.Expiration.Format(time.RFC3339)

	if got != want {
		t.Errorf("got %s; want %s", got, want)
	}
}
