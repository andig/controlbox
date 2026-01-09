package main

import (
	"fmt"
	"time"
)

// Logging interface

func (h *controlbox) Trace(args ...interface{}) {
	// h.print("TRACE", args...)
}

func (h *controlbox) Tracef(format string, args ...interface{}) {
	// h.printFormat("TRACE", format, args...)
}

func (h *controlbox) Debug(args ...interface{}) {
	// h.print("DEBUG", args...)
}

func (h *controlbox) Debugf(format string, args ...interface{}) {
	// h.printFormat("DEBUG", format, args...)
}

func (h *controlbox) Info(args ...interface{}) {
	h.print("INFO ", args...)
}

func (h *controlbox) Infof(format string, args ...interface{}) {
	h.printFormat("INFO ", format, args...)
}

func (h *controlbox) Error(args ...interface{}) {
	h.print("ERROR", args...)
}

func (h *controlbox) Errorf(format string, args ...interface{}) {
	h.printFormat("ERROR", format, args...)
}

func (h *controlbox) currentTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func (h *controlbox) print(msgType string, args ...interface{}) {
	value := fmt.Sprintln(args...)
	fmt.Printf("%s %s %s", h.currentTimestamp(), msgType, value)
}

func (h *controlbox) printFormat(msgType, format string, args ...interface{}) {
	value := fmt.Sprintf(format, args...)
	fmt.Println(h.currentTimestamp(), msgType, value)
}
