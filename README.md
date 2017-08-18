This is a Go library to treat CloudWatch Log streams as io.Writers and io.Readers.


## Usage

```go
group := AttachGroup("group", cloudwatchlogs.New(defaults.DefaultConfig))
w, err := group.AttachStream("stream")

fmt.Fprintln(w, "Hello World")

w.Flush()
```

or

```go
group := AttachGroup("group", cloudwatchlogs.New(defaults.DefaultConfig))
w, err := group.AttachStream("stream")

io.WriteString(w, "Hello World")

r, err := group.Open("stream")
io.Copy(os.Stdout, r)
```

## Dependencies

This library depends on [aws-sdk-go](https://github.com/aws/aws-sdk-go/).
