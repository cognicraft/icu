package icu

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type Specification string

func (s Specification) Ranges() Ranges {
	rs := Ranges{}
	if s == "" {
		return rs
	}
	ps := strings.Split(string(s), ",")
	for _, p := range ps {
		lrps := strings.Split(p, ";q=")
		c := Tag(strings.TrimSpace(lrps[0]))
		if c == "" {
			continue
		}
		switch len(lrps) {
		case 1:
			rs = append(rs, Range{
				Tag:     c,
				Quality: 1,
			})
		case 2:
			qStr := strings.TrimSpace(lrps[1])
			q, err := strconv.ParseFloat(qStr, 64)
			if err != nil {
				continue
			}
			rs = append(rs, Range{
				Tag:     c,
				Quality: q,
			})
		}
	}
	sort.Stable(rs)
	return rs
}

func (s Specification) Tag() Tag {
	rs := s.Ranges()
	if len(rs) > 0 {
		return rs[0].Tag
	}
	return ""
}

type Range struct {
	Tag     Tag
	Quality float64
}

type Ranges []Range

func (a Ranges) Len() int           { return len(a) }
func (a Ranges) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Ranges) Less(i, j int) bool { return a[i].Quality > a[j].Quality }

func ExtractSpecification(r *http.Request) Specification {
	al := r.Header.Get(HeaderAcceptLanguage)
	return Specification(al)
}

const (
	HeaderAcceptLanguage = "Accept-Language"
)
