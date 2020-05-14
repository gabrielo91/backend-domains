package controllers

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"model"

	"github.com/PuerkitoBio/goquery"
)

func RequestSSLabs(name string) ([]uint8, error) {
	urlDomain := fmt.Sprintf("https://api.ssllabs.com/api/v3/analyze?host=%s", name)
	resp, err := http.Get(urlDomain)
	if err != nil {
		fmt.Println("Error")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}
	return body, nil
}

func GetHTMLParameters(url string) (string, string, error) {
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", "", err
	}

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

func hashedVariable(value []byte) string {
	h := sha1.New()
	h.Write(value)
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

type Endpoints struct {
	IpAddress         string
	ServerName        string
	StatusMessage     string
	Grade             string
	GradeTrustIgnored string
	HasWarnings       bool
	IsExceptional     bool
	Progress          int
	Duration          int
	Delegation        int
}

type DomainInfo struct {
	Host            string
	Port            int
	Protocol        string
	IsPublic        bool
	Status          string
	StartTime       int
	TestTime        int
	EngineVersion   string
	CriteriaVersion string
	Endpoints       []Endpoints
}

type Servers struct {
	Address   string
	Ssl_grade string
	Country   string
	Owner     string
}

type DomainInfoSend struct {
	Servers_changed    bool
	Ssl_grade          string
	Previous_ssl_grade string
	Logo               string
	Title              string
	Is_down            bool 
	Servers            []Servers
}

func whoisParameters(ip string) (string, string) {
	app := "whois"
	cmd := fmt.Sprintf("%s %s |grep country -i -m 1 |cut -d ':' -f 2 |xargs", app, ip)
	country, err := exec.Command("bash", "-c", cmd).Output()
	cmd2 := fmt.Sprintf("%s %s |grep organization -i -m 1 |cut -d ':' -f 2 |xargs", app, ip)
	organization, err := exec.Command("bash", "-c", cmd2).Output()

	if err != nil {
		return "", ""
	}

	return strings.TrimSuffix(string(country), "\n"), strings.TrimSuffix(string(organization), "\n")
}

func OrganizeServers(body string, logo string, title string, serverChanged bool, previousGrade string) DomainInfoSend {

	domainJson := body
	var domainInfo DomainInfo
	var domainInfoSend DomainInfoSend
	json.Unmarshal([]byte(domainJson), &domainInfo)

	domainInfoSend.Logo = logo
	domainInfoSend.Title = title
	domainInfoSend.Servers_changed = serverChanged

	grades := map[string]int{
		"A":  7,
		"A+": 6,
		"B":  5,
		"c":  4,
		"D":  3,
		"E":  2,
		"F":  1}

	var smallerGrade int = 8
	for _, value := range domainInfo.Endpoints {
		var server Servers
		server.Address = value.IpAddress
		server.Ssl_grade = value.Grade
		server.Country, server.Owner = whoisParameters(value.IpAddress)
		if grades[value.Grade] < smallerGrade {
			smallerGrade = grades[value.Grade]
			domainInfoSend.Ssl_grade = value.Grade
		}
		domainInfoSend.Servers = append(domainInfoSend.Servers, server)
	}
	if previousGrade == "" {
		domainInfoSend.Previous_ssl_grade = domainInfoSend.Ssl_grade
	} else {
		domainInfoSend.Previous_ssl_grade = previousGrade
	}

	if len(domainInfoSend.Servers) == 0 {
		domainInfoSend.Is_down = false
	} else {
		domainInfoSend.Is_down = true
	}
	return domainInfoSend
}

type AnswerRequest struct {
	Body          []uint8
	Name          string
	Logo          string
	Title         string
	RequestHash   string 
	PreviousGrade string
	UpdatedDate   int64
}

func ValidateConditions(db *sql.DB, data AnswerRequest) DomainInfoSend {
	var result DomainInfoSend
	if data.RequestHash == "" {
		bodyHashed := hashedVariable(data.Body)
		result = OrganizeServers(string(data.Body), data.Logo, data.Title, false, "")
		model.CreateRowDatabase(db, data.Name, "", bodyHashed, result.Ssl_grade)
		fmt.Println("New information saved")
	} else {
		const hour int64 = 3.6e+6
		timeInterval := (time.Now().UnixNano() / int64(time.Millisecond)) - data.UpdatedDate
		if timeInterval > hour {
			bodyHashed := hashedVariable(data.Body)
			var timeNow int64 = time.Now().UnixNano() / int64(time.Millisecond)
			if bodyHashed != data.RequestHash {
				result = OrganizeServers(string(data.Body), data.Logo, data.Title, true, "")
				model.UpdateRowDatabase(db, data.Name, bodyHashed, timeNow, result.Ssl_grade)
				fmt.Println("Change server")
			} else {
				result = OrganizeServers(string(data.Body), data.Logo, data.Title, false, data.PreviousGrade)
				fmt.Println("Without change")

			}
		} else {
			result = OrganizeServers(string(data.Body), data.Logo, data.Title, false, data.PreviousGrade)
			fmt.Println("Without hour change")
		}
	}
	return result
}