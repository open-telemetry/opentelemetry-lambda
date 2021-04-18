package main

import (
	"encoding/json"
	"fmt"
	"log"
)

var logPrefix = fmt.Sprintf("[%s]", extensionName)

func prettyPrint(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return ""
	}
	return string(data)
}

func logln(a ...interface{}) {
	// TODO(jbd): When allocation here is a problem,
	// replace log with zap.
	args := []interface{}{logPrefix}
	args = append(args, a...)
	log.Println(args...)
}
