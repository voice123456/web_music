package models

// Song 表示一首歌曲的信息
type Song struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Album  string `json:"album,omitempty"`
	Cover  string `json:"cover,omitempty"`
	Source string `json:"source"`
	URL    string `json:"url,omitempty"`
}
