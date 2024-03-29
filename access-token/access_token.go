package access_token

import (
	"encoding/json"
	"fmt"
	"github.com/HeartGarlic/douyin-openapi/cache"
	"github.com/HeartGarlic/douyin-openapi/util"
	"sync"
	"time"
)

// 正式地址
const accessTokenURL = "https://developer.toutiao.com/api/apps/v2/token"

// 沙盒地址
const sandBoxTokenURL = "https://open-sandbox.douyin.com/api/apps/v2/token"

// AccessToken 管理AccessToken 的基础接口
type AccessToken interface {
	GetCacheKey() string             // 获取缓存的key
	SetCacheKey(key string)          // 设置缓存key
	GetAccessToken() (string, error) // 获取token
}

// accessTokenLock 初始化全局锁防止并发获取token
var accessTokenLock = new(sync.Mutex)

// DefaultAccessToken 默认的token管理类
type DefaultAccessToken struct {
	AppId               string      // app_id	string	是	小程序的 app_id
	AppSecret           string      // app_secret	string	是	小程序的密钥
	GrantType           string      // grant_type	string	是	固定值“client_credentials”
	Cache               cache.Cache // 缓存组件
	accessTokenLock     *sync.Mutex // 读写锁
	accessTokenCacheKey string      // 缓存的key
	SandBox             bool        // 是否沙盒地址 默认 false 线上地址
}

// NewDefaultAccessToken 实例化默认的token管理类
func NewDefaultAccessToken(appId, appSecret string, cache cache.Cache, IsSandbox bool) AccessToken {
	if cache == nil {
		panic(any("cache is need"))
	}
	token := &DefaultAccessToken{
		AppId:               appId,
		AppSecret:           appSecret,
		GrantType:           "client_credential",
		Cache:               cache,
		accessTokenCacheKey: fmt.Sprintf("douyin_openapi_access_token_%s", appId),
		accessTokenLock:     accessTokenLock,
		SandBox:             IsSandbox,
	}
	return token
}

// GetCacheKey 获取缓存key
func (dd *DefaultAccessToken) GetCacheKey() string {
	return dd.accessTokenCacheKey
}

// SetCacheKey 设置缓存key
func (dd *DefaultAccessToken) SetCacheKey(key string) {
	dd.accessTokenCacheKey = key
}

// GetAccessToken 获取token
func (dd *DefaultAccessToken) GetAccessToken() (string, error) {
	// 先尝试从缓存中获取如果不存在就调用接口获取
	if val := dd.Cache.Get(dd.GetCacheKey()); val != nil {
		return val.(string), nil
	}

	// 加锁防止并发获取接口
	dd.accessTokenLock.Lock()
	defer dd.accessTokenLock.Unlock()

	// 双捡防止重复获取
	if val := dd.Cache.Get(dd.GetCacheKey()); val != nil {
		return val.(string), nil
	}

	// 开始调用接口获取token
	api := accessTokenURL
	if dd.SandBox {
		api = sandBoxTokenURL
	}
	reqAccessToken, err := GetTokenFromServer(api, dd.AppId, dd.AppSecret)
	if err != nil {
		return "", err
	}
	// 设置缓存
	expires := reqAccessToken.Data.ExpiresIn - 1500
	err = dd.Cache.Set(dd.GetCacheKey(), reqAccessToken.Data.AccessToken, time.Duration(expires)*time.Second)
	if err != nil {
		return "", err
	}
	return reqAccessToken.Data.AccessToken, nil
}

// ResAccessToken 获取token的返回结构体
type ResAccessToken struct {
	ErrNo   int                `json:"err_no,omitempty"`
	ErrTips string             `json:"err_tips,omitempty"`
	Data    ResAccessTokenData `json:"data,omitempty"`
}

type ResAccessTokenData struct {
	AccessToken string `json:"access_token,omitempty"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
}

// GetTokenFromServer 从抖音服务器获取token
func GetTokenFromServer(apiUrl string, appId, appSecret string) (resAccessToken ResAccessToken, err error) {
	params := map[string]interface{}{
		"appid":      appId,
		"secret":     appSecret,
		"grant_type": "client_credential",
	}
	body, err := util.PostJSON(apiUrl, params)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &resAccessToken)
	if err != nil {
		return
	}
	if resAccessToken.ErrNo != 0 {
		err = fmt.Errorf("get access_token error : errcode=%v , errormsg=%v", resAccessToken.ErrTips, resAccessToken.ErrNo)
		return
	}
	return
}
