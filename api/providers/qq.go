package providers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"web_music/models"
)

// 定义错误
var (
	ErrFetchFailed = errors.New("获取数据失败")
	ErrParseError  = errors.New("解析响应失败")
)

// QQ音乐API常量
const (
// 新版搜索API
// qqSearchAPI = "https://u.y.qq.com/cgi-bin/musicu.fcg"
)

// SearchQQMusic 搜索QQ音乐
func SearchQQMusic(keyword string) ([]models.Song, error) {
	maxRetries := 3
	var err error

	for i := 0; i < maxRetries; i++ {
		var songs []models.Song
		songs, err = trySearchQQMusic(keyword)
		if err == nil {
			return songs, nil
		}

		if i < maxRetries-1 {
			log.Printf("QQ音乐搜索失败，正在重试(%d/%d): %v", i+1, maxRetries, err)
			time.Sleep(time.Duration(i+1) * 2 * time.Second) // 指数退避
		}
	}

	return []models.Song{}, fmt.Errorf("多次尝试后搜索QQ音乐失败: %w", err)
}

// 将原来的SearchQQMusic函数代码移动到这个新函数中
func trySearchQQMusic(keyword string) ([]models.Song, error) {
	log.Printf("搜索QQ音乐: %s", keyword)

	// 构建请求URL
	// 接口文档：https://github.com/jsososo/QQMusicApi
	apiURL := "https://u.y.qq.com/cgi-bin/musicu.fcg"

	// 构建请求体
	requestBody := map[string]interface{}{
		"req_0": map[string]interface{}{
			"module": "music.search.SearchCgiService",
			"method": "DoSearchForQQMusicDesktop",
			"param": map[string]interface{}{
				"query":        keyword,
				"num_per_page": 20,
				"page_num":     1,
				"search_type":  0, // 0-歌曲，8-专辑，9-歌词，7-歌单，1-歌手，2-mv
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("构建请求体失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", "https://y.qq.com/")
	req.Header.Set("Origin", "https://y.qq.com")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
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
	log.Printf("QQ音乐搜索响应状态码: %d", resp.StatusCode)
	if len(body) > 500 {
		log.Printf("QQ音乐搜索响应体 (前500字符): %s", string(body[:500]))
	} else {
		log.Printf("QQ音乐搜索响应体: %s", string(body))
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	// 检查是否有返回结果
	req0, ok := result["req_0"].(map[string]interface{})
	if !ok || req0["code"].(float64) != 0 {
		code := 0.0
		if ok {
			code = req0["code"].(float64)
		}
		log.Printf("QQ音乐API响应错误: %v", code)
		return []models.Song{}, nil
	}

	// 提取歌曲列表
	var songs []models.Song
	if data, ok := req0["data"].(map[string]interface{}); ok {
		if body, ok := data["body"].(map[string]interface{}); ok {
			if song, ok := body["song"].(map[string]interface{}); ok {
				if list, ok := song["list"].([]interface{}); ok {
					for _, item := range list {
						songInfo, ok := item.(map[string]interface{})
						if !ok {
							continue
						}

						// 提取歌手信息
						var artists []string
						if singer, ok := songInfo["singer"].([]interface{}); ok {
							for _, s := range singer {
								if singerInfo, ok := s.(map[string]interface{}); ok {
									if name, ok := singerInfo["name"].(string); ok {
										artists = append(artists, name)
									}
								}
							}
						}

						// 提取专辑信息
						albumName := ""
						if album, ok := songInfo["album"].(map[string]interface{}); ok {
							if name, ok := album["name"].(string); ok {
								albumName = name
							}
						}

						// 提取歌曲封面
						coverURL := ""
						if album, ok := songInfo["album"].(map[string]interface{}); ok {
							if mid, ok := album["mid"].(string); ok {
								coverURL = fmt.Sprintf("https://y.gtimg.cn/music/photo_new/T002R300x300M000%s.jpg", mid)
							}
						}

						// 创建歌曲对象
						songMid := ""
						if mid, ok := songInfo["mid"].(string); ok {
							songMid = mid
						}

						songTitle := ""
						if title, ok := songInfo["title"].(string); ok {
							songTitle = title
						}

						if songMid != "" && songTitle != "" {
							song := models.Song{
								ID:     songMid,
								Title:  songTitle,
								Artist: strings.Join(artists, ", "),
								Album:  albumName,
								Cover:  coverURL,
								Source: "qq",
							}
							songs = append(songs, song)
						}
					}
				}
			}
		}
	}

	if len(songs) == 0 {
		log.Printf("QQ音乐搜索无结果: %s", keyword)
	}

	return songs, nil
}

// GetQQMusicURL 获取QQ音乐播放URL
func GetQQMusicURL(mid string) (string, error) {
	log.Printf("获取QQ音乐URL: %s", mid)

	// 构建请求URL
	apiURL := "https://u.y.qq.com/cgi-bin/musicu.fcg"

	// 构建请求体
	requestBody := map[string]interface{}{
		"req_0": map[string]interface{}{
			"module": "vkey.GetVkeyServer",
			"method": "CgiGetVkey",
			"param": map[string]interface{}{
				"guid":       "10000",
				"songmid":    []string{mid},
				"songtype":   []int{0},
				"uin":        "0",
				"loginflag":  1,
				"platform":   "20",
				"h5platform": "Android",
				"h5uin":      "0",
				"h5guid":     "10000",
				"h5channel":  "mqq",
				"h5version":  "1.0",
				"h5from":     "mqq",
				"h5tag":      "mqq",
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("构建请求体失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 11; Pixel 4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Mobile Safari/537.36")
	req.Header.Set("Referer", "https://y.qq.com/")
	req.Header.Set("Origin", "https://y.qq.com")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	var reader io.Reader = resp.Body

	// 处理gzip压缩
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return "", fmt.Errorf("创建gzip reader失败: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 调试输出
	log.Printf("QQ音乐URL获取响应状态码: %d", resp.StatusCode)
	if len(body) > 500 {
		log.Printf("QQ音乐URL获取响应体 (前500字符): %s", string(body[:500]))
	} else {
		log.Printf("QQ音乐URL获取响应体: %s", string(body))
	}

	// 清理响应体中的ANSI转义序列
	cleanBody := cleanANSI(body)

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(cleanBody, &result); err != nil {
		// 如果解析失败，尝试使用备用方法
		return getQQMusicURLAlternative(mid)
	}

	// 提取URL
	req0, ok := result["req_0"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("响应格式错误")
	}

	if req0["code"].(float64) != 0 {
		return "", fmt.Errorf("API返回错误代码: %v", req0["code"])
	}

	data, ok := req0["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("响应数据格式错误")
	}

	// 获取URL基础部分
	sip, ok := data["sip"].([]interface{})
	if !ok || len(sip) == 0 {
		return "", fmt.Errorf("无法获取URL基础部分")
	}
	urlBase := sip[0].(string)

	// 获取歌曲文件信息
	midurlinfo, ok := data["midurlinfo"].([]interface{})
	if !ok || len(midurlinfo) == 0 {
		return "", fmt.Errorf("无法获取歌曲文件信息")
	}

	info := midurlinfo[0].(map[string]interface{})
	purl, ok := info["purl"].(string)
	if !ok || purl == "" {
		// 尝试使用备用方法获取URL
		return getQQMusicURLAlternative(mid)
	}

	// 组合完整URL
	fullURL := urlBase + purl

	return fullURL, nil
}

// 清理ANSI转义序列
func cleanANSI(data []byte) []byte {
	// 移除ANSI转义序列的正则表达式
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return ansiRegex.ReplaceAll(data, []byte{})
}

// 备用的获取音乐URL方法
func getQQMusicURLAlternative(songMid string) (string, error) {
	log.Printf("使用备用方法获取QQ音乐URL: %s", songMid)

	// 构建请求URL - 使用不同的API端点
	apiURL := fmt.Sprintf("https://u.y.qq.com/cgi-bin/musics.fcg?format=json&data={%%22req_0%%22:{%%22module%%22:%%22vkey.GetVkeyServer%%22,%%22method%%22:%%22CgiGetVkey%%22,%%22param%%22:{%%22guid%%22:%%2210000%%22,%%22songmid%%22:[%%22%s%%22],%%22songtype%%22:[0],%%22uin%%22:%%220%%22,%%22loginflag%%22:1,%%22platform%%22:%%2220%%22}}}", songMid)

	// 创建请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建备用请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept-Encoding", "identity") // 禁用压缩

	// 发送请求
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			DisableCompression: true, // 禁用传输层压缩
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("备用请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取备用响应失败: %w", err)
	}

	// 调试输出
	log.Printf("QQ音乐备用URL获取响应状态码: %d", resp.StatusCode)
	log.Printf("QQ音乐备用URL获取响应头: %v", resp.Header)

	// 清理响应数据
	cleanedBody := bytes.TrimLeft(body, "\x00\x1b\x1f")        // 移除开头的特殊字符
	cleanedBody = bytes.TrimRight(cleanedBody, "\x00\x1b\x1f") // 移除结尾的特殊字符
	cleanedBody = cleanANSI(cleanedBody)                       // 清理ANSI转义序列

	// 检查响应是否为空
	if len(cleanedBody) == 0 {
		return "", fmt.Errorf("备用响应为空")
	}

	// 尝试解析JSON
	var result map[string]interface{}
	if err := json.Unmarshal(cleanedBody, &result); err != nil {
		log.Printf("备用响应解析失败，尝试第三方API: %v", err)
		return getQQMusicURLThirdOption(songMid)
	}

	// 提取URL
	req0, ok := result["req_0"].(map[string]interface{})
	if !ok {
		return getQQMusicURLThirdOption(songMid)
	}

	if code, ok := req0["code"].(float64); !ok || code != 0 {
		return getQQMusicURLThirdOption(songMid)
	}

	data, ok := req0["data"].(map[string]interface{})
	if !ok {
		return getQQMusicURLThirdOption(songMid)
	}

	// 获取URL基础部分
	sip, ok := data["sip"].([]interface{})
	if !ok || len(sip) == 0 {
		return getQQMusicURLThirdOption(songMid)
	}
	urlBase := sip[0].(string)

	// 获取歌曲文件信息
	midurlinfo, ok := data["midurlinfo"].([]interface{})
	if !ok || len(midurlinfo) == 0 {
		return getQQMusicURLThirdOption(songMid)
	}

	info := midurlinfo[0].(map[string]interface{})
	purl, ok := info["purl"].(string)
	if !ok || purl == "" {
		return getQQMusicURLThirdOption(songMid)
	}

	// 组合完整URL
	fullURL := urlBase + purl
	if fullURL == "" || !strings.HasPrefix(fullURL, "http") {
		return getQQMusicURLThirdOption(songMid)
	}

	return fullURL, nil
}

// 修改第三个备用方法
func getQQMusicURLThirdOption(songMid string) (string, error) {
	log.Printf("使用第三备用方法获取QQ音乐URL: %s", songMid)

	// 使用新的第三方API
	apiURL := fmt.Sprintf("https://api.zhuolin.wang/api.php?callback=jQuery&types=url&id=%s", songMid)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建第三备用请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Referer", "https://y.qq.com/")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Encoding", "identity")

	// 发送请求
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("第三备用请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取第三备用响应失败: %w", err)
	}

	// 从JSONP响应中提取JSON
	responseStr := string(body)
	jsonStr := strings.TrimPrefix(responseStr, "jQuery(")
	jsonStr = strings.TrimSuffix(jsonStr, ")")

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return "", fmt.Errorf("解析第三备用响应JSON失败: %w", err)
	}

	// 提取URL
	if url, ok := result["url"].(string); ok && url != "" {
		return url, nil
	}

	return "", fmt.Errorf("无法从第三备用方法获取音乐URL")
}

// 辅助函数：截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
