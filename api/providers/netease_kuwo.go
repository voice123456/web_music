package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"web_music/models"
)

// 使用国内稳定的音乐API聚合服务
const (
	// 此API支持网易云、QQ音乐和酷我等多个平台
	apiBaseURL = "https://api.music.itooi.cn/v1"
)

// SearchNetease 搜索网易云音乐
func SearchNetease(keyword string) ([]models.Song, error) {
	log.Printf("搜索网易云音乐(使用聚合API): %s", keyword)

	// 尝试使用稳定的第三方API
	apiURL := fmt.Sprintf("https://music.163.com/api/search/get?s=%s&type=1&limit=20&offset=0", url.QueryEscape(keyword))

	// 创建请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建网易云请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", "https://music.163.com/")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("网易云API请求失败: %v，尝试备用API", err)
		return searchNeteaseBackup(keyword)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取网易云响应失败: %v，尝试备用API", err)
		return searchNeteaseBackup(keyword)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("解析网易云响应失败: %v，尝试备用API", err)
		return searchNeteaseBackup(keyword)
	}

	// 检查响应状态
	code, ok := result["code"].(float64)
	if !ok || code != 200 {
		log.Printf("网易云API返回错误码: %v，尝试备用API", code)
		return searchNeteaseBackup(keyword)
	}

	// 提取歌曲列表
	var songs []models.Song
	if resultObj, ok := result["result"].(map[string]interface{}); ok {
		if songsObj, ok := resultObj["songs"].([]interface{}); ok {
			for _, item := range songsObj {
				song, ok := item.(map[string]interface{})
				if !ok {
					continue
				}

				// 获取歌曲ID
				id, ok := song["id"].(float64)
				if !ok {
					continue
				}

				// 获取歌曲名称
				name, ok := song["name"].(string)
				if !ok {
					continue
				}

				// 获取歌手信息
				var artists []string
				if artistsObj, ok := song["artists"].([]interface{}); ok {
					for _, a := range artistsObj {
						artist, ok := a.(map[string]interface{})
						if !ok {
							continue
						}
						if artistName, ok := artist["name"].(string); ok {
							artists = append(artists, artistName)
						}
					}
				}

				// 获取专辑信息
				albumName := ""
				coverURL := ""
				if album, ok := song["album"].(map[string]interface{}); ok {
					if aName, ok := album["name"].(string); ok {
						albumName = aName
					}
					if picUrl, ok := album["picUrl"].(string); ok {
						coverURL = picUrl
					}
				}

				// 创建歌曲对象
				songs = append(songs, models.Song{
					ID:     fmt.Sprintf("%.0f", id),
					Title:  name,
					Artist: strings.Join(artists, ", "),
					Album:  albumName,
					Cover:  coverURL,
					Source: "netease",
				})
			}
		}
	}

	if len(songs) == 0 {
		log.Printf("网易云音乐搜索无结果，尝试备用API")
		return searchNeteaseBackup(keyword)
	}

	return songs, nil
}

// 备用网易云搜索API
func searchNeteaseBackup(keyword string) ([]models.Song, error) {
	log.Printf("使用备用API搜索 netease: %s", keyword)

	apiURL := fmt.Sprintf("https://musicapi.leanapp.cn/search?keywords=%s&limit=20", url.QueryEscape(keyword))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return []models.Song{}, fmt.Errorf("创建备用网易云请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")

	// 发送请求
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return []models.Song{}, nil // 返回空结果而不是错误
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []models.Song{}, nil
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return []models.Song{}, nil
	}

	// 检查响应状态
	code, ok := result["code"].(float64)
	if !ok || code != 200 {
		return []models.Song{}, nil
	}

	// 提取歌曲列表
	var songs []models.Song
	if resultObj, ok := result["result"].(map[string]interface{}); ok {
		if songsObj, ok := resultObj["songs"].([]interface{}); ok {
			for _, item := range songsObj {
				song, ok := item.(map[string]interface{})
				if !ok {
					continue
				}

				// 获取歌曲ID
				id, ok := song["id"].(float64)
				if !ok {
					continue
				}

				// 获取歌曲名称
				name, ok := song["name"].(string)
				if !ok {
					continue
				}

				// 获取歌手信息
				var artists []string
				if artistsObj, ok := song["artists"].([]interface{}); ok {
					for _, a := range artistsObj {
						artist, ok := a.(map[string]interface{})
						if !ok {
							continue
						}
						if artistName, ok := artist["name"].(string); ok {
							artists = append(artists, artistName)
						}
					}
				}

				// 获取专辑信息
				albumName := ""
				coverURL := ""
				if album, ok := song["album"].(map[string]interface{}); ok {
					if aName, ok := album["name"].(string); ok {
						albumName = aName
					}
					if picUrl, ok := album["picUrl"].(string); ok {
						coverURL = picUrl
					}
				}

				// 创建歌曲对象
				songs = append(songs, models.Song{
					ID:     fmt.Sprintf("%.0f", id),
					Title:  name,
					Artist: strings.Join(artists, ", "),
					Album:  albumName,
					Cover:  coverURL,
					Source: "netease",
				})
			}
		}
	}

	return songs, nil
}

