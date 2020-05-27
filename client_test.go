package rpgoclient

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var C *Client

const (
	token      = ""
	project    = ""
	btsProject = ""
	btsUrl     = ""
	ua         = "testuseragent"
)

func TestClient_StartLaunch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.String(), "/api/v1/testproj/launch")

		var startLaunch *StartLaunchPayload
		err := json.NewDecoder(r.Body).Decode(&startLaunch)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, "testrun", startLaunch.Name)
		assert.NotNil(t, startLaunch.StartTime)
		assert.Equal(t, "test launch", startLaunch.Description)
		assert.Equal(t, []string{"tag1"}, startLaunch.Tags)
		assert.Equal(t, "DEFAULT", startLaunch.Mode)

		re := &StartLaunchResponse{Number: 1, Id: "id1"}
		data, _ := json.Marshal(re)
		_, _ = w.Write(data)
	}))
	defer ts.Close()
	C = New(ts.URL, project, token, btsProject, false)
	launchId, err := C.StartLaunch("testrun", "test launch", "", []string{"tag1"}, "DEFAULT")
	if err != nil {
		t.Error(err)
	}
	assert.NotNil(t, launchId)
	assert.Equal(t, 1, C.Stack.Len())
}

func TestClient_StartLaunchWithTime(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.String(), "/api/v1/testproj/launch")

		var startLaunch *StartLaunchPayload
		err := json.NewDecoder(r.Body).Decode(&startLaunch)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, "testrun", startLaunch.Name)
		assert.Equal(t, "2019-02-01T14:21:30.049064+03:00", startLaunch.StartTime)
		assert.Equal(t, "test launch", startLaunch.Description)
		assert.Equal(t, []string{"tag1"}, startLaunch.Tags)
		assert.Equal(t, "DEFAULT", startLaunch.Mode)

		re := &StartLaunchResponse{Number: 1, Id: "id1"}
		data, _ := json.Marshal(re)
		_, _ = w.Write(data)
	}))
	defer ts.Close()
	C = New(ts.URL, project, token, btsProject, false)
	launchId, err := C.StartLaunch("testrun", "test launch", "2019-02-01T14:21:30.049064+03:00", []string{"tag1"}, "DEFAULT")
	if err != nil {
		t.Error(err)
	}
	assert.NotNil(t, launchId)
	assert.Equal(t, 1, C.Stack.Len())
}

func TestClient_FinishLaunch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.String(), "/api/v1/testproj/launch/launch_id/finish")

		var finishLaunch *FinishLaunchPayload
		err := json.NewDecoder(r.Body).Decode(&finishLaunch)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, "PASSED", finishLaunch.Status)
		assert.NotNil(t, finishLaunch.Status)

		re := &FinishLaunchResponse{Id: "1"}
		data, _ := json.Marshal(re)
		_, _ = w.Write(data)
	}))
	defer ts.Close()

	C = New(ts.URL, project, token, btsProject, false)
	C.Stack.Push(nil)
	C.LaunchId = "launch_id"
	msg, err := C.FinishLaunch("PASSED", "")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, fmt.Sprintf("Launch with ID = '%s' successfully finished.", "launch_id"), msg)
	assert.Equal(t, 0, C.Stack.Len())
}

func TestClient_StartTestItem(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.String(), "/api/v1/testproj/item")

		var startTestItem *StartTestItemPayload
		err := json.NewDecoder(r.Body).Decode(&startTestItem)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, "test_item", startTestItem.Name)
		assert.NotNil(t, startTestItem.StartTime)
		assert.Equal(t, "launch_id", startTestItem.LaunchId)
		assert.Equal(t, "test suite description", startTestItem.Description)
		assert.Equal(t, []string{"tag1"}, startTestItem.Tags)
		assert.Equal(t, []map[string]string{{"key": "sdf", "value": "sdF"}}, startTestItem.Parameters)

		re := &StartTestItemResponse{Id: "item_id"}
		data, _ := json.Marshal(re)
		_, _ = w.Write(data)
	}))
	defer ts.Close()
	C = New(ts.URL, project, token, btsProject, false)
	C.LaunchId = "launch_id"
	params := make([]map[string]string, 0)
	params = append(params, map[string]string{"key": "sdf", "value": "sdF"})
	launchId, err := C.StartTestItem("test_item", "SUITE", "", "test suite description", []string{"tag1"}, params)
	if err != nil {
		t.Error(err)
	}
	assert.NotNil(t, launchId)
	assert.Equal(t, 1, C.Stack.Len())
}

