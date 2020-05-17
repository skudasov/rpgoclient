package rpgoclient

import (
	"log"
	"testing"
)

func TestClientGetItem(t *testing.T) {
	c := New("https://rp.dev.insolar.io/", "skudasov_personal", "f6f757d0-f9b5-4188-b950-41a8b89492e1", "jj", "btsUrl", true)
	_, err := c.GetItemIdByUUID("5f304ca4-0be6-41af-8078-03b6a29a0a2d")
	if err != nil {
		log.Fatal(err)
	}
	//_, err := c.GetItemIdByUniqId("368", "auto:58fc6db8f10ae27c6c49bf4e6ecb2b1c")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//res, err := c.StartLaunch("abc", "abc", time.Now().Format(time.RFC3339), []string{}, "DEFAULT")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Printf("res: %v\n", res)
}
