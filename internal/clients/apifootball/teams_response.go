package apifootball

type GetTeamsResponse struct {
	Get        string `json:"get"`
	Parameters struct {
		League string `json:"league"`
		Season string `json:"season"`
	} `json:"parameters"`
	Errors  []interface{} `json:"errors"`
	Results int           `json:"results"`
	Paging  struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"paging"`
	Data []struct {
		Team struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			Code     string `json:"code"`
			Country  string `json:"country"`
			Founded  int    `json:"founded"`
			National bool   `json:"national"`
			Logo     string `json:"logo"`
		} `json:"team"`
		Venue struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			Address  string `json:"address"`
			City     string `json:"city"`
			Capacity int    `json:"capacity"`
			Surface  string `json:"surface"`
			Image    string `json:"image"`
		} `json:"venue"`
	} `json:"response"`
}
