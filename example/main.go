package main

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/eltorocorp/cloudwatch"
)

func main() {
	config := &aws.Config{
		Region: aws.String("us-east-1"),
	}
	sess := session.Must(session.NewSession(config))

	g, err := cloudwatch.NewGroup("test-jc", cloudwatchlogs.New(sess))
	if err != nil {
		log.Fatal(err)
	}

	w, err := g.Attach("test-stream")
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(w, "some-prefix", log.Llongfile)

	logger.Println("Hi there. The current UTC time is ", time.Now().UTC())
}
