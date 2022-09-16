package modules

import (
	"regexp"
	"strings"
)

func GetSite(ip string, domain string, timeout int) []string {
	requestSite := RequestFunc("http://"+ip, domain, timeout)

	if len(requestSite) > 0 {
		re := regexp.MustCompile(`.*?<title>(.*?)</title>.*`)
		title := re.FindStringSubmatch(requestSite[2])
		if len(title) > 0 {
			explodeHttpCode := strings.Split(requestSite[0], " ")
			return []string{explodeHttpCode[0], requestSite[1], title[1]}
		}
	}

	return []string{}
}
