package dto

// Task represents a background task
type Task struct {
	Id    string `json:"Id"`
	Name  string `json:"Name"`
	State string `json:"State"`
}

// TaskQueued represents a queued task response
type TaskQueued struct {
	Id       string `json:"Id"`
	Name     string `json:"Name"`
	QueueId  string `json:"QueueId"`
	Position int    `json:"Position"`
}

// TaskResponse is a generic task response
type TaskResponse struct {
	Id                        string `json:"Id"`
	Name                      string `json:"Name"`
	State                     string `json:"State"`
	CurrentProgressPercentage int    `json:"CurrentProgressPercentage,omitempty"`
}

// MetadataRefreshTask represents a metadata refresh task
type MetadataRefreshTask struct {
	ItemId    string `json:"ItemId"`
	UserId    string `json:"UserId"`
	Overwrite bool   `json:"Overwrite"`
	Recursive bool   `json:"Recursive"`
	Status    Task   `json:"Status"`
}

// MessageDto represents a message response
type MessageDto struct {
	Message string `json:"Message"`
	Result  string `json:"Result"`
}
