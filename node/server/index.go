package server

import "net/http"

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hallo from m@+r0_)$($)?k"))
}