// GetNeteaseURL 获取网易云音乐URL
func GetNeteaseURL(id string) (string, error) {
	log.Printf("获取网易云音乐URL: %s", id)

	apiURL := fmt.Sprintf("https://music.163.com/api/song/enhance/player/url?ids=[%s]&br=320000", id)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建网易云URL请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", "https://music.163.com/")

	// 发送请求
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("获取网易云URL失败: %v，尝试备用API", err)
		return getNeteaseURLBackup(id)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取网易云URL响应失败: %v，尝试备用API", err)
		return getNeteaseURLBackup(id)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("解析网易云URL响应失败: %v，尝试备用API", err)
		return getNeteaseURLBackup(id)
	}

	// 检查响应状态
	code, ok := result["code"].(float64)
	if !ok || code != 200 {
		log.Printf("网易云URL API返回错误码: %v，尝试备用API", code)
		return getNeteaseURLBackup(id)
	}

	// 提取URL
	if data, ok := result["data"].([]interface{}); ok && len(data) > 0 {
		if item, ok := data[0].(map[string]interface{}); ok {
			if url, ok := item["url"].(string); ok && url != "" {
				return url, nil
			}
		}
	}

	return getNeteaseURLBackup(id)
}

// 备用网易云音乐URL获取API
func getNeteaseURLBackup(id string) (string, error) {
	log.Printf("使用备用API获取网易云URL: %s", id)

	apiURL := fmt.Sprintf("https://musicapi.leanapp.cn/song/url?id=%s", id)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建备用网易云URL请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")

	// 发送请求
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("备用网易云URL请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取备用网易云URL响应失败: %w", err)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析备用网易云URL响应失败: %w", err)
	}

	// 检查响应状态
	code, ok := result["code"].(float64)
	if !ok || code != 200 {
		return "", fmt.Errorf("备用网易云URL API返回错误码: %v", code)
	}

	// 提取URL
	if data, ok := result["data"].([]interface{}); ok && len(data) > 0 {
		if item, ok := data[0].(map[string]interface{}); ok {
			if url, ok := item["url"].(string); ok && url != "" {
				return url, nil
			}
		}
	}

	return "", fmt.Errorf("无法获取网易云音乐URL")
}

// SearchKuwo 搜索酷我音乐
func SearchKuwo(keyword string) ([]models.Song, error) {
	log.Printf("搜索酷我音乐(使用聚合API): %s", keyword)

	// 构建请求
	apiURL := fmt.Sprintf("http://www.kuwo.cn/api/www/search/searchMusicBykeyWord?key=%s&pn=1&rn=20", url.QueryEscape(keyword))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建酷我请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", "http://www.kuwo.cn/search/list?key="+url.QueryEscape(keyword))
	req.Header.Set("Cookie", "kw_token=JQOEP7QK8RS")
	req.Header.Set("csrf", "JQOEP7QK8RS")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("酷我API请求失败: %v，尝试备用API", err)
		return searchKuwoBackup(keyword)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取酷我响应失败: %v，尝试备用API", err)
		return searchKuwoBackup(keyword)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("解析酷我响应失败: %v，尝试备用API", err)
		return searchKuwoBackup(keyword)
	}

	// 检查响应状态
	code, ok := result["code"].(float64)
	if !ok || code != 200 {
		log.Printf("酷我API返回错误码: %v，尝试备用API", code)
		return searchKuwoBackup(keyword)
	}

	// 提取歌曲列表
	var songs []models.Song
	if data, ok := result["data"].(map[string]interface{}); ok {
		if list, ok := data["list"].([]interface{}); ok {
			for _, item := range list {
				song, ok := item.(map[string]interface{})
				if !ok {
					continue
				}

				// 获取歌曲ID
				rid, ok := song["rid"].(float64)
				if !ok {
					continue
				}

				// 获取歌曲名称
				name, ok := song["name"].(string)
				if !ok {
					continue
				}

				// 获取歌手
				artist, ok := song["artist"].(string)
				if !ok {
					artist = ""
				}

				// 获取专辑
				album, ok := song["album"].(string)
				if !ok {
					album = ""
				}

				// 获取封面
				pic, ok := song["pic"].(string)
				if !ok {
					pic = ""
				}

				// 创建歌曲对象
				songs = append(songs, models.Song{
					ID:     fmt.Sprintf("%.0f", rid),
					Title:  name,
					Artist: artist,
					Album:  album,
					Cover:  pic,
					Source: "kuwo",
				})
			}
		}
	}

	if len(songs) == 0 {
		log.Printf("酷我音乐搜索无结果，尝试备用API")
		return searchKuwoBackup(keyword)
	}

	return songs, nil
}

