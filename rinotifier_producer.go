package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sqs"
	"log"
	"os"
	"strconv"
	"time"
)

type ReservedInstance struct {
	End string
}

func riHandler() {
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

	var svc *ec2.EC2
	var sqs_svc *sqs.SQS

	debug, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil || !debug {
		svc = ec2.New(sess)
		sqs_svc = sqs.New(sess)
	} else {
		svc = ec2.New(sess, aws.NewConfig().WithLogLevel(aws.LogDebugWithHTTPBody))
		sqs_svc = sqs.New(sess, aws.NewConfig().WithLogLevel(aws.LogDebugWithHTTPBody))
	}

	// URL to our queue
	qURL := os.Getenv("QUEUE_URL")

	params := &ec2.DescribeReservedInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("state"),
				Values: []*string{aws.String("active")},
			},
		},
	}

	active_ris, err := svc.DescribeReservedInstances(params)
	if err != nil {
		log.Println("Could not DescribeReservedInstances, exiting", err)
		return
	}

	var expiring_ris []*ec2.ReservedInstances
	for _, ri := range active_ris.ReservedInstances {
		if isExpiring(ri, 365) {
			expiring_ris = append(expiring_ris, ri)
		}
	}
	fmt.Println(expiring_ris)
	if len(expiring_ris) > 0 {
		q, err := addToQueue(sqs_svc, qURL, expiring_ris)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(q)
	}
}

func isExpiring(ri *ec2.ReservedInstances, days int) bool {
	const ri_form = "2006-01-02 15:04:05 +0000"
	return ri.End.Before(time.Now().Add(time.Duration(days*24) * time.Hour))
}

func addToQueue(svc *sqs.SQS, qURL string, ri []*ec2.ReservedInstances) (string, error) {
	json_ri, err := json.Marshal(ri)
	if err != nil {
		log.Println("Could not marshal json, exiting", err)
		return "Could not marshal json", err
	}
	result, err := svc.SendMessage(&sqs.SendMessageInput{
		DelaySeconds: aws.Int64(10),
		MessageBody:  aws.String(string(json_ri)),
		QueueUrl:     &qURL,
	})

	return *result.MessageId, err
}

func main() {
	lambda.Start(riHandler)
}
