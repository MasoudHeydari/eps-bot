package model

import (
	"time"
)

type SearchQueryRequest struct {
	Query    string `json:"q"`
	Location int    `json:"loc"`
	Language string `json:"lang"`
}

type SERPResponse struct {
	URL         string
	Title       string
	Description string
	Phones      []string
	Emails      []string
	Keywords    []string
	IsRead      bool
	CreatedAt   time.Time
}

type ContactInfo struct {
	Emails []string `json:"emails"`
	Phones []string `json:"phones"`
}

type SearchQueryResponse struct {
	SQs []SearchQuery `json:"search_queries"`
}

type SearchQuery struct {
	Id         int       `json:"id"`
	Query      string    `json:"query"`
	Language   string    `json:"language"`
	Location   int       `json:"location"`
	IsCanceled bool      `json:"is_canceled"`
	CreatedAt  time.Time `json:"created_at"`
}
