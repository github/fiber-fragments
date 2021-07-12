package fragments

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Rel ...
type Rel string

const (
	StyleSheet = "stylesheet"
	Script     = "script"
)

// Link ...
type Link struct {
	URL    string
	Rel    string
	Params map[string]string
}

// Header ...
type Header string

// Link ...
func (s Header) Links() []Link {
	links := make([]Link, 0)

	for _, chunk := range strings.Split(string(s), ",") {

		l := Link{URL: "", Rel: "", Params: make(map[string]string)}

		for _, part := range strings.Split(chunk, ";") {
			part = strings.Trim(part, " ")
			if part == "" {
				continue
			}
			if part[0] == '<' && part[len(part)-1] == '>' {
				l.URL = strings.Trim(part, "<>")
				continue
			}

			key, val := parseParam(part)
			if key == "" {
				continue
			}

			if strings.ToLower(key) == "rel" {
				l.Rel = val

				continue
			}
			l.Params[key] = val
		}

		if l.URL != "" {
			links = append(links, l)
		}
	}

	return links
}

// FilterByStylesheet ...
func FilterByStylesheet(links ...Link) []Link {
	return FilterByRel(links, "stylesheet")
}

// FilterByStylesheet ...
func FilterByScript(links ...Link) []Link {
	return FilterByRel(links, "script")
}

// FilterByRel ...
func FilterByRel(links []Link, rel string) []Link {
	ll := make([]Link, 0)

	for _, l := range links {
		if l.Rel != rel {
			continue
		}

		ll = append(ll, l)
	}

	return ll
}

// CreateNodes ...
func CreateNodes(links []Link) []*html.Node {
	nodes := make([]*html.Node, 0)

	for _, s := range links {
		attr := make([]html.Attribute, 0)

		if s.Rel == Script {
			attr = append(attr, html.Attribute{Key: "src", Val: s.URL})
		}

		if s.Rel == StyleSheet {
			attr = append(attr, html.Attribute{Key: "href", Val: s.URL})
			attr = append(attr, html.Attribute{Key: "rel", Val: s.Rel})
		}

		for k, p := range s.Params {
			attr = append(attr, html.Attribute{Key: k, Val: p})
		}

		node := &html.Node{
			Type: html.ElementNode,
			Attr: attr,
		}

		if s.Rel == "script" {
			node.Data = "script"
			node.DataAtom = atom.Script
		}

		if s.Rel == "stylesheet" {
			node.Data = "link"
			node.DataAtom = atom.Link
		}

		nodes = append(nodes, node)
	}

	return nodes
}

func parseParam(raw string) (key, val string) {
	parts := strings.SplitN(raw, "=", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	if len(parts) != 2 {
		return "", ""
	}

	key = parts[0]
	val = strings.Trim(parts[1], "\"")

	return key, val
}
