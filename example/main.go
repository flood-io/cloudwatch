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

	g, err := cloudwatch.AttachGroup("test-group", cloudwatchlogs.New(sess))
	if err != nil {
		log.Fatal(err)
	}

	w, err := g.AttachStream("test-stream")
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(w, "", log.Lshortfile)
	logger.Println("Woot", time.Now().UTC())
	err = w.Flush()
	if err != nil {
		log.Fatal(err)
	}
}
