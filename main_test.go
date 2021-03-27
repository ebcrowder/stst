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

func TestMockGetSessionToken(t *testing.T) {
	api := &STSGetSessionTokenImpl{}

	input := &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int32(int32(900)),
		SerialNumber:    aws.String("serialNumber"),
		TokenCode:       aws.String("123456"),
	}

	resp, err := GetSessionToken(context.Background(), api, input)
	if err != nil {
		t.Error(err)
	}

	t.Log("SessionToken:      " + *resp.Credentials.SessionToken)
	t.Log("SecretAccessKey:  " + *resp.Credentials.SecretAccessKey)
	t.Log("AccessKeyId:  " + *resp.Credentials.AccessKeyId)
	t.Log("Expiration:  " + resp.Credentials.Expiration.Format(time.RFC3339))
}
