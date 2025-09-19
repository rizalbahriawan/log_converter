package response

type LogActResponse struct {
	Data []Activity `json:"data"`
}

type Activity struct {
	ID             int    `json:"id"`
	DateString     string `json:"dateString"`
	ActivityDetail string `json:"activityDetail"`
	Duration       int    `json:"duration"`
	Overtime       int    `json:"overtime"`
	ProjectName    string `json:"projectName"`
}

type ProjectTableResponse struct {
	Data []ProjectResponse `json:"data"`
}

type ProjectResponse struct {
	ProjectName string `json:"projectName"`
}
