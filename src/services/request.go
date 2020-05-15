package services

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func GetHTMLParameters(url string) (string, string, error) {
	var err error
	defer func() {
			if err != nil {
				fmt.Printf("Error getting HTTML: %v", err)
			}
	}()

	resp, err := http.Get(url)
	if err != nil { return  "", "", err}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil { return "", "", err}

	logo := ""
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		if name, _ := s.Attr("rel"); name == "shortcut icon" {
			logoURL, _ := s.Attr("href")
			logo += logoURL
		}
	})

	title := ""
	doc.Find("title").Each(func(i int, s *goquery.Selection) {
		title += s.Text()
	})

	return logo, title, nil
}

func RequestSSLabs(name string) ([]uint8, error) {

	urlDomain := fmt.Sprintf("https://api.ssllabs.com/api/v3/analyze?host=%s", name)
	var clean []uint8
	resp, err := http.Get(urlDomain)
	if err != nil {return clean, err	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil { return body, err}
	return body, nil
}
