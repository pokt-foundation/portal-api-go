package logger

// LogObject set any object as a loggable object
type LogObject struct {
	Name   string
	Fields map[string]interface{}
}

// LogName returns the name of the object to log
func (l *LogObject) LogName() string {
	return l.Name
}

// LogProperties returns the fields to log
func (l *LogObject) LogProperties() map[string]interface{} {
	return l.Fields
}

func MapObject(logName string, data map[string]interface{}) LoggerInterface {
	return &LogObject{
		Name:   logName,
		Fields: data,
	}
}