func TestClient_StartTestItemWithTimeRFC3339(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.String(), "/api/v1/testproj/item")

		var startTestItem *StartTestItemPayload
		err := json.NewDecoder(r.Body).Decode(&startTestItem)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, "test_item", startTestItem.Name)
		assert.Equal(t, "2019-02-01T14:21:30.049304+03:00", startTestItem.StartTime)
		assert.Equal(t, "launch_id", startTestItem.LaunchId)
		assert.Equal(t, "test suite description", startTestItem.Description)
		assert.Equal(t, []string{"tag1"}, startTestItem.Tags)
		assert.Equal(t, []map[string]string{{"key": "sdf", "value": "sdF"}}, startTestItem.Parameters)

		re := &StartTestItemResponse{Id: "item_id"}
		data, _ := json.Marshal(re)
		_, _ = w.Write(data)
	}))
	defer ts.Close()
	C = New(ts.URL, project, token, btsProject, false)
	C.LaunchId = "launch_id"
	params := make([]map[string]string, 0)
	params = append(params, map[string]string{"key": "sdf", "value": "sdF"})
	launchId, err := C.StartTestItem("test_item", "SUITE", "2019-02-01T14:21:30.049304+03:00", "test suite description", []string{"tag1"}, params)
	if err != nil {
		t.Error(err)
	}
	assert.NotNil(t, launchId)
	assert.Equal(t, 1, C.Stack.Len())
}

func TestClient_FinishTestItem(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.String(), "/api/v1/testproj/item/parent_item_id")

		var finishItem *FinishTestItemPayload
		err := json.NewDecoder(r.Body).Decode(&finishItem)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, "PASSED", finishItem.Status)
		assert.NotNil(t, finishItem.Status)

		re := &FinishTestItemResponse{Msg: fmt.Sprintf("Launch with ID = '%s' successfully finished.", "launch_id")}
		data, _ := json.Marshal(re)
		_, _ = w.Write(data)
	}))
	defer ts.Close()

	C = New(ts.URL, project, token, btsProject, false)
	C.Stack.Push("parent_item_id")
	msg, err := C.FinishTestItem("PASSED", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, fmt.Sprintf("Launch with ID = '%s' successfully finished.", "launch_id"), msg)
	assert.Equal(t, 0, C.Stack.Len())
}

func TestClient_FinishTestItemWithEndTime(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.String(), "/api/v1/testproj/item/parent_item_id")

		var finishItem *FinishTestItemPayload
		err := json.NewDecoder(r.Body).Decode(&finishItem)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, "PASSED", finishItem.Status)
		assert.Equal(t, "2019-02-01T14:21:30.049304+03:00", finishItem.EndTime)
		assert.NotNil(t, finishItem.Status)

		re := &FinishTestItemResponse{Msg: fmt.Sprintf("Launch with ID = '%s' successfully finished.", "launch_id")}
		data, _ := json.Marshal(re)
		_, _ = w.Write(data)
	}))
	defer ts.Close()

	C = New(ts.URL, project, token, btsProject, false)
	C.Stack.Push("parent_item_id")
	msg, err := C.FinishTestItem("PASSED", "2019-02-01T14:21:30.049304+03:00", nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, fmt.Sprintf("Launch with ID = '%s' successfully finished.", "launch_id"), msg)
	assert.Equal(t, 0, C.Stack.Len())
}

func TestClient_Log(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.String(), "/api/v1/testproj/log")

		var logPayload *LogPayload
		err := json.NewDecoder(r.Body).Decode(&logPayload)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, "last_item_id", logPayload.ItemId)
		assert.Equal(t, "DEBUG", logPayload.Level)
		assert.Equal(t, "logmessage", logPayload.Message)
		assert.NotNil(t, logPayload.Time)

		re := &LogResponse{Id: "log_id"}
		data, _ := json.Marshal(re)
		_, _ = w.Write(data)
	}))
	defer ts.Close()

	C = New(ts.URL, project, token, btsProject, false)
	C.Stack.Push(nil)
	C.Stack.Push("last_item_id")
	msg, err := C.Log("logmessage", "DEBUG")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "log_id", msg)
	assert.Equal(t, 2, C.Stack.Len())
}

