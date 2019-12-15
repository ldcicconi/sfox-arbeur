package main

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

var awsSFOXAPIKeysID = "SFOX_API_KEYS_EC2"

func getAPIKeysFromAWSSecrets() ([]string, error) {
	awsSession := session.Must(session.NewSession())
	awsClient := secretsmanager.New(awsSession, aws.NewConfig().WithRegion("us-east-2"))
	secretRequest := secretsmanager.GetSecretValueInput{
		SecretId: &awsSFOXAPIKeysID,
	}
	response, err := awsClient.GetSecretValue(&secretRequest)
	if err != nil {
		return nil, err
	}
	return strings.Split(*response.SecretString, ","), nil
}
