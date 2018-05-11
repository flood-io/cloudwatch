package cloudwatch

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Printf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Print(args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})

	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Println(args ...interface{})
	Warnln(args ...interface{})
	Warningln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})
	Panicln(args ...interface{})
}

var FallbackLogger Logger = NullLogger{}

type NullLogger struct{}

func (n NullLogger) Debugf(format string, args ...interface{})   {}
func (n NullLogger) Infof(format string, args ...interface{})    {}
func (n NullLogger) Printf(format string, args ...interface{})   {}
func (n NullLogger) Warnf(format string, args ...interface{})    {}
func (n NullLogger) Warningf(format string, args ...interface{}) {}
func (n NullLogger) Errorf(format string, args ...interface{})   {}
func (n NullLogger) Fatalf(format string, args ...interface{})   {}
func (n NullLogger) Panicf(format string, args ...interface{})   {}

func (n NullLogger) Debug(args ...interface{})   {}
func (n NullLogger) Info(args ...interface{})    {}
func (n NullLogger) Print(args ...interface{})   {}
func (n NullLogger) Warn(args ...interface{})    {}
func (n NullLogger) Warning(args ...interface{}) {}
func (n NullLogger) Error(args ...interface{})   {}
func (n NullLogger) Fatal(args ...interface{})   {}
func (n NullLogger) Panic(args ...interface{})   {}

func (n NullLogger) Debugln(args ...interface{})   {}
func (n NullLogger) Infoln(args ...interface{})    {}
func (n NullLogger) Println(args ...interface{})   {}
func (n NullLogger) Warnln(args ...interface{})    {}
func (n NullLogger) Warningln(args ...interface{}) {}
func (n NullLogger) Errorln(args ...interface{})   {}
func (n NullLogger) Fatalln(args ...interface{})   {}
func (n NullLogger) Panicln(args ...interface{})   {}
