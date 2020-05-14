package main

import (
	"fmt"
	"log"

	"controllers"
	"model"
	"utils"

	_ "github.com/lib/pq"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func GetDomainParameters(ctx *fasthttp.RequestCtx) {
	var data controllers.AnswerRequest
	var err error
	name := fmt.Sprintf("%s", ctx.UserValue("name"))
	data.Name = name
	data.Logo, data.Title, err = controllers.GetHTMLParameters(fmt.Sprintf("https://www.%s", ctx.UserValue("name")))
	data.Body, err = controllers.RequestSSLabs(name)
	db, err := model.ConnectDatabase()
	data.RequestHash, data.PreviousGrade, data.UpdatedDate = model.GetRowDatabase(db, name)
	result := controllers.ValidateConditions(db, data)
	if err != nil {
		utils.DoJSONWrite(ctx, 400, "Not found information")
	}
	utils.DoJSONWrite(ctx, 200, result)
}

func GetDomainsQueries(ctx *fasthttp.RequestCtx) {
	db, err := model.ConnectDatabase()
	if err != nil {
		fmt.Println("Falleeeeeeeeeeeeeeeeeeeeeee")
		utils.DoJSONWrite(ctx, 500, "Connection refused")
	} else {
		result := model.GetAllIdDatabase(db)
		utils.DoJSONWrite(ctx, 200, result)
	}
}

func main() {
	r := router.New()
	r.GET("/domain", GetDomainsQueries)
	r.GET("/domain/{name}", GetDomainParameters)
	fmt.Println("Listening ...")
	log.Fatal(fasthttp.ListenAndServe(":1206", r.Handler))
}
