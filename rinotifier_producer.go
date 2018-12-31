package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sqs"
	"log"
	"time"
)

type ReservedInstance struct {
	End string
}

func main() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
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

	svc := ec2.New(sess, aws.NewConfig().WithLogLevel(aws.LogDebugWithHTTPBody))
	///svc := ec2.New(sess)

	sqs_svc := sqs.New(sess)

	// URL to our queue
	// daily_qURL := "https://sqs.us-west-2.amazonaws.com/268368477672/rinotifier-test"
	weekly_qURL := "https://sqs.us-west-2.amazonaws.com/268368477672/rinotifier-test"

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
		log.Println("Could not DescribeReservedInstances", err)
	}
	for i, ri := range active_ris.ReservedInstances {
		if isExpiring(ri, 365) {
			// log.Println(ri)
			q, err := addToQueue(sqs_svc, weekly_qURL, ri)
			if err != nil {
				log.Println(err)
			}
			log.Println(i)
			log.Println(q)
		}
	}
}

func isExpiring(ri *ec2.ReservedInstances, days int) bool {
	const ri_form = "2006-01-02 15:04:05 +0000"
	return ri.End.Before(time.Now().Add(time.Duration(days*24) * time.Hour))
}

func addToQueue(svc *sqs.SQS, qURL string, ri *ec2.ReservedInstances) (string, error) {
	result, err := svc.SendMessage(&sqs.SendMessageInput{
		DelaySeconds: aws.Int64(10),
		MessageBody:  aws.String(ri.String()),
		QueueUrl:     &qURL,
	})

	return *result.MessageId, err
}
