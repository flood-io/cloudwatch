This is a Go library to treat CloudWatch Log streams as io.Writers and io.Readers.


## Usage

```go
group := NewGroup("group", cloudwatchlogs.New(defaults.DefaultConfig))
w, err := group.Attach("streamThatMightAlreadyExist")
fmt.Fprintln(w, "Hello World")
w.Flush()
```

or

```go
group := NewGroup("group", cloudwatchlogs.New(defaults.DefaultConfig))
w, err := group.Create("streamThatDoesntAlreadyExist")

io.WriteString(w, "Hello World")

r, err := group.Open("stream")
io.Copy(os.Stdout, r)
```

## Dependencies

This library depends on [aws-sdk-go](https://github.com/aws/aws-sdk-go/).
