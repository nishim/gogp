package main

import (
	"net/http"

	"github.com/nishim/gogp"
)

func main() {
	http.HandleFunc("/", gogp.Gogp) // ハンドラを登録してウェブページを表示させる
	http.ListenAndServe(":8080", nil)
}
