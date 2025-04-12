package server

import (
	"io"
	"net/http"
	"sync"

	"github.com/tidwall/gjson"
	"github.com/trianglehasfoursides/mathrock/node/sqlite"
	"go.uber.org/zap"
)

func createDb(w http.ResponseWriter, r *http.Request) {
	mtx := &sync.Mutex{}
	req, err := io.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte(""))
		return
	}
	name := gjson.Get(string(req), "name").String()
	if name == "" {
		w.Write([]byte(""))
		return
	}
	if err = sqlite.CreateDb(name, mtx); err != nil {
		w.Write([]byte(err.Error()))
		zap.L().Error(err.Error())
		return
	}
}
