package controllers

import (
	"crypto/sha1"
	"database/sql"
	"db"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"model/domains"
	"processing"
	"services"
	"time"

	"utils"

	"github.com/valyala/fasthttp"
)

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

func hashedVariable(value []byte) string {
	h := sha1.New()
	h.Write(value)
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
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
		server.Country, server.Owner = processing.WhoisParameters(value.IpAddress)
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
		domains.Insert(db, data.Name, "", bodyHashed, result.Ssl_grade)
		fmt.Println("New information saved")
	} else {
		const hour int64 = 3.6e+6
		timeInterval := (time.Now().UnixNano() / int64(time.Millisecond)) - data.UpdatedDate
		if timeInterval > hour {
			bodyHashed := hashedVariable(data.Body)
			var timeNow int64 = time.Now().UnixNano() / int64(time.Millisecond)
			if bodyHashed != data.RequestHash {
				result = OrganizeServers(string(data.Body), data.Logo, data.Title, true, "")
				domains.Update(db, data.Name, bodyHashed, timeNow, result.Ssl_grade)
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

func GetDomainParameters(ctx *fasthttp.RequestCtx) {

	var err error
	defer func() {
		if err != nil {
				fmt.Printf("error: %v", err)
				utils.DoJSONWrite(ctx, 400, err)
		}
	}()

	var data AnswerRequest
	name := fmt.Sprintf("%s", ctx.UserValue("name"))
	data.Name = name
	data.Logo, data.Title, err = services.GetHTMLParameters(fmt.Sprintf("https://www.%s", ctx.UserValue("name")))
	if err != nil { return }
	data.Body, err = services.RequestSSLabs(name)
	if err != nil { return }
	db, err := db.ConnectDatabase()
	if err != nil { return }
	data.RequestHash, data.PreviousGrade, data.UpdatedDate, err = domains.Find(db, name)
	result := ValidateConditions(db, data)
	utils.DoJSONWrite(ctx, 200, result)

}

func GetDomainsQueries(ctx *fasthttp.RequestCtx) {
	var err error
	defer func() {
		if err != nil {
				fmt.Printf("error: %v", err)
				utils.DoJSONWrite(ctx, 400, err)
		}
	}()

	db, err := db.ConnectDatabase()
	if err != nil {return	} 
	result, err := domains.FindIById(db)
	if err != nil {return	} 
	utils.DoJSONWrite(ctx, 200, result)
}

