package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"web_music/models"
)

// 酷我音乐API常量
const (
	kuwoSearchURL = "http://www.kuwo.cn/api/www/search/searchMusicBykeyWord"
	kuwoTokenURL  = "http://www.kuwo.cn/search/key"
	kuwoPlayURL   = "http://www.kuwo.cn/api/v1/www/music/playUrl"
	kuwoUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

	// 备选API地址
	kuwoMobileSearchURL = "http://mobilecdnbj.kugou.com/api/v3/search/song"
	// kuwoMobilePlayURL   = "http://antiserver.kuwo.cn/anti.s?type=convert_url&format=mp3&response=url&rid=MUSIC_%s"
)

// 酷我音乐Token
var (
	kuwoToken    string
	kuwoTokenExp time.Time
	kuwoClient   *http.Client
	retryCount   = 0
)

func init() {
	// 创建Cookie Jar
	jar, _ := cookiejar.New(nil)

	// 创建HTTP客户端
	kuwoClient = &http.Client{
		Jar:     jar,
		Timeout: 15 * time.Second,
	}

	// 初始获取一次Token
	refreshKuwoToken()
}

// 刷新酷我音乐Token
func refreshKuwoToken() {
	log.Println("刷新酷我音乐Token")

	// 创建请求
	req, err := http.NewRequest("GET", kuwoTokenURL, nil)
	if err != nil {
		log.Printf("创建Token请求失败: %v", err)
		return
	}

	// 设置请求头
	req.Header.Set("User-Agent", kuwoUserAgent)
	req.Header.Set("Referer", "http://www.kuwo.cn/")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Connection", "keep-alive")

	// 发送请求
	resp, err := kuwoClient.Do(req)
	if err != nil {
		log.Printf("请求Token失败: %v", err)
		return
	}
	defer resp.Body.Close()

	// 从Cookie中获取csrf Token
	for _, cookie := range kuwoClient.Jar.Cookies(req.URL) {
		if cookie.Name == "kw_token" {
			kuwoToken = cookie.Value
			break
		}
	}

	// 设置Token过期时间
	kuwoTokenExp = time.Now().Add(30 * time.Minute)
	log.Printf("成功获取酷我Token: %s", kuwoToken)
}

// 使用标准API搜索
func searchKuwoStandard(keyword string) ([]models.Song, error) {
	// 构建请求URL
	apiURL := fmt.Sprintf("%s?key=%s&pn=1&rn=20", kuwoSearchURL, url.QueryEscape(keyword))

	// 创建请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", kuwoUserAgent)
	req.Header.Set("Referer", "http://www.kuwo.cn/search/list?key="+url.QueryEscape(keyword))
	req.Header.Set("csrf", kuwoToken)
	req.Header.Set("Cookie", "kw_token="+kuwoToken)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")

	// 发送请求
	resp, err := kuwoClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 调试输出
	log.Printf("酷我搜索响应: %s", string(body))

	// 解析响应
	var result struct {
		Code int `json:"code"`
		Data struct {
			Total string `json:"total"`
			List  []struct {
				RID       string `json:"rid"`
				Name      string `json:"name"`
				Artist    string `json:"artist"`
				AlbumName string `json:"album"`
				PicURL    string `json:"pic"`
				MusicRID  string `json:"musicrid"`
			} `json:"list"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	// 检查响应状态
	if result.Code != 200 {
		return nil, fmt.Errorf("酷我音乐API返回错误代码: %d", result.Code)
	}

	// 检查是否有结果
	if len(result.Data.List) == 0 {
		return []models.Song{}, nil // 返回空数组而不是错误
	}

	// 转换为通用格式
	var songs []models.Song
	for _, item := range result.Data.List {
		// 提取音乐ID
		musicID := strings.TrimPrefix(item.MusicRID, "MUSIC_")

		// 构建歌曲
		song := models.Song{
			ID:     musicID,
			Title:  item.Name,
			Artist: item.Artist,
			Album:  item.AlbumName,
			Cover:  item.PicURL,
			Source: "kuwo",
		}
		songs = append(songs, song)
	}

	return songs, nil
}

// 使用移动端API搜索
func searchKuwoMobile(keyword string) ([]models.Song, error) {
	// 构建请求参数
	params := url.Values{}
	params.Set("keyword", keyword)
	params.Set("page", "1")
	params.Set("pagesize", "20")
	params.Set("showtype", "1")

	// 创建请求
	req, err := http.NewRequest("GET", kuwoMobileSearchURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 10; SM-G981B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.132 Mobile Safari/537.36")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var result struct {
		Status int `json:"status"`
		Data   struct {
			Total int `json:"total"`
			Info  []struct {
				Hash       string `json:"hash"`
				SongName   string `json:"songname"`
				SingerName string `json:"singername"`
				AlbumName  string `json:"album_name"`
				AlbumID    string `json:"album_id"`
				FileHash   string `json:"FileHash"`
			} `json:"info"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	// 检查响应状态
	if result.Status != 1 {
		return nil, fmt.Errorf("酷我移动端API返回错误状态: %d", result.Status)
	}

	// 转换为通用格式
	var songs []models.Song
	for _, item := range result.Data.Info {
		song := models.Song{
			ID:     item.FileHash,
			Title:  item.SongName,
			Artist: item.SingerName,
			Album:  item.AlbumName,
			Source: "kuwo",
		}
		songs = append(songs, song)
	}

	return songs, nil
}
