package main

import (
	"fmt"
	"log"

	"controllers"

	_ "github.com/lib/pq"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func main() {
	r := router.New()
	r.GET("/domain", controllers.GetDomainsQueries)
	r.GET("/domain/{name}", controllers.GetDomainParameters)
	fmt.Println("Listening ...")
	log.Fatal(fasthttp.ListenAndServe(":1206", r.Handler))
}
