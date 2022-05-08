package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmlogrus"
)

var log = &logrus.Logger{                                                    
	Out:   os.Stderr,
	Hooks: make(logrus.LevelHooks),                                  
	Level: logrus.DebugLevel,                                        
	Formatter: &logrus.JSONFormatter{                                
		FieldMap: logrus.FieldMap{                               
			logrus.FieldKeyTime:  "@timestamp",              
			logrus.FieldKeyLevel: "log.level",               
			logrus.FieldKeyMsg:   "message",                 
			logrus.FieldKeyFunc:  "function.name", // non-ECS
		},                                                       
	},                                                               
}                                                                        

func init() {
    apm.DefaultTracer.SetLogger(log)
	log.AddHook(&apmlogrus.Hook{})
}
