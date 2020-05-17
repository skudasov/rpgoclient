package rpgoclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang-collections/collections/stack"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"time"
)

type Client struct {
	BaseURL    *url.URL
	ApiURL     string
	Project    string
	BTSProject string
	BTSUrl     string
	Token      string
	UserAgent  string

	Stack    *stack.Stack
	LaunchId string
	Retries  int

	httpClient *http.Client
	l          *zap.SugaredLogger
}

func New(baseUrl string, project string, token string, btsProject string, btsUrl string, dumptransport bool, options ...func(*Client) error) *Client {
	c := &Client{}
	c.httpClient = NewLoggingHTTPClient(dumptransport, 120)

	u, err := url.Parse(baseUrl)
	if err != nil {
		log.Fatal(err)
	}
	c.BaseURL = u
	c.Project = project
	c.Token = token
	c.ApiURL = "/api/v1"
	c.Retries = 3
	c.l = NewLogger("info")
	c.Stack = stack.New()
	c.LaunchId = ""
	c.BTSProject = btsProject
	c.BTSUrl = btsUrl

	for _, op := range options {
		err := op(c)
		if err != nil {
			c.l.Fatalf("option failed: %s", err)
		}
	}
	return c
}

func (c *Client) GetBaseUrl() string {
	return c.BaseURL.String()
}

func (c *Client) GetLaunchId() string {
	return c.LaunchId
}

func (c *Client) GetProject() string {
	return c.Project
}

func (c *Client) GetToken() string {
	return c.Token
}

func WithBaseUrl(baseUrl string) func(client *Client) error {
	return func(c *Client) error {
		u, err := url.Parse(baseUrl)
		if err != nil {
			return err
		}
		c.BaseURL = u
		return nil
	}
}

func WithHttpClient(httpClient *http.Client) func(client *Client) error {
	return func(c *Client) error {
		c.httpClient = httpClient
		return nil
	}
}

func WithVerbosity(verbosity string) func(client *Client) error {
	return func(c *Client) error {
		c.l = NewLogger(verbosity)
		return nil
	}
}

func WithRetries(retries int) func(client *Client) error {
	return func(c *Client) error {
		c.Retries = retries
		return nil
	}
}

func (c *Client) StartLaunch(name string, description string, startTimeStringRFC3339 string, tags []string, mode string) (StartLaunchResponse, error) {
	var startTime string
	if startTimeStringRFC3339 != "" {
		startTime = startTimeStringRFC3339
	} else {
		startTime = time.Now().Format(time.RFC3339)
	}
	p := StartLaunchPayload{
		Name:        name,
		StartTime:   startTime,
		Description: description,
		Tags:        tags,
		Mode:        mode,
	}
	req, err := c.newRequest("POST", fmt.Sprintf("%s/%s/launch", c.ApiURL, c.Project), p, "application/json")
	if err != nil {
		return StartLaunchResponse{}, err
	}
	var respBody StartLaunchResponse
	_, err = c.do(req, &respBody)
	if err != nil {
		return StartLaunchResponse{}, err
	}
	if respBody.Id != "" {
		c.LaunchId = respBody.Id
	}
	c.Stack.Push(nil)
	c.LaunchId = respBody.Id
	c.l.Debugf("created new test launch: %s", respBody.Id)
	return respBody, err
}

func (c *Client) FinishLaunch(status string, endTimeStringRFC3339 string) (FinishLaunchResponse, error) {
	var endTime string
	if endTimeStringRFC3339 != "" {
		endTime = endTimeStringRFC3339
	} else {
		endTime = time.Now().Format(time.RFC3339)
	}
	p := FinishLaunchPayload{
		Status:  status,
		EndTime: endTime,
	}
	if c.LaunchId == "" {
		return FinishLaunchResponse{}, noLaunchIdErr
	}
	req, err := c.newRequest("PUT", fmt.Sprintf("%s/%s/launch/%s/finish", c.ApiURL, c.Project, c.LaunchId), p, "application/json")
	if err != nil {
		return FinishLaunchResponse{}, err
	}
	var respBody FinishLaunchResponse
	_, err = c.do(req, &respBody)
	if err != nil {
		return respBody, err
	}
	if c.Stack.Len() >= 1 {
		c.Stack.Pop()
	}
	c.l.Debugf("launch finished: %s", c.LaunchId)
	return respBody, err
}

