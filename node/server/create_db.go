package server

import (
	"io"
	"net/http"

	"github.com/tidwall/gjson"
	"github.com/trianglehasfoursides/mathrock/node/sqlite"
	"go.uber.org/zap"
)

func createDb(w http.ResponseWriter, r *http.Request) {
	req, err := io.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	name := gjson.Get(string(req), "name").String()
	if name == "" {
		w.Write([]byte(""))
		return
	}
	if err = sqlite.CreateDb(name); err != nil {
		w.Write([]byte(err.Error()))
		zap.L().Error(err.Error())
		return
	}
}
