package mocks

import "io"

// LoggerMock is a minimal logger for testing that does nothing.
type LoggerMock struct{}

func (LoggerMock) Level() string     { return "" }
func (LoggerMock) Writer() io.Writer { return nil }

func (LoggerMock) Info(string, ...interface{})  {}
func (LoggerMock) Error(string, ...interface{}) {}
func (LoggerMock) Warn(string, ...interface{})  {}
func (LoggerMock) Debug(string, ...interface{}) {}
func (LoggerMock) Fatal(string, ...interface{}) {}