func TestClient_LogBatch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.String(), "/api/v1/testproj/log")

		_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			log.Fatal(err)
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		p, err := mr.NextPart()

		var logPayloadBatch []LogPayload
		err = json.NewDecoder(p).Decode(&logPayloadBatch)
		if err != nil {
			log.Fatal(err)
		}

		re := &LogResponse{Id: "log_id"}
		data, _ := json.Marshal(re)
		_, _ = w.Write(data)
	}))
	defer ts.Close()

	C = New(ts.URL, project, token, btsProject, false)
	C.Stack.Push(nil)
	C.Stack.Push("last_item_id")
	err := C.LogBatch([]LogPayload{
		{ItemId: "", Time: time.Now().String(), Message: "abc", Level: "INFO"},
		{ItemId: "", Time: time.Now().String(), Message: "def", Level: "INFO"},
	})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, C.Stack.Len())
}

func TestClient_LogNotAttachableToLaunchItem(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.String(), "/api/v1/testproj/log")

		var logPayload *LogPayload
		err := json.NewDecoder(r.Body).Decode(&logPayload)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, "last_item_id", logPayload.ItemId)
		assert.Equal(t, "DEBUG", logPayload.Level)
		assert.Equal(t, "logmessage", logPayload.Message)
		assert.NotNil(t, logPayload.Time)

		re := &LogResponse{Id: "log_id"}
		data, _ := json.Marshal(re)
		_, _ = w.Write(data)
	}))
	defer ts.Close()

	C = New(ts.URL, project, token, btsProject, false)
	C.Stack.Push(nil) // launch item
	_, err := C.Log("logmessage", "DEBUG")
	if err != nil {
		assert.Equal(t, "cannot attach log to launch item, only to test items", err.Error())
	}
	assert.Equal(t, 1, C.Stack.Len())
}

func TestClient_FailedRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.String(), "/api/v1/testproj/launch")

		var startLaunch *StartLaunchPayload
		err := json.NewDecoder(r.Body).Decode(&startLaunch)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, "testrun", startLaunch.Name)
		assert.NotNil(t, startLaunch.StartTime)
		assert.Equal(t, []string{"tag1"}, startLaunch.Tags)
		assert.Equal(t, "DEFAULT", startLaunch.Mode)

		re := &StartLaunchResponse{Number: 1, Id: "id1"}
		data, _ := json.Marshal(re)
		http.Error(w, "wrong data", http.StatusNotFound)
		_, _ = w.Write(data)
	}))
	defer ts.Close()
	C = New(ts.URL, project, token, btsProject, false)
	launchId, err := C.StartLaunch("testrun", "", "", []string{"tag1"}, "DEFAULT")
	assert.Error(t, err)
	assert.Empty(t, launchId)
	assert.Equal(t, 0, C.Stack.Len())
}

func TestOptions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.String(), "/api/v1/testproj/launch")

		var startLaunch *StartLaunchPayload
		err := json.NewDecoder(r.Body).Decode(&startLaunch)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, "testrun", startLaunch.Name)
		assert.NotNil(t, startLaunch.StartTime)
		assert.Equal(t, []string{"tag1"}, startLaunch.Tags)
		assert.Equal(t, "DEFAULT", startLaunch.Mode)

		re := &StartLaunchResponse{Number: 1, Id: "id1"}
		data, _ := json.Marshal(re)
		http.Error(w, "wrong data", http.StatusNotFound)
		_, _ = w.Write(data)
	}))
	defer ts.Close()
	C = New(ts.URL, project, token, btsProject, false, WithHttpClient(&http.Client{}), WithRetries(5), WithVerbosity("debug"))
	launchId, err := C.StartLaunch("testrun", "", "", []string{"tag1"}, "DEFAULT")
	assert.Error(t, err)
	assert.Empty(t, launchId)
	assert.Equal(t, 0, C.Stack.Len())
}
