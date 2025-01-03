package store

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Ex 38
type PaginatedFeedQuery struct {
	//For limit validate section lte can be 2000 or anylimit, Its upto the application level data
	//As our app is small, we are using limit as 20
	Limit  int      `json:"limit" validate:"gte=1,lte=20"`
	Offset int      `json:"offset" validate:"gte=0"`
	Sort   string   `json:"sort" validate:"oneof=asc desc"`
	Tags   []string `json:"tags" validate:"max=5"`
	//ex 39 search keyword is used to filter both Title and content easier for both at a time in frontend
	Search string `json:"search" validate:"max=100"`
	Since  string `json:"since"`
	Until  string `json:"until"`
}

// Ex 38 This will parse URL before validating by Validator function, extracts limit, offset and sort from request URL
func (fq PaginatedFeedQuery) Parse(r *http.Request) (PaginatedFeedQuery, error) {

	qs := r.URL.Query()
	limit := qs.Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return fq, err
		}
		fq.Limit = l
	}
	offset := qs.Get("offset")
	if offset != "" {
		l, err := strconv.Atoi(offset)
		if err != nil {
			return fq, err
		}
		fq.Offset = l
	}

	sort := qs.Get("sort")
	if sort != "" {
		fq.Sort = sort
	}

	tags := qs.Get("tags")
	if tags != "" {
		fq.Tags = strings.Split(tags, ",")
	}

	search := qs.Get("search")
	if search != "" {
		fq.Search = search
	}

	since := qs.Get("since")
	if since != "" {
		fq.Since = parseTime(since)
	}

	until := qs.Get("until")
	if since != "" {
		fq.Since = parseTime(until)
	}
	return fq, nil
}

func parseTime(s string) string {
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return ""
	}

	return t.Format(time.DateTime)
}