func (c *Client) StartTestItem(name string, itemType string, startTimeStringRFC3339 string, description string, tags []string, parameters []map[string]string) (StartTestItemResponse, error) {
	var startTime string
	if startTimeStringRFC3339 != "" {
		startTime = startTimeStringRFC3339
	} else {
		startTime = time.Now().Format(time.RFC3339)
	}
	p := StartTestItemPayload{
		Name:        name,
		StartTime:   startTime,
		Description: description,
		Tags:        tags,
		LaunchId:    c.LaunchId,
		Type:        itemType,
		Parameters:  parameters,
	}
	c.l.Debugf("starting test item of type: %s", p.Type)
	var u string
	parentItemId := c.Stack.Peek()
	if parentItemId != nil {
		u = fmt.Sprintf("%s/%s/item/%s", c.ApiURL, c.Project, parentItemId)
	} else {
		u = fmt.Sprintf("%s/%s/item", c.ApiURL, c.Project)
	}
	req, err := c.newRequest("POST", u, p, "application/json")
	if err != nil {
		return StartTestItemResponse{}, err
	}
	var respBody StartTestItemResponse
	_, err = c.do(req, &respBody)
	if err != nil {
		return StartTestItemResponse{}, err
	}
	c.Stack.Push(respBody.Id)
	c.l.Debugf("started test item: %s", respBody.Id)
	return respBody, err
}

func (c *Client) StartTestItemId(parentItemId string, name string, itemType string, startTimeStringRFC3339 string, description string, tags []string, parameters []map[string]string) (StartTestItemResponse, error) {
	var startTime string
	if startTimeStringRFC3339 != "" {
		startTime = startTimeStringRFC3339
	} else {
		startTime = time.Now().Format(time.RFC3339)
	}
	p := StartTestItemPayload{
		Name:        name,
		StartTime:   startTime,
		Description: description,
		Tags:        tags,
		LaunchId:    c.LaunchId,
		Type:        itemType,
		Parameters:  parameters,
	}
	c.l.Debugf("starting test item of type: %s", p.Type)
	var u string
	if parentItemId != "" {
		u = fmt.Sprintf("%s/%s/item/%s", c.ApiURL, c.Project, parentItemId)
	} else {
		u = fmt.Sprintf("%s/%s/item", c.ApiURL, c.Project)
	}
	req, err := c.newRequest("POST", u, p, "application/json")
	if err != nil {
		return StartTestItemResponse{}, err
	}
	var respBody StartTestItemResponse
	_, err = c.do(req, &respBody)
	if err != nil {
		return StartTestItemResponse{}, err
	}
	c.Stack.Push(respBody.Id)
	c.l.Debugf("started test item: %s", respBody.Id)
	return respBody, err
}

func (c *Client) FinishTestItem(status string, endTimeStringRFC3339 string, issue map[string]interface{}) (string, error) {
	if issue == nil && status == "SKIPPED" {
		issue = make(map[string]interface{})
		issue["issue_type"] = "NOT_ISSUE"
	}
	var endTime string
	if endTimeStringRFC3339 != "" {
		endTime = endTimeStringRFC3339
	} else {
		endTime = time.Now().Format(time.RFC3339)
	}
	p := FinishTestItemPayload{
		EndTime: endTime,
		Status:  status,
		Issue:   issue,
	}
	itemId := c.Stack.Pop()
	c.l.Debugf("finishing test item", "itemid", itemId, "status", status, "issue", issue)
	req, err := c.newRequest("PUT", fmt.Sprintf("%s/%s/item/%s", c.ApiURL, c.Project, itemId), p, "application/json")
	if err != nil {
		return "", err
	}
	var respBody FinishTestItemResponse
	_, err = c.do(req, &respBody)
	if err != nil {
		return "", err
	}
	c.l.Debugf("finished test item: %s", respBody.Msg)
	return respBody.Msg, err
}

func (c *Client) FinishTestItemId(id string, status string, endTimeStringRFC3339 string, issue map[string]interface{}) (string, error) {
	if issue == nil && status == "SKIPPED" {
		issue = make(map[string]interface{})
		issue["issue_type"] = "NOT_ISSUE"
	}
	var endTime string
	if endTimeStringRFC3339 != "" {
		endTime = endTimeStringRFC3339
	} else {
		endTime = time.Now().Format(time.RFC3339)
	}
	p := FinishTestItemPayload{
		EndTime: endTime,
		Status:  status,
		Issue:   issue,
	}
	c.l.Debugf("finishing test item with id: %d, status: %s, issue: %s", id, status, issue)
	req, err := c.newRequest("PUT", fmt.Sprintf("%s/%s/item/%s", c.ApiURL, c.Project, id), p, "application/json")
	if err != nil {
		return "", err
	}
	var respBody FinishTestItemResponse
	_, err = c.do(req, &respBody)
	if err != nil {
		return "", err
	}
	c.l.Debugf("finished test item: %s", respBody.Msg)
	return respBody.Msg, err
}

