// nmon2influxdb
// import nmon data in InfluxDB
// author = adejoux@djouxtech.net

package nmon2influxdblib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
)

//
//helper functions

//CheckError check error message and display it
func CheckError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

// CheckInfo wrap info message
func CheckInfo(e error) {
	if e != nil {
		log.Printf("info: %s", e)
	}
}

//ReplaceComma replaces comma by html tabs tag
func ReplaceComma(s string) string {
	return "<tr><td>" + strings.Replace(s, ",", "</td><td>", 1) + "</td></tr>"
}

//PrintHTTPResponse print raw http response for debugging purpose
func PrintHTTPResponse(response *http.Response) {
	responseDump, err := httputil.DumpResponse(response, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(responseDump))
}

//PrintHTTPRequest print raw http request for debugging purpose
func PrintHTTPRequest(request *http.Request) {
	requestDump, err := httputil.DumpRequest(request, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(requestDump))
}

//PrintPrettyJSON helper used to display JSON output in a nicer way
func PrintPrettyJSON(contents []byte) {
	text := GetPrettyJSON(contents)
	log.Println("output:", string(text.Bytes()))
}

//GetPrettyJSON returns pretty json string
func GetPrettyJSON(contents []byte) bytes.Buffer {
	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, contents, "", "\t")
	if error != nil {
		log.Println("JSON parse error: ", error)
	}

	return prettyJSON
}
