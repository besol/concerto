package utils

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func ScrapeErrorMessage(message string, regExpression string) string {

	re, err := regexp.Compile(regExpression)
	scrapped := re.FindStringSubmatch(message)

	if scrapped == nil || err != nil || len(scrapped) < 2 {
		// couldn't scrape, return generic error
		message = "Error executing operation"
	} else {
		// return scrapped response
		message = scrapped[1]
	}

	return message
}

func CheckReturnCode(res int, mesg []byte) {
	if res >= 300 {

		message := string(mesg[:])
		log.Debugf("Concerto API response: %s", message)

		f := func(c rune) bool {
			return c == ',' || c == ':' || c == '{' || c == '}' || c == '"' || c == ']' || c == '['
		}

		// check if response is a web page.
		if strings.Contains(message, "<html>") {
			scrapResponse := "<title>(.*?)</title>"
			message = ScrapeErrorMessage(message, scrapResponse)
		} else if strings.Contains(message, "{\"errors\":{") {
			scrapResponse := "{\"errors\":(.*?)}"

			message = ScrapeErrorMessage(message, scrapResponse)
			result := strings.Split(message, ",")
			if result != nil && len(result) >= 1 {
				message = result[0]
			}
			// Separate into fields with func.
			fields := strings.FieldsFunc(message, f)
			message = strings.Join(fields[:], " ")

		} else if strings.Contains(message, "{\"error\":") {
			scrapResponse := "{\"error\":\"(.*?)\"}"
			message = ScrapeErrorMessage(message, scrapResponse)
		}

		// if it's not a web page or json-formatted message, return the raw message
		log.Fatal(fmt.Sprintf("HTTP request failed: [%s]", message))
	}
}

func Exists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}
