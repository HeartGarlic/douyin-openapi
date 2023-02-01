package douyin_openapi

import (
	"fmt"
	accessToken "github.com/HeartGarlic/douyin-openapi/access-token"
	"github.com/HeartGarlic/douyin-openapi/cache"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

// 声明测试常量
const (
	AppId     = ""
	AppSecret = ""
	Token     = ""
	Salt      = ""
	NotifyUrl = ""
)

// 声明一个缓存实例
var Cache cache.Cache

// 声明全局openApi实例
var OpenApi *DouYinOpenApi

func init() {
	Cache = cache.NewMemory()
	OpenApi = NewDouYinOpenApi(DouYinOpenApiConfig{
		AppId:     AppId,
		AppSecret: AppSecret,
		IsSandbox: true,
		Token:     Token,
		Salt:      Salt,
	})
}

// 不足位数补零
func sup(i int64, n int) string {
	m := fmt.Sprintf("%d", i)
	for len(m) < n {
		m = fmt.Sprintf("0%s", m)
	}
	return m
}

var num int64

// GenerateSignOrderNo 生成单号
func GenerateSignOrderNo(prefix string) string {
	t := time.Now()
	s := t.Format("")
	m := t.UnixNano()/1e6 - t.UnixNano()/1e9*1e3
	ms := sup(m, 3)
	p := os.Getpid() % 1000
	ps := sup(int64(p), 3)
	i := atomic.AddInt64(&num, 1)
	r := i % 10000
	rs := sup(r, 4)
	n := fmt.Sprintf("%s%s%s%s%s", prefix, s, ms, ps, rs)
	return n
}

// 测试获取新的token
func TestDouyinOpenapi_NewDefaultAccessToken(t *testing.T) {
	token := accessToken.NewDefaultAccessToken(AppId, AppSecret, Cache, true)
	getAccessToken, err := token.GetAccessToken()
	if err != nil {
		t.Errorf("got a error: %s", err.Error())
		return
	}
	t.Logf("got a value: %s", getAccessToken)
}

// 基准测试看获取token的次数?
func BenchmarkDouyinOpenapi_NewDefaultAccessToken(b *testing.B) {
	token := accessToken.NewDefaultAccessToken(AppId, AppSecret, Cache, true)
	for i := 0; i < b.N; i++ {
		getAccessToken, err := token.GetAccessToken()
		b.Logf("get token: %s %+v", getAccessToken, err)
	}
}

// 测试小程序登录
func TestDouYinOpenApi_Code2Session(t *testing.T) {
	session, err := OpenApi.Code2Session("1111", "")
	if err != nil {
		t.Errorf("got a error %s", err.Error())
		return
	}
	t.Logf("got a value %+v", session)
}

// 测试下单接口
func TestDouYinOpenApi_CreateOrder(t *testing.T) {
	outOrderNo := GenerateSignOrderNo("")
	params := CreateOrderParams{
		OutOrderNo:      outOrderNo,
		TotalAmount:     1,
		Subject:         "爽豆充值",
		Body:            "爽豆充值",
		ValidTime:       300,
		CpExtra:         "123",
		NotifyUrl:       NotifyUrl,
		ExpandOrderInfo: ExpandOrderInfo{},
	}
	res, err := OpenApi.CreateOrder(params)
	if err != nil {
		t.Errorf("got a error %s", err.Error())
		return
	}
	t.Logf("got a value: %+v outOrderNo: %s", res, outOrderNo)
	return
}

func TestDouYinOpenApi_QueryOrder(t *testing.T) {
	gotQueryOrderResponse, err := OpenApi.QueryOrder("1934820001", "")
	if err != nil {
		t.Errorf("got a error %s", err.Error())
		return
	}
	t.Logf("got a value %+v", gotQueryOrderResponse)
}

func TestDouYinOpenApi_PayCallback(t *testing.T) {
	body := "{\n  \"timestamp\": \"1602507471\",\n  \"nonce\": \"797\",\n  \"msg\": \"{\\\"appid\\\":\\\"tt07e3715e98c9aac0\\\",\\\"cp_orderno\\\":\\\"out_order_no_1\\\",\\\"cp_extra\\\":\\\"\\\",\\\"way\\\":\\\"2\\\",\\\"payment_order_no\\\":\\\"2021070722001450071438803941\\\",\\\"total_amount\\\":9980,\\\"status\\\":\\\"SUCCESS\\\",\\\"seller_uid\\\":\\\"69631798443938962290\\\",\\\"extra\\\":\\\"null\\\",\\\"item_id\\\":\\\"\\\",\\\"order_id\\\":\\\"N71016888186626816\\\"}\",\n  \"msg_signature\": \"52fff5f7a4bf4a921c2daf83c75cf0e716432c73\",\n  \"type\": \"payment\"\n}"
	gotPayCallbackResponse, err := OpenApi.PayCallback(body, false)
	if err != nil {
		t.Errorf("got a error %s", err.Error())
		return
	}
	t.Logf("got a value %+v", gotPayCallbackResponse)
}

func TestDouYinOpenApi_CreateRefund(t *testing.T) {
	outRefundNo := GenerateSignOrderNo("") // 6887470001
	params := CreateRefundParams{
		OutOrderNo:   "1934820001",
		OutRefundNo:  outRefundNo,
		Reason:       "发起退款",
		RefundAmount: 1,
		NotifyUrl:    NotifyUrl,
	}
	gotCreateRefundResponse, err := OpenApi.CreateRefund(params)
	if err != nil {
		t.Errorf("CreateRefund() error = %v outRefundNo: %s", err, outRefundNo)
		return
	}
	t.Logf("got a value %+v", gotCreateRefundResponse)
	return
}
