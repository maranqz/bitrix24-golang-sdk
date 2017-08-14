package bitrix24

type Logger interface {
	Check(lvl interface{}, msg string) interface{}
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Panic(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	Sync() error
}
