package client

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	log "jwaoo.com/logger"
)

var authCache string = ""
var expiresAt time.Time

func tokenValide() bool {
	if authCache == "" {
		return false
	}
	now := time.Now().UTC()
	duration := expiresAt.Sub(now)
	if duration.Minutes() > 30 {
		return true
	}
	return false
}

func GetECRToken(regin string) string {
	if tokenValide() {
		return authCache
	}
	authCache = ""

	svc := ecr.New(session.New(), aws.NewConfig().WithRegion(regin))
	input := &ecr.GetAuthorizationTokenInput{}

	result, err := svc.GetAuthorizationToken(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ecr.ErrCodeServerException:
				log.Println(ecr.ErrCodeServerException, aerr.Error())
			case ecr.ErrCodeInvalidParameterException:
				log.Println(ecr.ErrCodeInvalidParameterException, aerr.Error())
			default:
				log.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Println(err.Error())
		}
		return ""
	}
	for _, v := range result.AuthorizationData {
		tokenBytes, err := base64.StdEncoding.DecodeString(aws.StringValue(v.AuthorizationToken))
		if err != nil {
			continue
		}
		av := strings.Split(string(tokenBytes), ":")
		if len(av) != 2 {
			continue
		}
		authCache = string(tokenBytes)
		expiresAt = *v.ExpiresAt
		log.Println(av[0])
		log.Println(av[1])
		log.Println(aws.StringValue(v.ProxyEndpoint))
		return authCache
		//c,_ := client.GetClient()
		//c.PullImageWithAuth("001082450179.dkr.ecr.us-west-1.amazonaws.com/zookeeper:v1", av[0], av[1]);
	}
	return ""

}
