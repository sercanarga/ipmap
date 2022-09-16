package modules

import (
	"regexp"
)

func GetDomainTitle(url string) []string {
	getTitle := RequestFunc("http://"+url, url, 5000)
	re := regexp.MustCompile(`.*?<title>(.*?)</title>.*`)

	if len(getTitle) > 0 {
		match := re.FindStringSubmatch(getTitle[2])
		return []string{match[1], getTitle[3]}
	}

	return []string{}
}
