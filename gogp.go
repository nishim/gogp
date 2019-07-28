package gogp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type request struct {
	URL string `json:"url"`
}

// OGP metadata.
type OGP struct {
	Title       string `json:"title"`
	Type        string `json:"type"`
	Image       string `json:"image"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

// Gogp returns ogp json string.
func Gogp(w http.ResponseWriter, r *http.Request) {
	reqbody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("ioutil.ReadAll: %v", err)
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}

	req := request{}
	if err = json.Unmarshal(reqbody, &req); err != nil {
		log.Printf("json.Unmarshal: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	res, err := http.Get(req.URL)
	if err != nil {
		log.Println(err)
		http.Error(w, "Cannot access "+req.URL, http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "Cannot load response body", http.StatusInternalServerError)
		return
	}

	htmltext := string(bytes)
	doc, err := html.Parse(strings.NewReader(htmltext))
	if err != nil {
		log.Println(err)
		http.Error(w, "Broken html", http.StatusInternalServerError)
		return
	}
	ogp := &OGP{}
	traverse(doc, ogp)

	ogpb, err := json.Marshal(ogp)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Fprint(w, string(ogpb))
}

func traverse(node *html.Node, ogp *OGP) {
	if node.DataAtom == atom.Body {
		return
	}

	if node.DataAtom == atom.Meta {
		mk, mv := "", ""
		for _, attr := range node.Attr {
			if attr.Key == "property" && strings.HasPrefix(attr.Val, "og:") {
				mk = attr.Val
				continue
			}

			if attr.Key == "content" {
				mv = attr.Val
				continue
			}
		}

		if mk != "" {
			switch mk {
			case "og:title":
				ogp.Title = mv
			case "og:type":
				ogp.Type = mv
			case "og:url":
				ogp.URL = mv
			case "og:image":
				ogp.Image = mv
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		traverse(child, ogp)
	}
}