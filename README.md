#### Go client for Report Portal

Examples:
```go
package main

import (
	"github.com/f4hrenh9it/rpgoclient"
	"net/http"
)

func main() {
	c := rpgoclient.New(
		"http://localhost:8080",
		"superadmin_personal",
		"e4f04653-7666-4b77-81ce-c7b584215123",
		rpgoclient.WithHttpClient(&http.Client{}),
		rpgoclient.WithRetries(5),
		rpgoclient.WithVerbosity("debug"),
		)
    
	c.StartLaunch("testrun", "test launch", "", []string{"tag1"}, "DEFAULT")
    
	params := make([]map[string]string, 0)
	params = append(params, map[string]string{"key": "sdf", "value": "sdF"})
    
	c.StartTestItem("test_item_1", "SUITE", "","description root", []string{"tag1"}, params)
	c.StartTestItem("test_item_2_child", "SUITE", "", "description child", []string{"tag1"}, params)
	c.FinishTestItem("FAILED", "", nil)
	c.FinishTestItem("FAILED", "", nil)
    
	c.StartTestItem("test_item_3", "SUITE", "","description", []string{"tag1"}, params)
	c.FinishTestItem("PASSED", "",nil)
    
	c.FinishLaunch( "FAILED", "")
    
	// you can use methods with Id suffix if you need parallel stateless client
	id, _, := c.StartTestItemId("parent_item_id","test_item_1", "SUITE", "","description root", []string{"tag1"}, params)
	c.LogId(id, "logmsg", "DEBUG")
	c.FinishTestItemId(id,"FAILED", "", nil)
}

```