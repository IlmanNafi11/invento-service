package dto

type StatisticData struct {
	TotalProject *int `json:"total_project,omitempty"`
	TotalModul   *int `json:"total_modul,omitempty"`
	TotalUser    *int `json:"total_user,omitempty"`
	TotalRole    *int `json:"total_role,omitempty"`
}

type StatisticResponse struct {
	Data StatisticData `json:"data"`
}
