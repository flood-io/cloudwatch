package cloudwatch

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

// Throttling and limits from http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_limits.html
const (
	// The maximum rate of a GetLogEvents request is 10 requests per second per AWS account.
	readThrottle = time.Second / 10

	// The maximum rate of a PutLogEvents request is 5 requests per second per log stream.
	writeThrottle = time.Second / 5
)

// now is a function that returns the current time.Time. It's a variable so that
// it can be stubbed out in unit tests.
var now = time.Now

// Group wraps a log stream group and provides factory methods for creating
// readers and writers for streams.
type Group struct {
	group  string
	client *cloudwatchlogs.CloudWatchLogs
}

// AttachGroup creates a reference to a log group.
//
// If the group already exists, it is used.
// If the group doesn't exist, it is created.
func AttachGroup(group string, client *cloudwatchlogs.CloudWatchLogs) (*Group, error) {
	createLogGroupInput := &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(group),
	}
	_, err := client.CreateLogGroup(createLogGroupInput)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == cloudwatchlogs.ErrCodeResourceAlreadyExistsException {
				err = nil
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return &Group{
		group:  group,
		client: client,
	}, nil
}

// AttachStream creates a log stream in the group and returns an Writer for it.
//
// If the requested stream doesn't exist, it is created.
// If the requested stream already exists, the requested stream is used.
func (g *Group) AttachStream(stream string) (*Writer, error) {
	if _, err := g.client.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  &g.group,
		LogStreamName: &stream,
	}); err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == cloudwatchlogs.ErrCodeResourceAlreadyExistsException {
				err = nil
			}
		}
		if err != nil {
			return nil, err
		}
	}

	return NewWriter(g.group, stream, g.client), nil
}

// Open returns an Reader to read from the log stream.
func (g *Group) Open(stream string) (*Reader, error) {
	return NewReader(g.group, stream, g.client), nil
}
