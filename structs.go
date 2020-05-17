package rpgoclient

type StartLaunchPayload struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	StartTime   string   `json:"start_time"`
	Mode        string   `json:"mode"`
}

type StartLaunchResponse struct {
	Number int    `json:"number"`
	Id     string `json:"id"`
}

type FinishLaunchPayload struct {
	Status  string `json:"status"`
	EndTime string `json:"end_time"`
}

type FinishLaunchResponse struct {
	Id     string `json:"id"`
	Link   string `json:"link"`
	Number int    `json:"number"`
}

type StartTestItemPayload struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Tags        []string            `json:"tags"`
	StartTime   string              `json:"start_time"`
	LaunchId    string              `json:"launch_id"`
	Type        string              `json:"type"`
	Parameters  []map[string]string `json:"parameters"`
}

type StartTestItemResponse struct {
	Id       string `json:"id"`
	UniqueId string `json:"uniqueId"`
}

type FinishTestItemPayload struct {
	Status  string                 `json:"status"`
	EndTime string                 `json:"end_time"`
	Issue   map[string]interface{} `json:"issue"`
}

type LinkIssue struct {
	Issues      []Issue `json:"issues"`
	TestItemIds []int   `json:"testItemIds"`
}

type Issue struct {
	BtsProject string `json:"btsProject"`
	BtsUrl     string `json:"btsUrl"`
	SubmitDate int64  `json:"submitDate"`
	TicketId   string `json:"ticketId"`
	Url        string `json:"url"`
}

type FinishTestItemResponse struct {
	Msg string `json:"msg"`
}

type LogPayload struct {
	ItemId  string `json:"item_id"`
	Time    string `json:"time"`
	Message string `json:"message"`
	Level   string `json:"level"`
}

type LogResponse struct {
	Id string `json:"id"`
}

type GetItemResponse struct {
	Id int `json:"id"`
}

type GetItemIdByUniqIdResponse struct {
	Content []ItemContent `json:"content"`
}

type ItemContent struct {
	Id int `json:"id"`
}
