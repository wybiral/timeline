package types

type Update struct {
	URL    string `json:"url"`
	Body   string `json:"body"`
	Date   string `json:"date"`
	Title  string `json:"title"`
	Thumb  string `json:"thumb"`
	Source struct {
		URL      string `json:"url"`
		Name     string `json:"name"`
		Category string `json:"category"`
	} `json:"source"`
	Timestamp int64 `json:"timestamp"`
}
