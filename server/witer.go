package main

import (
	"fmt"
	"strings"
)

type LogWriter struct {
	Format string
	Logger func(interface{})
}

func (l *LogWriter) Write(b []byte) (int, error) {
	l.Logger(fmt.Sprintf(l.Format, strings.TrimSpace(string(b))))
	return len(b), nil
}
