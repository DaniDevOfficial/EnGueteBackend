package group

type GroupError struct {
	Error string `json:"error"`
}

type GroupSuccess struct {
	Message string `json:"message"`
}

type RequestNewGroup struct {
	GroupName string `json:"groupName"`
}

type ResponseNewGroup struct {
	GroupId string `json:"groupId"`
}
