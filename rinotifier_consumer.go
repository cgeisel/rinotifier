package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"log"
	"os"
	"strconv"
)

func riHandler(ctx context.Context, sqsEvent events.SQSEvent) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		log.Println("Could not create session, exiting", err)
		return
	}

	_, err = sess.Config.Credentials.Get()
	if err != nil {
		log.Println("Could not get credentials, exiting", err)
		return
	}

	// URL to our topic
	sns_topic := os.Getenv("SNS_TOPIC")
	var sns_svc *sns.SNS

	debug, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil || !debug {
		sns_svc = sns.New(sess)
	} else {
		sns_svc = sns.New(sess, aws.NewConfig().WithLogLevel(aws.LogDebugWithHTTPBody))
	}

	for _, message := range sqsEvent.Records {
		log.Printf("The message %s for event source %s = %s \n", message.MessageId, message.EventSource, message.Body)
		publish(sns_svc, sns_topic, message.Body)
	}
}

func publish(sns_svc *sns.SNS, sns_topic string, message string) {
	response, err := sns_svc.Publish(&sns.PublishInput{
		TopicArn: &sns_topic,
		Message:  aws.String(message),
	})

	if err != nil {
		log.Println("SNS Error: ", err)
	}

	log.Println(response)
}

func main() {
	lambda.Start(riHandler)
}
