package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"web_music/api/providers"
	"web_music/models"
)

// 定义错误
var (
	ErrUnsupportedProvider = errors.New("不支持的音乐源")
	ErrFetchFailed         = errors.New("获取数据失败")
	ErrParseError          = errors.New("解析响应失败")
)

// SearchResponse 定义搜索结果响应结构
type SearchResponse struct {
	Songs []models.Song `json:"songs"`
	Total int           `json:"total"`
}

// SongURLResponse 定义获取歌曲URL的响应结构
type SongURLResponse struct {
	URL  string `json:"url"`
	Code int    `json:"code"`
	Msg  string `json:"msg,omitempty"`
}

// SearchHandler 处理音乐搜索请求
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	// 设置跨域头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 处理预检请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 仅支持GET请求
	if r.Method != "GET" {
		http.Error(w, "仅支持GET请求", http.StatusMethodNotAllowed)
		return
	}

	// 获取查询参数
	keyword := r.URL.Query().Get("keyword")
	sourcesParam := r.URL.Query().Get("sources")

	if keyword == "" {
		http.Error(w, "缺少关键词参数", http.StatusBadRequest)
		return
	}

	// 解析音乐源
	var sources []string
	if sourcesParam != "" {
		sources = strings.Split(sourcesParam, ",")
	} else {
		// 默认使用所有源
		sources = []string{"qq", "netease", "kuwo"}
	}

	// 搜索歌曲
	songs, err := searchAllProviders(keyword, sources)
	if err != nil {
		log.Printf("搜索失败: %v", err)
		http.Error(w, "搜索失败", http.StatusInternalServerError)
		return
	}

	// 构造响应
	response := SearchResponse{
		Songs: songs,
		Total: len(songs),
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SongHandler 处理获取歌曲URL的请求
func SongHandler(w http.ResponseWriter, r *http.Request) {
	// 设置跨域头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 处理预检请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 仅支持GET请求
	if r.Method != "GET" {
		http.Error(w, "仅支持GET请求", http.StatusMethodNotAllowed)
		return
	}

	// 获取查询参数
	id := r.URL.Query().Get("id")
	source := r.URL.Query().Get("source")

	if id == "" || source == "" {
		http.Error(w, "缺少必要参数", http.StatusBadRequest)
		return
	}

	// 获取歌曲URL
	url, err := getSongURL(id, source)
	if err != nil {
		log.Printf("获取歌曲URL失败: %v", err)

		// 返回错误响应
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SongURLResponse{
			URL:  "",
			Code: http.StatusInternalServerError,
			Msg:  "获取歌曲URL失败",
		})
		return
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SongURLResponse{
		URL:  url,
		Code: http.StatusOK,
	})
}

// searchAllProviders 从多个音乐源搜索歌曲
func searchAllProviders(keyword string, sources []string) ([]models.Song, error) {
	var allSongs []models.Song

	// 遍历所有请求的音乐源
	for _, source := range sources {
		var songs []models.Song
		var err error

		switch source {
		case "qq":
			songs, err = providers.SearchQQMusic(keyword)
		case "netease":
			songs, err = providers.SearchNetease(keyword)
		case "kuwo":
			songs, err = providers.SearchKuwo(keyword)
		default:
			log.Printf("不支持的音乐源: %s", source)
			continue
		}

		if err != nil {
			log.Printf("从 %s 搜索失败: %v", source, err)
			continue
		}

		// 合并结果
		allSongs = append(allSongs, songs...)
	}

	return allSongs, nil
}

// getSongURL 获取歌曲的URL
func getSongURL(id, source string) (string, error) {
	switch source {
	case "qq":
		return providers.GetQQMusicURL(id)
	case "netease":
		return providers.GetNeteaseURL(id)
	case "kuwo":
		return providers.GetKuwoURL(id)
	default:
		return "", ErrUnsupportedProvider
	}
}
