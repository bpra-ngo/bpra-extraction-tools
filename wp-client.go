package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type WPClient struct {
	username string
	password string
	baseUrl  string
}

type Post struct {
	Date       string  `json:"date"`
	Title      WSField `json:"title"`
	Content    WSField `json:"content"`
	Author     int     `json:"author"`
	Status     string  `json:"status"` // one of publish, future, draft, pending, private
	Categories []int   `json:"categories"`
	Tags       []int   `json:"tags"`
}

type CreatePostRequest struct {
	Date       string `json:"date"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Status     string `json:"status"`
	Author     int    `json:"author"`
	Categories []int  `json:"categories"`
	Tags       []int  `json:"tags"`
}

type Tag struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type CreateTagRequest struct {
	Name string `json:"name"`
}

type WSField struct {
	Rendered    string `json:"rendered"`
	IsProtected bool   `json:"protected"`
}

type CategoryEnum string
type TagEnum string

var issueCategory = map[CategoryEnum]int{
	"вестник": 6,
	"брой 1":  2,
	"брой 2":  5,
	"брой 3":  11,
	"брой 4":  12,
	"брой 5":  13,
	"брой 6":  14,
	"брой 7":  15,
	"брой 8":  16,
	"брой 9":  17,
	"брой 10": 1,
	"брой 11": 19,
	"брой 12": 20,
	"брой 13": 21,
	"брой 14": 22,
	"брой 15": 23,
	"брой 16": 24,
	"брой 17": 25,
	"брой 18": 26,
	"брой 19": 27,
	"брой 20": 28,
	"брой 21": 29,
	"брой 22": 8,
	"брой 23": 10,
}

var issueTag = map[TagEnum]int{
	"статия": 9,
	"pdf":    7,
}

func (wp *WPClient) GetAllPosts() []Post {
	var out []Post
	req, _ := http.NewRequest("GET", wp.baseUrl+"/posts", http.NoBody)
	body := wp.doRequest(req)
	if err := json.Unmarshal(body, &out); err != nil {
		log.Debug(body)
		panic(err)
	}
	return out
}

func (wp *WPClient) CreatePost(issue Issue, header string, article Article) Post {
	var out Post
	payload := CreatePostRequest{
		Date:       issue.IssueDate.Format("2006-01-02T15:04:05"),
		Title:      header,
		Content:    article.Text,
		Author:     1,
		Status:     "draft",
		Categories: []int{issueCategory[CategoryEnum(fmt.Sprintf("брой %d", issue.IssueNum))]},
		Tags:       []int{issueTag[TagEnum("статия")]},
	}

	existingTags := wp.GetTags()
	for _, t := range existingTags {
		issueTag[TagEnum(t.Name)] = t.Id
	}

	for _, quote := range article.LegalQuotes {
		tag, exists := issueTag[TagEnum(quote)]
		if exists {
			payload.Tags = append(payload.Tags, tag)
		} else {
			newTag := wp.CreateTag(quote)
			payload.Tags = append(payload.Tags, newTag.Id)
			issueTag[TagEnum(newTag.Name)] = newTag.Id
		}
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(payload)
	req, _ := http.NewRequest("POST", wp.baseUrl+"/posts", &buf)
	body := wp.doRequest(req)
	if err := json.Unmarshal(body, &out); err != nil {
		panic(err)
	}
	return out
}

func (wp *WPClient) GetTags() []Tag {
	var out []Tag
	req, _ := http.NewRequest("GET", wp.baseUrl+"/tags", http.NoBody)
	body := wp.doRequest(req)
	if err := json.Unmarshal(body, &out); err != nil {
		log.Debug(body)
		panic(err)
	}
	return out
}

func (wp *WPClient) CreateTag(name string) Tag {
	var out Tag
	payload := CreateTagRequest{Name: name}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(payload)
	req, _ := http.NewRequest("POST", wp.baseUrl+"/tags", &buf)
	body := wp.doRequest(req)
	if err := json.Unmarshal(body, &out); err != nil {
		log.Debug(body)
		panic(err)
	}
	return out
}

func (wp *WPClient) doRequest(req *http.Request) []byte {
	client := &http.Client{}
	req.SetBasicAuth(wp.username, wp.password)
	req.Header.Add("Content-Type", "application/json")

	r, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	if r.StatusCode != 201 && r.StatusCode != 200 {
		log.Debug(string(body))
		panic(r)
	}
	return body
}