func (c *Client) LinkIssue(itemId int, ticketId string, link string) (string, error) {
	p := LinkIssue{
		Issues: []Issue{
			{
				BtsProject: c.BTSProject,
				BtsUrl:     c.BTSUrl,
				// the way we handle tickets breaks workflow of Report Portal,
				// bugs should be created after run, not before
				// we link tickets that already have been parsed out of logs
				// so submit date cannot be obtained without jira client here
				SubmitDate: time.Now().Unix(),
				TicketId:   ticketId,
				Url:        link,
			},
		},
		TestItemIds: []int{
			itemId,
		},
	}
	c.l.Debugf("linking item issues with id: %d, ticketId: %s, link: %s", itemId, ticketId, link)
	req, err := c.newRequest("PUT", fmt.Sprintf("%s/%s/item/issue/link", c.ApiURL, c.Project), p, "application/json")
	if err != nil {
		return "", err
	}
	var respBody FinishTestItemResponse
	_, err = c.do(req, &respBody)
	if err != nil {
		return "", err
	}
	c.l.Debugf("linked item issues: %s", respBody.Msg)
	return respBody.Msg, err
}

func (c *Client) LogBatch(messages []LogPayload) error {
	body := bytes.NewBufferString("")
	bodyWriter := multipart.NewWriter(body)
	mh := make(textproto.MIMEHeader)
	mh.Set("Content-Type", "application/json")
	mh.Set("Content-Disposition", `form-data; name="json_request_part"`)
	pWriter, err := bodyWriter.CreatePart(mh)
	if nil != err {
		return err
	}
	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(messages)
	if err != nil {
		return err
	}
	if _, err := io.Copy(pWriter, buf); err != nil {
		return err
	}
	if err := bodyWriter.Close(); err != nil {
		return err
	}
	req, err := c.newRequest("POST", fmt.Sprintf("%s/%s/log", c.ApiURL, c.Project), body, bodyWriter.FormDataContentType())
	if err != nil {
		return err
	}
	var respBody LogResponse
	_, err = c.do(req, &respBody)
	if err != nil {
		return err
	}
	c.l.Debugf("log batch attached id: %s", respBody.Id)
	return nil
}

func (c *Client) Log(message string, level string) (string, error) {
	var lastItemId string
	if c.Stack.Peek() != nil {
		lastItemId = c.Stack.Peek().(string)
	} else {
		return "", logNotAttachableToLaunchErr
	}
	p := LogPayload{
		ItemId:  lastItemId,
		Time:    time.Now().Format(time.RFC3339),
		Message: message,
		Level:   level,
	}
	c.l.Debugf("attaching log to test item: %s, msg: %s, lvl: %s", p.ItemId, p.Message, p.Level)

	req, err := c.newRequest("POST", fmt.Sprintf("%s/%s/log", c.ApiURL, c.Project), p, "application/json")
	if err != nil {
		return "", err
	}
	var respBody LogResponse
	_, err = c.do(req, &respBody)
	if err != nil {
		return "", err
	}
	c.l.Debugf("log attached id: %s", respBody.Id)
	return respBody.Id, err
}

func (c *Client) LogId(id string, message string, level string) (string, error) {
	p := LogPayload{
		ItemId:  id,
		Time:    time.Now().Format(time.RFC3339),
		Message: message,
		Level:   level,
	}
	c.l.Debugf("attaching log to test item: %s, msg: %s, lvl: %s", p.ItemId, p.Message, p.Level)

	req, err := c.newRequest("POST", fmt.Sprintf("%s/%s/log", c.ApiURL, c.Project), p, "application/json")
	if err != nil {
		return "", err
	}
	var respBody LogResponse
	_, err = c.do(req, &respBody)
	if err != nil {
		return "", err
	}
	c.l.Debugf("log attached id: %s", respBody.Id)
	return respBody.Id, err
}

func (c *Client) GetItemIdByUUID(uuid string) (GetItemResponse, error) {
	req, err := c.newRequest("GET", fmt.Sprintf("%s/%s/item/%s", c.ApiURL, c.Project, uuid), nil, "")
	if err != nil {
		return GetItemResponse{}, err
	}
	var respBody GetItemResponse
	_, err = c.do(req, &respBody)
	if err != nil {
		return GetItemResponse{}, err
	}
	c.l.Debugf("get item id by uuid: %s\n", respBody)
	return respBody, err
}

func (c *Client) newRequest(method, path string, body interface{}, contentType string) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := c.BaseURL.ResolveReference(rel)
	var buf io.ReadWriter
	if body != nil && contentType == "application/json" {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	} else if body != nil {
		buf = body.(io.ReadWriter)
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Add("Authorization", fmt.Sprintf("bearer %s", c.Token))
	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	var err error
	var resp *http.Response
	for i := 0; i <= c.Retries; i++ {
		resp, err = c.httpClient.Do(req)
		if err != nil {
			c.l.Error(err)
		}
		if resp != nil && resp.StatusCode >= 400 {
			c.l.Errorf("request failed: status: %s", resp.Status)
			if resp.Body != nil {
				bb, _ := ioutil.ReadAll(resp.Body)
				defer resp.Body.Close()
				c.l.Errorf("body: %s, err: %s", resp.Status, string(bb), err)
			}
			continue
		}
		err = json.NewDecoder(resp.Body).Decode(v)
		err = resp.Body.Close()
		return resp, err
	}
	return nil, httpRetriesReachedErr
}
