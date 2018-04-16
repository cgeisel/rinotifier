package main

import (
  "fmt"
	"github.com/aws/aws-sdk-go/aws"
  // "github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
  // "log"
)

func main() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	if err != nil {
		fmt.Println("Could not create session, exiting", err)
		return
	}

	_, err = sess.Config.Credentials.Get()
	if err != nil {
		fmt.Println("Could not get credentials, exiting", err)
		return
	}
  //
	// sess.Handlers.Send.PushFront(func(r *request.Request) {
	// 	// Log every request made and its payload
	// 	log.Println("Request: %s/%s, Payload: %s", r.ClientInfo.ServiceName, r.Operation, r.Params)
	// })

	// svc := ec2.New(sess, aws.NewConfig().WithLogLevel(aws.LogDebugWithHTTPBody))
  svc := ec2.New(sess)

  params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Environment"),
				Values: []*string{aws.String("prod")},
			},
			{
				Name:   aws.String("instance-state-name"),
				Values: []*string{aws.String("running"), aws.String("pending")},
			},
		},
	}

	subs, err := svc.DescribeReservedInstances(params)
	if err != nil {
		fmt.Println("Could not DescribeReservedInstances", err)
	}
	fmt.Println("Reserved instances", subs)
}
