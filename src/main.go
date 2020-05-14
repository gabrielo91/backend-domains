package main


import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"model"
	"utils"
	"controllers"

	"crypto/sha1"

	_ "github.com/lib/pq"

	"github.com/fasthttp/router"
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
		server.Country, server.Owner = controllers.WhoisParameters(value.IpAddress)
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

	return domainInfoSend
}

func hashedVariable(value []byte) string {
	h := sha1.New()
	h.Write(value)
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

func GetDomainParameters(ctx *fasthttp.RequestCtx) {

	name := fmt.Sprintf("%s", ctx.UserValue("name"))
	urlDomain := fmt.Sprintf("https://api.ssllabs.com/api/v3/analyze?host=%s", ctx.UserValue("name"))
	resp, err := http.Get(urlDomain)
	if err != nil {
		fmt.Println((err))
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println((err))
	}

	logo, title, err := controllers.GetHTMLParameters(fmt.Sprintf("https://www.%s", ctx.UserValue("name")))
	if err != nil {
		fmt.Println(err)
	}

	db, err := model.ConnectDatabase()
	requestHash, previousGrade, updated_at := model.GetRowDatabase(db, name)
	fmt.Println(requestHash, previousGrade, updated_at)

	var result DomainInfoSend
	if requestHash == "" {
		bodyHashed := hashedVariable(body)
		result = OrganizeServers(string(body), logo, title, false, "")
		model.CreateRowDatabase(db, name, "", bodyHashed, result.Ssl_grade)
		fmt.Println("New information saved")

	} else {
		const hour int64 = 3.6e+6
		timeInterval := (time.Now().UnixNano() / int64(time.Millisecond)) - updated_at
		fmt.Println(timeInterval)
		if timeInterval > hour {
			bodyHashed := hashedVariable(body)
			var timeNow int64 = time.Now().UnixNano() / int64(time.Millisecond)
			if bodyHashed != requestHash {
				result = OrganizeServers(string(body), logo, title, true, "")
				model.UpdateRowDatabase(db, name, bodyHashed, timeNow, result.Ssl_grade)
				fmt.Println("Change server")
			} else {
				result = OrganizeServers(string(body), logo, title, false, previousGrade)
				fmt.Println("Without change")

			}
		} else {
			result = OrganizeServers(string(body), logo, title, false, previousGrade)
			fmt.Println("Without hour change")

		}

	}
	utils.DoJSONWrite(ctx, 200, result)

}


func GetRequests(ctx *fasthttp.RequestCtx) {
	db, err := model.ConnectDatabase()
	if err != nil {
		fmt.Println(err)
	}
	result := model.GetAllIdDatabase(db)
	fmt.Println(result)
	utils.DoJSONWrite(ctx, 200, result)
}

func main() {
	r := router.New()
	r.GET("/domain", GetRequests)
	r.GET("/domain/{name}", GetDomainParameters)
	fmt.Println("Listening ...")
	log.Fatal(fasthttp.ListenAndServe(":1206", r.Handler))
}
