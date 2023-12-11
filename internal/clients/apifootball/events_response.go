package apifootball

type Event struct {
	Time struct {
		Elapsed int `json:"elapsed"`
		Extra   int `json:"extra"`
	} `json:"time"`
	Team struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Logo string `json:"logo"`
	} `json:"team"`
	Player struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"player"`
	Assist struct {
		ID   any `json:"id"`
		Name any `json:"name"`
	} `json:"assist"`
	Type     string `json:"type"`
	Detail   string `json:"detail"`
	Comments string `json:"comments"`
}

type GetEventsResponse struct {
	Get        string `json:"get"`
	Parameters struct {
		Fixture string `json:"fixture"`
	} `json:"parameters"`
	Errors  []any `json:"errors"`
	Results int   `json:"results"`
	Paging  struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"paging"`
	Data []Event `json:"response"`
}
