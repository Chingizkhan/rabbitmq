package main

import (
	"github.com/fatih/color"
)

type Logger struct {
	Project, ConnectionName string
}

func NewLogger(project, connectionName string) Logger {
	return Logger{
		Project:        project,
		ConnectionName: connectionName,
	}
}

func (l Logger) Error(fnName, info string, err error) {
	color.Red("project: %s, connection: %s, error in func '%s': %s (%s)", l.Project, l.ConnectionName, fnName, info, err)
}

func (l Logger) Info(fnName, info string) {
	color.Blue("project: %s, connection: %s, func invoked '%s': %s", l.Project, l.ConnectionName, fnName, info)
}
