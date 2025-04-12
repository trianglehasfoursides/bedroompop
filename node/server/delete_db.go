package server

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/tidwall/gjson"
	"github.com/trianglehasfoursides/mathrock/node/sqlite"
	"go.uber.org/zap"
)

func deleteDb(w http.ResponseWriter, r *http.Request) {
	var msg string
	mtx := &sync.Mutex{}
	req, err := io.ReadAll(r.Body)
	if err != nil {
		msg = fmt.Sprintf(`{"error": "%s"}`, err.Error())
		w.Write([]byte(msg))
		return
	}
	name := gjson.Get(string(req), "name").String()
	if name == "" {
		w.Write([]byte(`{"error": "database name can't be nil"}`))
		return
	}
	if err = sqlite.DeleteDb(name, mtx); err != nil {
		w.Write([]byte(err.Error()))
		zap.L().Error(err.Error())
		return
	}
}
