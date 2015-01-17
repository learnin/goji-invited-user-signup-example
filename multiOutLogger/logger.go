package multiOutLogger

type Logger interface {
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})
	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
	Panicln(v ...interface{})
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})

	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Warningf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Warning(v ...interface{})
	Error(v ...interface{})
	Debugln(v ...interface{})
	Infoln(v ...interface{})
	Warnln(v ...interface{})
	Warningln(v ...interface{})
	Errorln(v ...interface{})
}

type MultiOutLogger struct {
	loggers []Logger
}

func New(loggers ...Logger) *MultiOutLogger {
	return &MultiOutLogger{loggers: loggers}
}

func (l *MultiOutLogger) Fatal(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Fatal(v)
	}
}

func (l *MultiOutLogger) Fatalf(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Fatalf(format, v)
	}
}

func (l *MultiOutLogger) Fatalln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Fatalln(v)
	}
}

func (l *MultiOutLogger) Panic(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Panic(v)
	}
}

func (l *MultiOutLogger) Panicf(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Panicf(format, v)
	}
}

func (l *MultiOutLogger) Panicln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Panicln(v)
	}
}

func (l *MultiOutLogger) Print(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Print(v)
	}
}

func (l *MultiOutLogger) Printf(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Printf(format, v)
	}
}

func (l *MultiOutLogger) Println(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Println(v)
	}
}

func (l *MultiOutLogger) Debugf(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Debugf(format, v)
	}
}

func (l *MultiOutLogger) Infof(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Infof(format, v)
	}
}

func (l *MultiOutLogger) Warnf(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Warnf(format, v)
	}
}

func (l *MultiOutLogger) Warningf(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Warningf(format, v)
	}
}

func (l *MultiOutLogger) Errorf(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Errorf(format, v)
	}
}

func (l *MultiOutLogger) Debug(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Debug(v)
	}
}

func (l *MultiOutLogger) Info(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Info(v)
	}
}

func (l *MultiOutLogger) Warn(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Warn(v)
	}
}

func (l *MultiOutLogger) Warning(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Warning(v)
	}
}

func (l *MultiOutLogger) Error(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Error(v)
	}
}

func (l *MultiOutLogger) Debugln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Debugln(v)
	}
}

func (l *MultiOutLogger) Infoln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Infoln(v)
	}
}

func (l *MultiOutLogger) Warnln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Warnln(v)
	}
}

func (l *MultiOutLogger) Warningln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Warningln(v)
	}
}

func (l *MultiOutLogger) Errorln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Errorln(v)
	}
}
