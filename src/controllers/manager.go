package controllers

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func WhoisParameters(ip string) (string, string) {
	app := "whois"
	cmd := fmt.Sprintf("%s %s |grep country -i -m 1 |cut -d ':' -f 2 |xargs", app, ip)
	country, err := exec.Command("bash", "-c", cmd).Output()
	cmd2 := fmt.Sprintf("%s %s |grep organization -i -m 1 |cut -d ':' -f 2 |xargs", app, ip)
	organization, err := exec.Command("bash", "-c", cmd2).Output()

	if err != nil {
		fmt.Println(err.Error())
		return "", ""
	}

	return strings.TrimSuffix(string(country), "\n"), strings.TrimSuffix(string(organization), "\n")
}

func GetHTMLParameters(url string) (string, string, error) {
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}

	// Convert HTML into goquery document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", "", err
	}

	logo := ""
	title := ""

	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		if name, _ := s.Attr("rel"); name == "shortcut icon" {
			logoURL, _ := s.Attr("href")
			logo += logoURL
		}
	})

	doc.Find("title").Each(func(i int, s *goquery.Selection) {
		title += s.Text()

	})

	return logo, title, nil
}
