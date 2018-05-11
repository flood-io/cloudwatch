package cloudwatch

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
)

const (
	// See: http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_PutLogEvents.html
	perEventBytes          = 26
	maximumBytesPerPut     = 1048576
	maximumLogEventsPerPut = 10000

	// See: http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_limits.html
	maximumBytesPerEvent = 262144 - perEventBytes

	dataAlreadyAcceptedCode  = "DataAlreadyAcceptedException"
	invalidSequenceTokenCode = "InvalidSequenceTokenException"
)

type RejectedLogEventsInfoError struct {
	Info *cloudwatchlogs.RejectedLogEventsInfo
}

func (e *RejectedLogEventsInfoError) Error() string {
	return fmt.Sprintf("log messages were rejected")
}

type WriterOptions struct {
	FlushEvery time.Duration
}

// Writer is an io.Writer implementation that writes lines to a cloudwatch logs
// stream.
type Writer struct {
	group, stream, sequenceToken *string

	client cloudwatchlogsiface.CloudWatchLogsAPI

	closed bool
	err    error

	events eventsBuffer

	flushTicker <-chan time.Time

	sync.Mutex // This protects calls to flush.
}

func NewWriter(group, stream string, client cloudwatchlogsiface.CloudWatchLogsAPI, opts WriterOptions) *Writer {
	w := &Writer{
		group:       aws.String(group),
		stream:      aws.String(stream),
		client:      client,
		flushTicker: time.Tick(opts.FlushEvery),
	}
	go w.start() // start flushing
	return w
}

// Write takes b, and creates cloudwatch log events for each individual line.
// If Flush returns an error, subsequent calls to Write will fail.
func (w *Writer) Write(b []byte) (int, error) {
	if w.closed {
		return 0, io.ErrClosedPipe
	}

	if w.err != nil {
		return 0, w.err
	}

	return w.buffer(b)
}

// starts continously flushing the buffered events.
func (w *Writer) start() error {
	for {
		// Exit if the stream is closed.
		if w.closed {
			return nil
		}

		<-w.flushTicker
		w.Flush()
	}
}

// Closes the writer. Any subsequent calls to Write will return
// io.ErrClosedPipe.
func (w *Writer) Close() {
	w.closed = true
	w.Flush() // Flush remaining buffer.
	return
}

// Flush flushes the events that are currently buffered.
func (w *Writer) Flush() {
	w.Lock()
	defer w.Unlock()

	events := w.events.drain()

	// No events to flush.
	if len(events) == 0 {
		return
	}

	w.flush(events)
	return
}

// flush flushes a slice of log events. This method should be called
// sequentially to ensure that the sequence token is updated properly.
func (w *Writer) flush(events []*cloudwatchlogs.InputLogEvent) {

	nextSequenceToken, err := w.putLogEvents(events, w.sequenceToken)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == dataAlreadyAcceptedCode {
				// already submitted, just grab the correct sequence token
				parts := strings.Split(awsErr.Message(), " ")
				nextSequenceToken = &parts[len(parts)-1]
				// TODO log locally...
				FallbackLogger.Errorln(
					"Data already accepted, ignoring error",
					"errorCode: ", awsErr.Code(),
					"message: ", awsErr.Message(),
					"logGroupName: ", *w.group,
					"logStreamName: ", *w.stream,
				)
				err = nil
			} else if awsErr.Code() == invalidSequenceTokenCode {
				// sequence code is bad, grab the correct one and retry
				parts := strings.Split(awsErr.Message(), " ")
				token := parts[len(parts)-1]
				nextSequenceToken, err = w.putLogEvents(events, &token)
			}
		}
	}

	if err != nil {
		w.err = err
		FallbackLogger.Errorln("error flushing", err)
	} else {
		w.sequenceToken = nextSequenceToken
	}

	// if resp.RejectedLogEventsInfo != nil {
	// w.err = &RejectedLogEventsInfoError{Info: resp.RejectedLogEventsInfo}
	// return w.err
	// }

	return
}

func (w *Writer) putLogEvents(events []*cloudwatchlogs.InputLogEvent, sequenceToken *string) (nextSequenceToken *string, err error) {
	resp, err := w.client.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
		LogEvents:     events,
		LogGroupName:  w.group,
		LogStreamName: w.stream,
		SequenceToken: sequenceToken,
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() != invalidSequenceTokenCode {
				FallbackLogger.Errorf(
					"Failed to put log: events: errorCode: %s message: %s, origError: %s log-group: %s log-stream: %s",
					awsErr.Code(),
					awsErr.Message(),
					awsErr.OrigErr(),
					*w.group,
					*w.stream,
				)
			}
		} else {
			FallbackLogger.Errorf("Failed to put log: %s", err)
		}

		return
	}

	nextSequenceToken = resp.NextSequenceToken
	return
}

// buffer splits up b into individual log events and inserts them into the
// buffer.
func (w *Writer) buffer(b []byte) (int, error) {
	r := bufio.NewReader(bytes.NewReader(b))

	var (
		n   int
		eof bool
	)

	for !eof {
		b, err := r.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				eof = true
			} else {
				break
			}
		}

		if len(b) == 0 {
			continue
		}

		w.events.add(&cloudwatchlogs.InputLogEvent{
			Message:   aws.String(string(b)),
			Timestamp: aws.Int64(now().UnixNano() / 1000000),
		})

		n += len(b)
	}

	return n, nil
}

// eventsBuffer represents a buffer of cloudwatch events that are protected by a
// mutex.
type eventsBuffer struct {
	sync.Mutex
	events []*cloudwatchlogs.InputLogEvent
}

func (b *eventsBuffer) add(event *cloudwatchlogs.InputLogEvent) {
	b.Lock()
	defer b.Unlock()

	b.events = append(b.events, event)
}

func (b *eventsBuffer) drain() []*cloudwatchlogs.InputLogEvent {
	b.Lock()
	defer b.Unlock()

	events := b.events[:]
	b.events = nil
	return events
}
