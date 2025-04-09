package providers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"web_music/models"
)

// 网易云音乐API常量
const (
	neteaseBaseURL   = "https://music.163.com/weapi"
	neteaseSearchURL = neteaseBaseURL + "/cloudsearch/get/web"
	neteaseSongURL   = neteaseBaseURL + "/song/enhance/player/url/v1"
	neteasePublicKey = "010001"
	neteaseModulus   = "00e0b509f6259df8642dbc35662901477df22677ec152b5ff68ace615bb7b725152b3ab17a876aea8a5aa76d2e417629ec4ee341f56135fccf695280104e0312ecbda92557c93870114af6c9d05c4f7f0c3685b7a46bee255932575cce10b424d813cfe4875d3e82047b97ddef52741d546b8e289dc6935b3ece0462db0a22b8e7"
	neteaseIV        = "0102030405060708"
	neteasePresetKey = "0CoJUm6Qyw8W8jud"
	neteaseUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
)

var neteaseClient *http.Client

func init() {
	jar, _ := cookiejar.New(nil)
	neteaseClient = &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}

	// 预先访问一次主页获取必要的Cookie
	initNeteaseSession()
}

// 初始化网易云会话，获取必要Cookie
func initNeteaseSession() {
	req, err := http.NewRequest("GET", "https://music.163.com", nil)
	if err != nil {
		log.Printf("创建网易云初始化请求失败: %v", err)
		return
	}

	req.Header.Set("User-Agent", neteaseUserAgent)

	_, err = neteaseClient.Do(req)
	if err != nil {
		log.Printf("初始化网易云会话失败: %v", err)
	}
}

// 重命名为带Legacy后缀
// SearchNeteaseOriginal 搜索网易云音乐(原始API)
func SearchNeteaseOriginal(keyword string) ([]models.Song, error) {
	log.Printf("搜索网易云音乐(改用第三方接口): %s", keyword)

	// 使用开放接口
	apiURL := fmt.Sprintf("https://netease-cloud-music-api-taupe-nine.vercel.app/search?keywords=%s&limit=20", keyword)

	// 创建请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
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

	// 解析响应
	var result struct {
		Code   int `json:"code"`
		Result struct {
			Songs []struct {
				ID      int    `json:"id"`
				Name    string `json:"name"`
				Artists []struct {
					Name string `json:"name"`
				} `json:"artists"`
				Album struct {
					Name   string `json:"name"`
					PicURL string `json:"picUrl"`
				} `json:"album"`
			} `json:"songs"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	// 检查响应状态
	if result.Code != 200 {
		return nil, fmt.Errorf("API返回错误代码: %d", result.Code)
	}

	// 转换为通用格式
	var songs []models.Song
	for _, item := range result.Result.Songs {
		// 构建歌手名称
		var artists []string
		for _, artist := range item.Artists {
			artists = append(artists, artist.Name)
		}
		artistName := strings.Join(artists, ", ")

		song := models.Song{
			ID:     fmt.Sprintf("%d", item.ID),
			Title:  item.Name,
			Artist: artistName,
			Album:  item.Album.Name,
			Cover:  item.Album.PicURL,
			Source: "netease",
		}
		songs = append(songs, song)
	}

	return songs, nil
}

// 重命名为带Legacy后缀
// GetNeteaseURLOriginal 获取网易云音乐的播放URL(原始API)
func GetNeteaseURLOriginal(id string) (string, error) {
	log.Printf("获取网易云音乐URL(改用第三方接口): %s", id)

	// 使用开放接口
	apiURL := fmt.Sprintf("https://netease-cloud-music-api-taupe-nine.vercel.app/song/url?id=%s", id)

	// 创建请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var result struct {
		Code int `json:"code"`
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析JSON失败: %w", err)
	}

	// 检查响应和URL
	if result.Code != 200 || len(result.Data) == 0 || result.Data[0].URL == "" {
		return "", fmt.Errorf("未获取到音乐URL")
	}

	return result.Data[0].URL, nil
}

// 从Cookies中获取CSRF Token
func getCsrfFromCookies(domain string) string {
	neteaseURL, _ := url.Parse("https://" + domain)
	for _, cookie := range neteaseClient.Jar.Cookies(neteaseURL) {
		if cookie.Name == "__csrf" {
			return cookie.Value
		}
	}
	return ""
}

// 网易云音乐加密参数函数
func encryptNetease(params map[string]interface{}) (url.Values, error) {
	// 将参数转换为JSON字符串
	jsonData, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	// 生成16位随机字符串作为密钥
	secretKey := randomString(16)

	// 第一次AES加密
	encText := aesEncrypt(string(jsonData), neteasePresetKey)

	// 第二次AES加密
	encText = aesEncrypt(encText, secretKey)

	// RSA加密密钥
	encSecKey := rsaEncrypt(secretKey, neteasePublicKey, neteaseModulus)

	// 构建表单数据
	data := url.Values{}
	data.Set("params", encText)
	data.Set("encSecKey", encSecKey)

	return data, nil
}

// AES加密
func aesEncrypt(text, key string) string {
	// 创建加密器
	block, _ := aes.NewCipher([]byte(key))
	// 填充
	blockSize := block.BlockSize()
	padding := blockSize - len(text)%blockSize
	padText := make([]byte, len(text)+padding)
	copy(padText, []byte(text))
	for i := 0; i < padding; i++ {
		padText[len(text)+i] = byte(padding)
	}

	// 加密
	blockMode := cipher.NewCBCEncrypter(block, []byte(neteaseIV))
	crypted := make([]byte, len(padText))
	blockMode.CryptBlocks(crypted, padText)

	// Base64编码
	return base64.StdEncoding.EncodeToString(crypted)
}

// RSA加密
func rsaEncrypt(text, publicKey, modulus string) string {
	// 翻转文本
	rText := []byte(text)
	for i, j := 0, len(rText)-1; i < j; i, j = i+1, j-1 {
		rText[i], rText[j] = rText[j], rText[i]
	}

	// 将文本转换为bigint
	textInt := new(big.Int)
	textInt.SetBytes(rText)

	// 将模数和公钥转换为bigint
	modulusInt, _ := new(big.Int).SetString(modulus, 16)
	publicKeyInt, _ := new(big.Int).SetString(publicKey, 16)

	// RSA加密
	encryptedInt := new(big.Int).Exp(textInt, publicKeyInt, modulusInt)

	// 将结果转换为十六进制字符串
	result := fmt.Sprintf("%0256x", encryptedInt)
	return result
}

// 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// MD5加密
func md5Encrypt(text string) string {
	hash := md5.New()
	hash.Write([]byte(text))
	return hex.EncodeToString(hash.Sum(nil))
}
