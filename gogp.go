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
	SiteName    string `json:"site_name"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Image       string `json:"image"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Favicon     string `json:"favicon"`
}

// Gogp returns ogp json string.
func Gogp(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	v, ok := q["url"]
	if !ok {
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}

	res, err := http.Get(v[0])
	if err != nil {
		log.Println(err)
		http.Error(w, "Cannot access "+v[0], http.StatusInternalServerError)
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

	if node.DataAtom == atom.Link {
		mk, mv := "", ""
		for _, attr := range node.Attr {
			if attr.Key == "rel" && attr.Val == "icon" {
				mk = attr.Val
				continue
			}

			if attr.Key == "href" {
				mv = attr.Val
				continue
			}
		}

		if mk != "" {
			switch mk {
			case "icon":
				ogp.Favicon = mv
			}
		}
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
			case "og:site_name":
				ogp.SiteName = mv
			case "og:title":
				ogp.Title = mv
			case "og:type":
				ogp.Type = mv
			case "og:url":
				ogp.URL = mv
			case "og:image":
				ogp.Image = mv
			case "og:description":
				ogp.Description = mv
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		traverse(child, ogp)
	}
}
