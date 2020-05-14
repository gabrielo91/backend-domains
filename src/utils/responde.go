package utils

import (
	"encoding/json"

	_ "github.com/lib/pq"
	"github.com/valyala/fasthttp"
)

var (
	strContentType     = []byte("Content-Type")
	strApplicationJSON = []byte("application/json")
)

func DoJSONWrite(ctx *fasthttp.RequestCtx, code int, obj interface{}) {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.SetCanonical(strContentType, strApplicationJSON)
	ctx.Response.SetStatusCode(code)
	if err := json.NewEncoder(ctx).Encode(obj); err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}
}