// 备用酷我搜索API
func searchKuwoBackup(keyword string) ([]models.Song, error) {
	log.Printf("使用备用API搜索 kuwo: %s", keyword)

	apiURL := fmt.Sprintf("https://musicapi.leanapp.cn/search?keywords=%s&type=1002&limit=20", url.QueryEscape(keyword))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return []models.Song{}, fmt.Errorf("创建备用酷我请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")

	// 发送请求
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return []models.Song{}, nil // 返回空结果而不是错误
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []models.Song{}, nil
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return []models.Song{}, nil
	}

	// 提取歌曲列表
	var songs []models.Song
	if resultObj, ok := result["result"].(map[string]interface{}); ok {
		if songs, ok := resultObj["songs"].([]interface{}); ok {
			for _, item := range songs {
				song, ok := item.(map[string]interface{})
				if !ok {
					continue
				}

				// 获取歌曲ID
				id, ok := song["id"].(float64)
				if !ok {
					continue
				}

				// 获取歌曲名称
				name, ok := song["name"].(string)
				if !ok {
					continue
				}

				// 获取歌手信息
				var artists []string
				if artistsObj, ok := song["ar"].([]interface{}); ok {
					for _, a := range artistsObj {
						artist, ok := a.(map[string]interface{})
						if !ok {
							continue
						}
						if artistName, ok := artist["name"].(string); ok {
							artists = append(artists, artistName)
						}
					}
				}

				// 获取专辑信息
				albumName := ""
				coverURL := ""
				if album, ok := song["al"].(map[string]interface{}); ok {
					if aName, ok := album["name"].(string); ok {
						albumName = aName
					}
					if picUrl, ok := album["picUrl"].(string); ok {
						coverURL = picUrl
					}
				}

				// 创建歌曲对象
				songs = append(songs, models.Song{
					ID:     fmt.Sprintf("%.0f", id),
					Title:  name,
					Artist: strings.Join(artists, ", "),
					Album:  albumName,
					Cover:  coverURL,
					Source: "kuwo",
				})
			}
		}
	}

	return songs, nil
}

// GetKuwoURL 获取酷我音乐URL
func GetKuwoURL(id string) (string, error) {
	log.Printf("获取酷我音乐URL: %s", id)

	apiURL := fmt.Sprintf("http://www.kuwo.cn/api/v1/www/music/playUrl?mid=%s&type=convert_url3&br=320kmp3", id)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建酷我URL请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", "http://www.kuwo.cn/")
	req.Header.Set("Cookie", "kw_token=JQOEP7QK8RS")
	req.Header.Set("csrf", "JQOEP7QK8RS")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("获取酷我URL失败: %v，尝试备用API", err)
		return getKuwoURLBackup(id)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取酷我URL响应失败: %v，尝试备用API", err)
		return getKuwoURLBackup(id)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("解析酷我URL响应失败: %v，尝试备用API", err)
		return getKuwoURLBackup(id)
	}

	// 检查响应状态
	code, ok := result["code"].(float64)
	if !ok || code != 200 {
		log.Printf("酷我URL API返回错误码: %v，尝试备用API", code)
		return getKuwoURLBackup(id)
	}

	// 提取URL
	if data, ok := result["data"].(map[string]interface{}); ok {
		if url, ok := data["url"].(string); ok && url != "" {
			return url, nil
		}
	}

	return getKuwoURLBackup(id)
}

// 备用酷我音乐URL获取API
func getKuwoURLBackup(id string) (string, error) {
	log.Printf("使用备用API获取酷我URL: %s", id)

	apiURL := fmt.Sprintf("https://musicapi.leanapp.cn/song/url?id=%s&source=kuwo", id)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建备用酷我URL请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")

	// 发送请求
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("备用酷我URL请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取备用酷我URL响应失败: %w", err)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析备用酷我URL响应失败: %w", err)
	}

	// 检查响应状态
	code, ok := result["code"].(float64)
	if !ok || code != 200 {
		return "", fmt.Errorf("备用酷我URL API返回错误码: %v", code)
	}

	// 提取URL
	if data, ok := result["data"].([]interface{}); ok && len(data) > 0 {
		if item, ok := data[0].(map[string]interface{}); ok {
			if url, ok := item["url"].(string); ok && url != "" {
				return url, nil
			}
		}
	}

	return "", fmt.Errorf("无法获取酷我音乐URL")
}
