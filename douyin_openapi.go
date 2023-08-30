package douyin_openapi

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	accessToken "github.com/HeartGarlic/douyin-openapi/access-token"
	"github.com/HeartGarlic/douyin-openapi/cache"
	"github.com/HeartGarlic/douyin-openapi/util"
	"sort"
	"strings"
)

const (
	code2Session         = "/api/apps/v2/jscode2session"                                              // 小程序登录地址
	createOrder          = "https://developer.toutiao.com/api/apps/ecpay/v1/create_order"             // 预下单
	queryOrder           = "https://developer.toutiao.com/api/apps/ecpay/v1/query_order"              // 订单查询
	createRefund         = "https://developer.toutiao.com/api/apps/ecpay/v1/create_refund"            // 退款
	queryRefund          = "https://developer.toutiao.com/api/apps/ecpay/v1/query_refund"             // 退款结果查询
	settle               = "https://developer.toutiao.com/api/apps/ecpay/v1/settle"                   // 结算
	querySettle          = "https://developer.toutiao.com/api/apps/ecpay/v1/query_settle"             // 结算结果查询
	unsettleAmount       = "https://developer.toutiao.com/api/apps/ecpay/v1/unsettle_amount"          // 可结算金额查询
	createReturn         = "https://developer.toutiao.com/api/apps/ecpay/v1/create_return"            // 退分账
	queryReturn          = "https://developer.toutiao.com/api/apps/ecpay/v1/query_return"             // 退分账结果查询
	queryMerchantBalance = "https://developer.toutiao.com/api/apps/ecpay/saas/query_merchant_balance" // 商户余额查询
	merchantWithdraw     = "https://developer.toutiao.com/api/apps/ecpay/saas/merchant_withdraw"      // 商户提现
	queryWithdrawOrder   = "https://developer.toutiao.com/api/apps/ecpay/saas/query_withdraw_order"   // 提现结果查询
	orderV2Push          = "https://developer.toutiao.com/api/apps/order/v2/push"                     // 订单推送
	imAuthorizeSendMsg   = "https://open.douyin.com/im/authorize/send/msg/"                           // 主动发送私信
	imAuthorizeRecallMsg = "https://open.douyin.com/im/authorize/recall/msg/"                         // 撤回私信
	urlLinkGenerate      = "https://developer.toutiao.com/api/apps/url_link/generate"                 // 生成Link
	imGroupFansList      = "https://open.douyin.com/im/group/fans/list/"                              // 查询群信息
)

// DouYinOpenApiConfig 实例化配置
type DouYinOpenApiConfig struct {
	AppId       string
	AppSecret   string
	AccessToken accessToken.AccessToken
	Cache       cache.Cache
	IsSandbox   bool
	Token       string
	Salt        string
}

// DouYinOpenApi 基类
type DouYinOpenApi struct {
	Config  DouYinOpenApiConfig
	BaseApi string
}

// NewDouYinOpenApi 实例化一个抖音openapi实例
func NewDouYinOpenApi(config DouYinOpenApiConfig) *DouYinOpenApi {
	if config.Cache == nil {
		config.Cache = cache.NewMemory()
	}
	if config.AccessToken == nil {
		config.AccessToken = accessToken.NewDefaultAccessToken(config.AppId, config.AppSecret, config.Cache, config.IsSandbox)
	}
	BaseApi := "https://developer.toutiao.com"
	if config.IsSandbox {
		BaseApi = "https://open-sandbox.douyin.com"
	}
	return &DouYinOpenApi{
		Config:  config,
		BaseApi: BaseApi,
	}
}

// GetApiUrl 获取api地址
func (d *DouYinOpenApi) GetApiUrl(url string) string {
	return fmt.Sprintf("%s%s", d.BaseApi, url)
}

// PostJson 封装公共的请求方法
func (d *DouYinOpenApi) PostJson(api string, params interface{}, response interface{}) (err error) {
	body, err := util.PostJSON(api, params)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return
	}
	return
}

// Code2SessionParams 小程序登录 所需参数
type Code2SessionParams struct {
	Appid         string `json:"appid,omitempty"`
	Secret        string `json:"secret,omitempty"`
	AnonymousCode string `json:"anonymous_code,omitempty"`
	Code          string `json:"code,omitempty"`
}

// Code2SessionResponse 小程序登录返回值
type Code2SessionResponse struct {
	ErrNo   int                      `json:"err_no,omitempty"`
	ErrTips string                   `json:"err_tips,omitempty"`
	Data    Code2SessionResponseData `json:"data,omitempty"`
}

type Code2SessionResponseData struct {
	SessionKey      string `json:"session_key,omitempty"`
	Openid          string `json:"openid,omitempty"`
	AnonymousOpenid string `json:"anonymous_openid,omitempty"`
	UnionId         string `json:"unionid,omitempty"`
}

// Code2Session 小程序登录
func (d *DouYinOpenApi) Code2Session(code, anonymousCode string) (code2SessionResponse Code2SessionResponse, err error) {
	params := Code2SessionParams{
		Appid:         d.Config.AppId,
		Secret:        d.Config.AppSecret,
		AnonymousCode: anonymousCode,
		Code:          code,
	}
	err = d.PostJson(d.GetApiUrl(code2Session), params, &code2SessionResponse)
	if err != nil {
		return
	}
	if code2SessionResponse.ErrNo != 0 {
		return code2SessionResponse, fmt.Errorf("小程序登录错误: %s %d", code2SessionResponse.ErrTips, code2SessionResponse.ErrNo)
	}
	return
}

// CreateOrderParams 预下单接口参数
type CreateOrderParams struct {
	AppId           string          `json:"app_id,omitempty"`            // app_id string 是 64 小程序APPID tt07e3715e98c9aac0
	OutOrderNo      string          `json:"out_order_no,omitempty"`      // out_order_no string 是 64 开发者侧的订单号。 只能是数字、大小写字母_-*且在同一个app_id下唯一 7056505317450041644
	TotalAmount     int64           `json:"total_amount,omitempty"`      // total_amount number 是 取值范围： [1,10000000000] 支付价格。 单位为[分] 100，即1元
	Subject         string          `json:"subject,omitempty"`           // subject string 是 128 商品描述。 长度限制不超过 128 字节且不超过 42 字符 抖音商品XYZ
	Body            string          `json:"body,omitempty"`              // body string 是 128 商品详情 长度限制不超过 128 字节且不超过 42 字符 抖音商品XYZ
	ValidTime       int64           `json:"valid_time,omitempty"`        // valid_time number 是 取值范围： [300,172800] 订单过期时间(秒)。最小5分钟，最大2天，小于5分钟会被置为5分钟，大于2天会被置为2天 900，即15分钟
	Sign            string          `json:"sign,omitempty"`              // sign string 是 344 签名，详见签名DEMO 21fc77aeeaad725d9500062a888888a2a3d
	CpExtra         string          `json:"cp_extra,omitempty"`          // cp_extra string 否 2048 开发者自定义字段，回调原样回传。 超过最大长度会被截断 502205261403349
	NotifyUrl       string          `json:"notify_url,omitempty"`        // notify_url string 否 256 商户自定义回调地址，必须以 https 开头，支持 443 端口。 指定时，支付成功后抖音会请求该地址通知开发者 https://api.iiyyeixin.com/Notify/bytedancePay
	ThirdpartyId    string          `json:"thirdparty_id,omitempty"`     // thirdparty_id 条件选填 服务商模式接入必传 64 第三方平台服务商 id，非服务商模式留空 tt84a4f2177777e29df
	StoreUid        string          `json:"store_uid,omitempty"`         // store_uid string 条件选填 多门店模式下可传 64 可用此字段指定本单使用的收款商户号（目前为灰度功能，需要联系平台运营添加白名单，白名单添加1小时后生效；未在白名单的小程序，该字段不生效） 70084531288883795888
	DisableMsg      int             `json:"disable_msg,omitempty"`       // disable_msg number 否 是否屏蔽支付完成后推送用户抖音消息，1-屏蔽 0-非屏蔽，默认为0。 特别注意： 若接入POI, 请传1。因为POI订单体系会发消息，所以不用再接收一次担保支付推送消息， 1
	MsgPage         string          `json:"msg_page,omitempty"`          // msg_page string 否 支付完成后推送给用户的抖音消息跳转页面，开发者需要传入在app.json中定义的链接，如果不传则跳转首页。 pages/orderDetail/orderDetail?no = DYMP8218048851499944448\u0026id = 797775
	ExpandOrderInfo ExpandOrderInfo `json:"expand_order_info,omitempty"` // expand_order_info 否 - 订单拓展信息，详见下面 expand_order_info参数说明 { "original_delivery_fee":10, "actual_delivery_fee":10 }
	LimitPayWay     string          `json:"limit_pay_way,omitempty"`     // limit_pay_way string 否 64 屏蔽指定支付方式，屏蔽多个支付方式，请使用逗号","分割，枚举值： 屏蔽微信支付：LIMIT_WX 屏蔽支付宝支付：LIMIT_ALI 屏蔽抖音支付：LIMIT_DYZF 特殊说明：若之前开通了白名单，平台会保留之前屏蔽逻辑；若传入该参数，会优先以传入的为准，白名单则无效 屏蔽抖音支付和微信支付： "LIMIT_DYZF,LIMIT_WX"
}

type ExpandOrderInfo struct {
	OriginalDeliveryFee int
	ActualDeliveryFee   int
}

// CreateOrderResponse 预下单返回值
type CreateOrderResponse struct {
	ErrNo   int                     `json:"err_no,omitempty"`
	ErrTips string                  `json:"err_tips,omitempty"`
	Data    CreateOrderResponseData `json:"data,omitempty"`
}

type CreateOrderResponseData struct {
	OrderId    string `json:"order_id,omitempty"`
	OrderToken string `json:"order_token,omitempty"`
}

// CreateOrder 预下单
func (d *DouYinOpenApi) CreateOrder(params CreateOrderParams) (createOrderResponse CreateOrderResponse, err error) {
	params.AppId = d.Config.AppId
	params.Sign = d.GenerateSign(params)
	err = d.PostJson(createOrder, params, &createOrderResponse)
	if err != nil {
		return
	}
	if createOrderResponse.ErrNo != 0 {
		return createOrderResponse, fmt.Errorf("%s %d", createOrderResponse.ErrTips, createOrderResponse.ErrNo)
	}
	return
}

// GenerateSign 生成请求签名
func (d *DouYinOpenApi) GenerateSign(params interface{}) string {
	var paramsMap map[string]interface{}
	var paramsArr []string
	j, _ := json.Marshal(&params)
	err := json.Unmarshal(j, &paramsMap)
	if err != nil {
		return ""
	}
	for k, v := range paramsMap {
		if k == "other_settle_params" || k == "app_id" || k == "thirdparty_id" || k == "sign" || k == "salt" || k == "token" {
			continue
		}
		value := strings.TrimSpace(fmt.Sprintf("%v", v))
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") && len(value) > 1 {
			value = value[1 : len(value)-1]
		}
		value = strings.TrimSpace(value)
		if value == "" || value == "null" {
			continue
		}
		paramsArr = append(paramsArr, value)
	}
	paramsArr = append(paramsArr, d.Config.Salt)
	sort.Strings(paramsArr)
	return fmt.Sprintf("%x", md5.Sum([]byte(strings.Join(paramsArr, "&"))))
}

// QueryOrderParams 订单查询接口参数
type QueryOrderParams struct {
	AppId        string `json:"app_id,omitempty"`
	OutOrderNo   string `json:"out_order_no,omitempty"`
	Sign         string `json:"sign,omitempty"`
	ThirdpartyId string `json:"thirdparty_id,omitempty"`
}

type QueryOrderResponse struct {
	ErrNo       int         `json:"err_no,omitempty"`
	ErrTips     string      `json:"err_tips,omitempty"`
	OutOrderNo  string      `json:"out_order_no,omitempty"`
	OrderId     string      `json:"order_id,omitempty"`
	PaymentInfo PaymentInfo `json:"payment_info,omitempty"`
}

type PaymentInfo struct {
	TotalFee    int    `json:"total_fee,omitempty"`
	OrderStatus string `json:"order_status,omitempty"` // SUCCESS：成功 TIMEOUT：超时未支付 PROCESSING：处理中 FAIL：失败
	PayTime     string `json:"pay_time,omitempty"`     // 支付时间， 格式为"yyyy-MM-dd hh:mm:ss"
	Way         int    `json:"way,omitempty"`          // 支付渠道， 1-微信支付，2-支付宝支付，10-抖音支付
	ChannelNo   string `json:"channel_no,omitempty"`
	SellerUid   string `json:"seller_uid,omitempty"`
	ItemId      string `json:"item_id,omitempty"`
	CpsInfo     string `json:"cps_info,omitempty"`
}

// QueryOrder 支付结果查询
func (d *DouYinOpenApi) QueryOrder(outOrderNo, thirdpartyId string) (queryOrderResponse QueryOrderResponse, err error) {
	queryParams := QueryOrderParams{
		AppId:        d.Config.AppId,
		OutOrderNo:   outOrderNo,
		ThirdpartyId: thirdpartyId,
	}
	queryParams.Sign = d.GenerateSign(queryParams)
	err = d.PostJson(queryOrder, queryParams, &queryOrderResponse)
	if err != nil {
		return
	}
	if queryOrderResponse.ErrNo != 0 {
		return queryOrderResponse, fmt.Errorf("%s %d", queryOrderResponse.ErrTips, queryOrderResponse.ErrNo)
	}
	return
}

// PayCallbackResponse 支付回调结构体
type PayCallbackResponse struct {
	Timestamp    string                  `json:"timestamp,omitempty"`
	Nonce        string                  `json:"nonce,omitempty"`
	Msg          string                  `json:"msg,omitempty"`
	MsgStruct    PayCallbackResponseData `json:"msg_struct"`
	MsgSignature string                  `json:"msg_signature,omitempty"`
	Type         string                  `json:"type,omitempty"`
}

type PayCallbackResponseData struct {
	Appid          string `json:"appid,omitempty"`
	CpOrderNo      string `json:"cp_orderno,omitempty"`
	CpExtra        string `json:"cp_extra,omitempty"`
	Way            string `json:"way,omitempty"`
	PaymentOrderNo string `json:"payment_order_no,omitempty"`
	TotalAmount    int    `json:"total_amount,omitempty"`
	Status         string `json:"status,omitempty"`
	SellerUid      string `json:"seller_uid,omitempty"`
	Extra          string `json:"extra,omitempty"`
	ItemId         string `json:"item_id,omitempty"`
	OrderId        string `json:"order_id,omitempty"`
}

// CheckResponseSign 校验回调签名
func (d *DouYinOpenApi) CheckResponseSign(oldSign string, strArr []string) error {
	sort.Strings(strArr)
	h := sha1.New()
	h.Write([]byte(strings.Join(strArr, "")))
	newSign := fmt.Sprintf("%x", h.Sum(nil))
	if newSign != oldSign {
		return fmt.Errorf("回调验签失败 newSign:%s oldSign:%s", newSign, oldSign)
	}
	return nil
}

// PayCallback 支付结果回调
func (d *DouYinOpenApi) PayCallback(body string, checkSign bool) (payCallbackResponse PayCallbackResponse, err error) {
	// 解析数据
	err = json.Unmarshal([]byte(body), &payCallbackResponse)
	if err != nil {
		return
	}
	// 判断是否需要校验签名
	if checkSign {
		sortedString := make([]string, 0)
		sortedString = append(sortedString, d.Config.Token)
		sortedString = append(sortedString, payCallbackResponse.Timestamp)
		sortedString = append(sortedString, payCallbackResponse.Nonce)
		sortedString = append(sortedString, payCallbackResponse.Msg)
		err = d.CheckResponseSign(payCallbackResponse.MsgSignature, sortedString)
		if err != nil {
			return
		}
	}
	var msgStruct PayCallbackResponseData
	// 解析 msg 数据到结构体
	err = json.Unmarshal([]byte(payCallbackResponse.Msg), &msgStruct)
	if err != nil {
		return
	}
	payCallbackResponse.MsgStruct = msgStruct
	return
}

// CreateRefundParams 发起退款参数
type CreateRefundParams struct {
	AppId        string `json:"app_id,omitempty"`
	OutOrderNo   string `json:"out_order_no,omitempty"`
	OutRefundNo  string `json:"out_refund_no,omitempty"`
	Reason       string `json:"reason,omitempty"`
	RefundAmount int    `json:"refund_amount,omitempty"`
	Sign         string `json:"sign,omitempty"`
	CpExtra      string `json:"cp_extra,omitempty"`
	NotifyUrl    string `json:"notify_url,omitempty"`
	ThirdpartyId string `json:"thirdparty_id,omitempty"`
	DisableMsg   int    `json:"disable_msg,omitempty"`
	MsgPage      string `json:"msg_page,omitempty"`
}

type CreateRefundResponse struct {
	ErrNo    int    `json:"err_no,omitempty"`
	ErrTips  string `json:"err_tips,omitempty"`
	RefundNo string `json:"refund_no,omitempty"`
}

// CreateRefund 发起退款
func (d *DouYinOpenApi) CreateRefund(params CreateRefundParams) (createRefundResponse CreateRefundResponse, err error) {
	params.AppId = d.Config.AppId
	params.Sign = d.GenerateSign(params)
	err = d.PostJson(createRefund, params, &createRefundResponse)
	if err != nil {
		return
	}
	if createRefundResponse.ErrNo != 0 {
		return createRefundResponse, fmt.Errorf("CreateRefund error %s %d", createRefundResponse.ErrTips, createRefundResponse.ErrNo)
	}
	return
}

// QueryRefundParams 查询退款参数
type QueryRefundParams struct {
	OutRefundNo  string `json:"out_refund_no,omitempty"`
	AppId        string `json:"app_id,omitempty"`
	Sign         string `json:"sign,omitempty"`
	ThirdpartyId string `json:"thirdparty_id,omitempty"`
}

type QueryRefundParamsResponse struct {
	ErrNo      int    `json:"err_no,omitempty"`
	ErrTips    string `json:"err_tips,omitempty"`
	RefundInfo struct {
		RefundNo     string `json:"refund_no,omitempty"`
		RefundAmount int    `json:"refund_amount,omitempty"`
		RefundStatus string `json:"refund_status,omitempty"`
		RefundedAt   int    `json:"refunded_at,omitempty"`
		IsAllSettled bool   `json:"is_all_settled,omitempty"`
		CpExtra      string `json:"cp_extra,omitempty"`
		Msg          string `json:"msg,omitempty"`
	} `json:"refundInfo"`
}

// QueryRefund 退款结果查询
func (d *DouYinOpenApi) QueryRefund(outRefundNo, thirdpartyId string) (queryRefundParamsResponse QueryRefundParamsResponse, err error) {
	params := QueryRefundParams{
		OutRefundNo:  outRefundNo,
		AppId:        d.Config.AppId,
		ThirdpartyId: thirdpartyId,
	}
	params.Sign = d.GenerateSign(params)
	err = d.PostJson(queryRefund, params, &queryRefundParamsResponse)
	if err != nil {
		return
	}
	if queryRefundParamsResponse.ErrNo != 0 {
		return queryRefundParamsResponse, fmt.Errorf("QueryRefund error %s %d", queryRefundParamsResponse.ErrTips, queryRefundParamsResponse.ErrNo)
	}
	return
}

// RefundCallbackResponse 退款回调结果
type RefundCallbackResponse struct {
	Timestamp    string `json:"timestamp"`
	Nonce        string `json:"nonce"`
	Msg          string `json:"msg"`
	MsgStruct    RefundCallbackResponseMsg
	MsgSignature string `json:"msg_signature"`
	Type         string `json:"type"`
}

type RefundCallbackResponseMsg struct {
	Appid        string `json:"appid"`
	CpRefundNo   string `json:"cp_refundno"`
	CpExtra      string `json:"cp_extra"`
	Status       string `json:"status"`
	RefundAmount int    `json:"refund_amount"`
	IsAllSettled bool   `json:"is_all_settled"`
	RefundedAt   int    `json:"refunded_at"`
	Message      string `json:"message"`
	OrderId      string `json:"order_id"`
	RefundNo     string `json:"refund_no"`
}

// RefundCallback 退款结果回调
func (d *DouYinOpenApi) RefundCallback(body string, checkSign bool) (refundCallbackResponse RefundCallbackResponse, err error) {
	err = json.Unmarshal([]byte(body), &refundCallbackResponse)
	if err != nil {
		return
	}
	// 判断是否需要校验签名
	if checkSign {
		sortedString := make([]string, 0)
		sortedString = append(sortedString, d.Config.Token)
		sortedString = append(sortedString, refundCallbackResponse.Timestamp)
		sortedString = append(sortedString, refundCallbackResponse.Nonce)
		sortedString = append(sortedString, refundCallbackResponse.Msg)
		err = d.CheckResponseSign(refundCallbackResponse.MsgSignature, sortedString)
		if err != nil {
			return
		}
	}
	var msgStruct RefundCallbackResponseMsg
	// 解析 msg 数据到结构体
	err = json.Unmarshal([]byte(refundCallbackResponse.Msg), &msgStruct)
	if err != nil {
		return
	}
	refundCallbackResponse.MsgStruct = msgStruct
	return
}

// SettleParams 发起分账参数
type SettleParams struct {
	OutSettleNo  string `json:"out_settle_no,omitempty"`
	OutOrderNo   string `json:"out_order_no,omitempty"`
	SettleDesc   string `json:"settle_desc,omitempty"`
	NotifyUrl    string `json:"notify_url,omitempty"`
	CpExtra      string `json:"cp_extra,omitempty"`
	AppId        string `json:"app_id,omitempty"`
	Sign         string `json:"sign,omitempty"`
	SettleParams string `json:"settle_params,omitempty"`
	ThirdpartyId string `json:"thirdparty_id,omitempty"`
	Finish       string `json:"finish,omitempty"`
}

// SettleParamsItem 分账方参数
type SettleParamsItem struct {
	MerchantUid string `json:"merchant_uid,omitempty"`
	Amount      int    `json:"amount,omitempty"`
}

// SettleResponse 分账结果
type SettleResponse struct {
	ErrNo    int    `json:"err_no,omitempty"`
	ErrTips  string `json:"err_tips,omitempty"`
	SettleNo string `json:"settle_no,omitempty"`
}

// Settle 发起结算及分账
func (d *DouYinOpenApi) Settle(settleParams SettleParams, settleParamsItem ...SettleParamsItem) (settleResponse SettleResponse, err error) {
	settleParams.AppId = d.Config.AppId
	settleItem, _ := json.Marshal(settleParamsItem)
	settleParams.SettleParams = string(settleItem)
	settleParams.Sign = d.GenerateSign(settleParams)
	err = d.PostJson(settle, settleParams, &settleResponse)
	if err != nil {
		return
	}
	if settleResponse.ErrNo != 0 {
		err = fmt.Errorf("settle error %s %d", settleResponse.ErrTips, settleResponse.ErrNo)
		return
	}
	return
}

// QuerySettleParams 结算结果查询参数
type QuerySettleParams struct {
	Sign         string `json:"sign,omitempty"`
	AppId        string `json:"app_id,omitempty"`
	OutSettleNo  string `json:"out_settle_no,omitempty"`
	ThirdpartyId string `json:"thirdparty_id,omitempty"`
}

// QuerySettleResponse 结算结果返回值
type QuerySettleResponse struct {
	ErrNo      int    `json:"err_no"`
	ErrTips    string `json:"err_tips"`
	SettleInfo struct {
		SettleNo     string `json:"settle_no"`
		SettleAmount int    `json:"settle_amount"`
		SettleStatus string `json:"settle_status"`
		SettleDetail string `json:"settle_detail"`
		SettledAt    int    `json:"settled_at"`
		Rake         int    `json:"rake"`
		Commission   int    `json:"commission"`
		CpExtra      string `json:"cp_extra"`
		Msg          string `json:"msg"`
	} `json:"settle_info"`
}

// QuerySettle 结算结果查询 querySettle
func (d *DouYinOpenApi) QuerySettle(outSettleNo, thirdpartyId string) (querySettleResponse QuerySettleResponse, err error) {
	params := QuerySettleParams{
		AppId:        d.Config.AppId,
		OutSettleNo:  outSettleNo,
		ThirdpartyId: thirdpartyId,
	}
	params.Sign = d.GenerateSign(params)
	err = d.PostJson(querySettle, params, &querySettleResponse)
	if err != nil {
		return
	}
	if querySettleResponse.ErrNo != 0 {
		err = fmt.Errorf("QuerySettle error %s %d", querySettleResponse.ErrTips, querySettleResponse.ErrNo)
		return
	}
	return
}

type SettleCallbackResponse struct {
	Timestamp    string `json:"timestamp"`
	Nonce        string `json:"nonce"`
	Type         string `json:"type"`
	Msg          string `json:"msg"`
	MsgStruct    SettleCallbackResponseMsg
	MsgSignature string `json:"msg_signature"`
}

type SettleCallbackResponseMsg struct {
	AppId           string `json:"app_id"`
	CpSettleNo      string `json:"cp_settle_no"`
	CpExtra         string `json:"cp_extra"`
	Status          string `json:"status"`
	Rake            int    `json:"rake"`
	Commission      int    `json:"commission"`
	SettleDetail    string `json:"settle_detail"`
	SettledAt       int    `json:"settled_at"`
	Message         string `json:"message"`
	OrderId         string `json:"order_id"`
	ChannelSettleId string `json:"channel_settle_id"`
	SettleAmount    int    `json:"settle_amount"`
	SettleNo        string `json:"settle_no"`
	OutOrderNo      string `json:"out_order_no"`
	IsAutoSettle    bool   `json:"is_auto_settle"`
}

// SettleCallback 结算结果回调
func (d *DouYinOpenApi) SettleCallback(body string, checkSign bool) (settleCallbackResponse SettleCallbackResponse, err error) {
	err = json.Unmarshal([]byte(body), &settleCallbackResponse)
	if err != nil {
		return
	}
	// 判断是否需要校验签名
	if checkSign {
		sortedString := make([]string, 0)
		sortedString = append(sortedString, d.Config.Token)
		sortedString = append(sortedString, settleCallbackResponse.Timestamp)
		sortedString = append(sortedString, settleCallbackResponse.Nonce)
		sortedString = append(sortedString, settleCallbackResponse.Msg)
		err = d.CheckResponseSign(settleCallbackResponse.MsgSignature, sortedString)
		if err != nil {
			return
		}
	}
	var msgStruct SettleCallbackResponseMsg
	// 解析 msg 数据到结构体
	err = json.Unmarshal([]byte(settleCallbackResponse.Msg), &msgStruct)
	if err != nil {
		return
	}
	settleCallbackResponse.MsgStruct = msgStruct
	return
}

// UnsettleAmountParams 可分账余额查询
type UnsettleAmountParams struct {
	OutOrderNo     string `json:"out_order_no,omitempty"`
	AppId          string `json:"app_id,omitempty"`
	Sign           string `json:"sign,omitempty"`
	ThirdpartyId   string `json:"thirdparty_id,omitempty"`
	OutItemOrderNo string `json:"out_item_order_no,omitempty"`
}

type UnsettleAmountResponse struct {
	ErrNo   int    `json:"err_no"`
	ErrTips string `json:"err_tips"`
	Data    struct {
		OutOrderNo     string `json:"out_order_no"`
		UnsettleAmount int    `json:"unsettle_amount"`
		Detail         struct {
			PayInfo struct {
				OutOrderNo string `json:"out_order_no"`
				Amount     int    `json:"amount"`
			} `json:"pay_info"`
			RefundInfo []struct {
				OutRefundNo string `json:"out_refund_no"`
				Amount      int    `json:"amount"`
			} `json:"refund_info"`
			PaymentRake int `json:"payment_rake"`
			LifeRake    int `json:"life_rake"`
			Commission  int `json:"commission"`
		} `json:"detail"`
	} `json:"data"`
}

// UnsettleAmount 可分账余额查询 unsettleAmount
func (d *DouYinOpenApi) UnsettleAmount(outOrderNo, thirdpartyId, outItemOrderNo string) (unsettleAmountResponse UnsettleAmountResponse, err error) {
	params := UnsettleAmountParams{
		OutOrderNo:     outOrderNo,
		AppId:          d.Config.AppId,
		ThirdpartyId:   thirdpartyId,
		OutItemOrderNo: outItemOrderNo,
	}
	params.Sign = d.GenerateSign(params)
	err = d.PostJson(unsettleAmount, params, &unsettleAmountResponse)
	if err != nil {
		return
	}
	if unsettleAmountResponse.ErrNo != 0 {
		err = fmt.Errorf("UnsettleAmount error %s %d", unsettleAmountResponse.ErrTips, unsettleAmountResponse.ErrNo)
		return
	}
	return
}

// CreateReturnParams 退分账参数
type CreateReturnParams struct {
	AppId        string `json:"app_id,omitempty"`
	ThirdpartyId string `json:"thirdparty_id,omitempty"`
	OutSettleNo  string `json:"out_settle_no,omitempty"`
	SettleNo     string `json:"settle_no,omitempty"`
	OutReturnNo  string `json:"out_return_no,omitempty"`
	MerchantUid  string `json:"merchant_uid,omitempty"`
	ReturnAmount int    `json:"return_amount,omitempty"`
	ReturnDesc   string `json:"return_desc,omitempty"`
	CpExtra      string `json:"cp_extra,omitempty"`
	Sign         string `json:"sign,omitempty"`
}

type CreateReturnResponse struct {
	ErrNo      int    `json:"err_no"`
	ErrTips    string `json:"err_tips"`
	ReturnInfo struct {
		AppId        string `json:"app_id"`
		ThirdpartyId string `json:"thirdparty_id"`
		SettleNo     string `json:"settle_no"`
		OutSettleNo  string `json:"out_settle_no"`
		OutReturnNo  string `json:"out_return_no"`
		MerchantUid  string `json:"merchant_uid"`
		ReturnAmount int    `json:"return_amount"`
		ReturnStatus string `json:"return_status"`
		ReturnNo     string `json:"return_no"`
		FailReason   string `json:"fail_reason"`
		FinishTime   int    `json:"finish_time"`
		CpExtra      string `json:"cp_extra"`
	} `json:"return_info"`
}

// CreateReturn 退分账 createReturn
func (d *DouYinOpenApi) CreateReturn(params CreateReturnParams) (createReturnResponse CreateReturnResponse, err error) {
	params.AppId = d.Config.AppId
	params.Sign = d.GenerateSign(params)
	err = d.PostJson(createReturn, params, &createReturnResponse)
	if err != nil {
		return
	}
	if createReturnResponse.ErrNo != 0 {
		err = fmt.Errorf("CreateReturn error %s %d", createReturnResponse.ErrTips, createReturnResponse.ErrNo)
		return
	}
	return
}

// QueryReturnParams 退分账结果查询 参数
type QueryReturnParams struct {
	AppId        string `json:"app_id,omitempty"`
	ReturnNo     string `json:"return_no,omitempty"`
	OutReturnNo  string `json:"out_return_no,omitempty"`
	Sign         string `json:"sign,omitempty"`
	ThirdpartyId string `json:"thirdparty_id,omitempty"`
}

// QueryReturnResponse 退分账结果查询
type QueryReturnResponse struct {
	ErrNo      int    `json:"err_no"`
	ErrTips    string `json:"err_tips"`
	ReturnInfo struct {
		AppId        string `json:"app_id"`
		ThirdpartyId string `json:"thirdparty_id"`
		SettleNo     string `json:"settle_no"`
		OutSettleNo  string `json:"out_settle_no"`
		OutReturnNo  string `json:"out_return_no"`
		MerchantUid  string `json:"merchant_uid"`
		ReturnAmount int    `json:"return_amount"`
		ReturnStatus string `json:"return_status"`
		ReturnNo     string `json:"return_no"`
		FailReason   string `json:"fail_reason"`
		FinishTime   int    `json:"finish_time"`
		CpExtra      string `json:"cp_extra"`
	} `json:"return_info"`
}

// QueryReturn 退分账结果查询 queryReturn
func (d *DouYinOpenApi) QueryReturn(returnNo, outReturnNo, thirdpartyId string) (queryReturnResponse QueryReturnResponse, err error) {
	params := QueryReturnParams{
		AppId:        d.Config.AppId,
		ReturnNo:     returnNo,
		OutReturnNo:  outReturnNo,
		ThirdpartyId: thirdpartyId,
	}
	params.Sign = d.GenerateSign(params)
	err = d.PostJson(queryReturn, params, &queryReturnResponse)
	if err != nil {
		return
	}
	if queryReturnResponse.ErrNo != 0 {
		err = fmt.Errorf("queryReturnResponse error %s %d", queryReturnResponse.ErrTips, queryReturnResponse.ErrNo)
		return
	}
	return
}

// QueryMerchantBalanceParams 可提现余额查询
type QueryMerchantBalanceParams struct {
	ThirdpartyId   string `json:"thirdparty_id,omitempty"`
	AppId          string `json:"app_id,omitempty"`
	MerchantUid    string `json:"merchant_uid,omitempty"`
	ChannelType    string `json:"channel_type,omitempty"` // alipay: 支付宝 wx: 微信	hz: 抖音支付
	Sign           string `json:"sign,omitempty"`
	MerchantEntity string `json:"merchant_entity"`
}

type QueryMerchantBalanceResponse struct {
	ErrNo       int    `json:"err_no"`
	ErrTips     string `json:"err_tips"`
	AccountInfo struct {
		OnlineBalance       int `json:"online_balance"`
		WithDrawableBalance int `json:"withdrawable_balacne"`
		FreezeBalance       int `json:"freeze_balance"`
	} `json:"account_info"`
	SettleInfo struct {
		SettleType    int    `json:"settle_type"`
		SettleAccount string `json:"settle_account"`
		BankcardNo    string `json:"bankcard_no"`
		BankName      string `json:"bank_name"`
	} `json:"settle_info"`
	MerchantEntity int `json:"merchant_entity"`
}

// QueryMerchantBalance 可提现余额查询
func (d *DouYinOpenApi) QueryMerchantBalance(params QueryMerchantBalanceParams) (queryMerchantBalanceResponse QueryMerchantBalanceResponse, err error) {
	params.AppId = d.Config.AppId
	params.Sign = d.GenerateSign(params)
	err = d.PostJson(queryMerchantBalance, params, &queryMerchantBalanceResponse)
	if err != nil {
		return
	}
	if queryMerchantBalanceResponse.ErrNo != 0 {
		err = fmt.Errorf("queryMerchantBalanceResponse error %s %d", queryMerchantBalanceResponse.ErrTips, queryMerchantBalanceResponse.ErrNo)
		return
	}
	return
}

// MerchantWithdrawParams 商户提现参数
type MerchantWithdrawParams struct {
	ThirdpartyId   string `json:"thirdparty_id,omitempty"`
	AppId          string `json:"app_id,omitempty"`
	MerchantUid    string `json:"merchant_uid,omitempty"`
	ChannelType    string `json:"channel_type,omitempty"` // alipay: 支付宝 wx: 微信  hz: 抖音支付 yeepay: 易宝
	WithdrawAmount int    `json:"withdraw_amount,omitempty"`
	OutOrderId     string `json:"out_order_id,omitempty"`
	Sign           string `json:"sign,omitempty"`
	Callback       string `json:"callback,omitempty"`
	CpExtra        string `json:"cp_extra,omitempty"`
	MerchantEntity int    `json:"merchant_entity,omitempty"`
}

type MerchantWithdrawResponse struct {
	ErrNo          int    `json:"err_no"`
	ErrTips        string `json:"err_tips"`
	OrderId        string `json:"order_id"`
	MerchantEntity int    `json:"merchant_entity"`
}

// MerchantWithdraw 提现
func (d *DouYinOpenApi) MerchantWithdraw(params MerchantWithdrawParams) (merchantWithdrawResponse MerchantWithdrawResponse, err error) {
	params.AppId = d.Config.AppId
	params.Sign = d.GenerateSign(params)
	err = d.PostJson(merchantWithdraw, params, &merchantWithdrawResponse)
	if err != nil {
		return
	}
	if merchantWithdrawResponse.ErrNo != 0 {
		err = fmt.Errorf("merchantWithdrawResponse error %s %d", merchantWithdrawResponse.ErrTips, merchantWithdrawResponse.ErrNo)
		return
	}
	return
}

// QueryWithdrawOrderParams 提现结果查询
type QueryWithdrawOrderParams struct {
	ThirdpartyId string `json:"thirdparty_id,omitempty"`
	AppId        string `json:"app_id,omitempty"`
	MerchantUid  string `json:"merchant_uid,omitempty"`
	ChannelType  string `json:"channel_type,omitempty"`
	OutOrderId   string `json:"out_order_id,omitempty"`
	Sign         string `json:"sign,omitempty"`
}

type QueryWithdrawOrderResponse struct {
	ErrNo     int    `json:"err_no"`
	ErrTips   string `json:"err_tips"`
	Status    string `json:"status"` // 状态枚举值: 成功:SUCCESS 失败: FAIL 处理中: PROCESSING 退票: REEXCHANGE 注： 退票：商户的提现申请请求通过渠道（微信/支付宝/抖音支付）提交给银行处理后，银行返回结果是处理成功，渠道返回给商户提现成功，但间隔一段时间后，银行再次通知渠道处理失败并返还款项给渠道，渠道再将该笔失败款返还至商户在渠道的账户余额中
	StatusMsg string `json:"statusMsg"`
}

// QueryWithdrawOrder 提现结果查询
func (d *DouYinOpenApi) QueryWithdrawOrder(params QueryWithdrawOrderParams) (queryWithdrawOrderResponse QueryWithdrawOrderResponse, err error) {
	params.AppId = d.Config.AppId
	params.Sign = d.GenerateSign(params)
	err = d.PostJson(queryWithdrawOrder, params, &queryWithdrawOrderResponse)
	if err != nil {
		return
	}
	if queryWithdrawOrderResponse.ErrNo != 0 {
		err = fmt.Errorf("queryWithdrawOrderResponse error %s %d", queryWithdrawOrderResponse.ErrTips, queryWithdrawOrderResponse.ErrNo)
		return
	}
	return
}

// MerchantWithdrawCallbackResponse 提现回调返回值解析
type MerchantWithdrawCallbackResponse struct {
	MsgSignature string `json:"msg_signature"`
	Nonce        string `json:"nonce"`
	Timestamp    string `json:"timestamp"`
	Type         string `json:"type"`
	Msg          string `json:"msg"`
	MsgStruct    MerchantWithdrawCallbackResponseMsg
}

type MerchantWithdrawCallbackResponseMsg struct {
	Status     string `json:"status"`
	Extra      string `json:"extra"`
	Message    string `json:"message"`
	WithdrawAt int    `json:"withdraw_at"`
	OrderId    string `json:"order_id"`
	OutOrderId string `json:"out_order_id"`
	ChOrderId  string `json:"ch_order_id"`
}

// MerchantWithdrawCallback 提现回调
func (d *DouYinOpenApi) MerchantWithdrawCallback(body string, checkSign bool) (merchantWithdrawCallbackResponse MerchantWithdrawCallbackResponse, err error) {
	err = json.Unmarshal([]byte(body), &merchantWithdrawCallbackResponse)
	if err != nil {
		return
	}
	// 判断是否需要校验签名
	if checkSign {
		sortedString := make([]string, 0)
		sortedString = append(sortedString, d.Config.Token)
		sortedString = append(sortedString, merchantWithdrawCallbackResponse.Timestamp)
		sortedString = append(sortedString, merchantWithdrawCallbackResponse.Nonce)
		sortedString = append(sortedString, merchantWithdrawCallbackResponse.Msg)
		err = d.CheckResponseSign(merchantWithdrawCallbackResponse.MsgSignature, sortedString)
		if err != nil {
			return
		}
	}
	var msgStruct MerchantWithdrawCallbackResponseMsg
	// 解析 msg 数据到结构体
	err = json.Unmarshal([]byte(merchantWithdrawCallbackResponse.Msg), &msgStruct)
	if err != nil {
		return
	}
	merchantWithdrawCallbackResponse.MsgStruct = msgStruct
	return
}

// OrderV2PushParams 订单推送
type OrderV2PushParams struct {
	ClientKey   string `json:"client_key,omitempty"`   // 否 第三方在抖音开放平台申请的 ClientKey 注意：POI 订单必传 awx1334dlkfjdf
	AccessToken string `json:"access_token,omitempty"` // 是 服务端 API 调用标识，通过 getAccessToken 获取
	ExtShopId   string `json:"ext_shop_id,omitempty"`  // 否 POI 店铺同步时使用的开发者侧店铺 ID，购买店铺 ID，长度 < 256 byte 注意：POI 订单必传 ext_112233
	AppName     string `json:"app_name,omitempty"`     // 是 做订单展示的字节系 app 名称，目前为固定值“douyin” douyin
	OpenId      string `json:"open_id,omitempty"`      // 是 小程序用户的 open_id，通过 code2Session 获取 d33432323423
	OrderStatus int64  `json:"order_status,omitempty"` // 否 普通小程序订单订单状态，POI 订单可以忽略 0：待支付 1：已支付 2：已取消 4：已核销（核销状态是整单核销,即一笔订单买了 3 个券，核销是指 3 个券核销的整单） 5：退款中 6：已退款 8：退款失败 注意：普通小程序订单必传，担保支付分账依赖该状态 4
	OrderType   int64  `json:"order_type,omitempty"`   // 是 订单类型，枚举值: 0：普通小程序订单（非POI订单） 9101：团购券订单（POI 订单） 9001：景区门票订单（POI订单）0
	UpdateTime  int64  `json:"update_time,omitempty"`  // 是 订单信息变更时间，13 位毫秒级时间戳 1643189272388
	Extra       string `json:"extra,omitempty"`        // 否 自定义字段，用于关联具体业务场景下的特殊参数，长度 < 2048byte
	OrderDetail string `json:"order_detail,omitempty"` // 是 订单详情，长度 < 2048byte
}

// OrderDetailPOI9101 POI 9101团购卷类型
type OrderDetailPOI9101 struct {
	// OrderV2PushParams
	// OrderDetail
}

type OrderDetailPOI9001 struct {
	// OrderV2PushParams
	// OrderDetail
}

// OrderDetailParams 订单详情
type OrderDetailParams struct {
	OrderId    string     `json:"order_id,omitempty"`    // 是 开发者侧业务单号。用作幂等控制。该订单号是和担保支付的支付单号绑定的，也就是预下单时传入的 out_order_no 字段，长度 <= 64byte 54bb46ba
	CreateTime int64      `json:"create_time,omitempty"` // 是 订单创建的时间，13 位毫秒时间戳 1648453349123
	Status     string     `json:"status,omitempty"`      // 是 订单状态，建议采用以下枚举值： 待支付 已支付 已取消 已超时 已核销 退款中 已退款 退款失败 已支付
	Amount     int64      `json:"amount,omitempty"`      //  是 订单商品总数 2
	TotalPrice int64      `json:"total_price,omitempty"` // 是 订单总价，单位为分 8800
	DetailUrl  string     `json:"detail_url,omitempty"`  // 是 小程序订单详情页 path，长度<=1024 byte
	ItemList   []ItemList `json:"item_list,omitempty"`   // list 是 子订单商品列表，不可为空
}

// ItemList 子订单商品列表
type ItemList struct {
	ItemCode string `json:"item_code,omitempty"` // 是 开发者侧商品 ID，长度 <= 64 byte test_item_code
	Img      string `json:"img,omitempty"`       // 是 子订单商品图片 URL，长度 <= 512 byte https://xxxxxxxxxxxxxxxxxxxxxx
	Title    string `json:"title,omitempty"`     // 是 子订单商品介绍标题，长度 <= 256 byte 好日子
	SubTitle string `json:"sub_title,omitempty"` // 否 子订单商品介绍副标题，长度 <= 256 byte
	Amount   int64  `json:"amount,omitempty"`    // 否 单类商品的数目 2
	Price    int64  `json:"price,omitempty"`     // 是 单类商品的总价，单位为分 4400
}

// OrderV2PushResponse 订单推送返回
type OrderV2PushResponse struct {
	ErrCode int    `json:"err_code"`
	ErrMsg  string `json:"err_msg"`
	Body    string `json:"body"`
}

// OrderV2Push 订单推送
func (d *DouYinOpenApi) OrderV2Push(normal OrderV2PushParams) (orderV2PushResponse OrderV2PushResponse, err error) {
	// normal.AppId = d.Config.AppId
	// normal.Sign = d.GenerateSign(params)
	err = d.PostJson(orderV2Push, normal, &orderV2PushResponse)
	if err != nil {
		return
	}
	if orderV2PushResponse.ErrCode != 0 {
		err = fmt.Errorf("OrderV2Push error %s %s", orderV2PushResponse.ErrMsg, orderV2PushResponse.Body)
		return
	}
	return
}

// UrlLinkGenerateParams urlLinkGenerate
type UrlLinkGenerateParams struct {
	AccessToken string `json:"access_token"`    // 接口调用凭证，调用getAccessToken生成的token
	MaAppID     string `json:"ma_app_id"`       // 小程序ID
	AppName     string `json:"app_name"`        // 宿主名称，可选 douyin，douyinlite
	Path        string `json:"path,omitempty"`  // 通过URL Link进入的小程序页面路径，必须是已经发布的小程序存在的页面，不可携带 query。path 为空时会跳转小程序主页。
	Query       string `json:"query,omitempty"` // 通过URL Link进入小程序时的 query（json形式），若无请填{}。最大1024个字符，只支持数字，大小写英文以及部分特殊字符：`{}!#$&'()*+,/:;=?@-._~%``。
	ExpireTime  int64  `json:"expire_time"`     // 到期失效的URL Link的失效时间。为 Unix 时间戳，实际失效时间为距离当前时间小时数，向上取整。最长间隔天数为180天。
}

// UrlLinkGenerateResponse urlLinkGenerate 返回值
type UrlLinkGenerateResponse struct {
	ErrNo   int    `json:"err_no"`
	ErrTips string `json:"err_tips"`
	UrlLink string `json:"url_link"`
}

// UrlLinkGenerate urlLinkGenerate
func (d *DouYinOpenApi) UrlLinkGenerate(params UrlLinkGenerateParams) (urlLinkGenerateResponse UrlLinkGenerateResponse, err error) {
	params.AccessToken, err = d.Config.AccessToken.GetAccessToken()
	if err != nil {
		return
	}
	err = d.PostJson(urlLinkGenerate, params, &urlLinkGenerateResponse)
	if err != nil {
		return
	}
	if urlLinkGenerateResponse.ErrNo != 0 {
		err = fmt.Errorf("urlLinkGenerateResponse error %s %d", urlLinkGenerateResponse.ErrTips, urlLinkGenerateResponse.ErrNo)
		return
	}
	return
}

// PostJsonWithHeader 封装公共的请求方法
func (d *DouYinOpenApi) PostJsonWithHeader(api string, params interface{}, response interface{}, header map[string]string) (err error) {
	body, err := util.PostJsonWithHeader(api, params, header)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return
	}
	return
}

// MsgTypeText 1 文本 2 图片 3 视频 9 群邀请卡片 4 文字/图片/电话拨打卡片 10 小程序卡片
const (
	MsgTypeText       = 1
	MsgTypeImage      = 2
	MsgTypeVideo      = 3
	MsgTypeGroup      = 9
	MsgTypeCard       = 4
	MsgTypeAppletCard = 10
)

// TextContent 文本消息内容
type TextContent struct {
	MsgType int `json:"msg_type"` // 消息类型：1 - 文本消息
	Text    struct {
		Text string `json:"text"` // 文本内容
	} `json:"text"` // 文本内容
}

// ImageContent 图片消息内容
type ImageContent struct {
	MsgType int `json:"msg_type"` // 消息类型：2 - 图片消息
	Image   struct {
		MediaID string `json:"media_id"` // 图片ID
	} `json:"image"` // 图篇内容
}

// VideoContent 视频消息内容
type VideoContent struct {
	MsgType int `json:"msg_type"` // 消息类型：3 - 视频消息
	Video   struct {
		ItemID string `json:"item_id"` // 视频ID
	} `json:"video"` // 视频内容
}

// GroupInvitationContent 群邀请卡片消息内容
type GroupInvitationContent struct {
	MsgType         int `json:"msg_type"` // 消息类型：3 - 视频消息
	GroupInvitation struct {
		GroupID string `json:"group_id"` // 群ID
	} `json:"group_invitation"`
}

// CardContent 文字/图片/电话拨打卡片消息内容
type CardContent struct {
	MsgType int `json:"msg_type"` // 消息类型：4 - 文字/图片/电话拨打卡片
	Card    struct {
		CardId string `json:"card_id"`
	} `json:"card"`
}

// AppletCardContent 小程序卡片消息内容
type AppletCardContent struct {
	MsgType    int `json:"msg_type"` // 消息类型：10 - 小程序卡片
	AppletCard struct {
		CardTemplateId string `json:"card_template_id,omitempty"`
		Path           string `json:"path,omitempty"`
		Query          string `json:"query,omitempty"`
		AppId          string `json:"app_id,omitempty"`
		Schema         string `json:"schema,omitempty"`
	} `json:"applet_card"`
}

// ImAuthorizeSendMsgParams 主动发送私信参数
type ImAuthorizeSendMsgParams struct {
	ToUserId string      `json:"to_user_id,omitempty"`
	Content  interface{} `json:"content,omitempty"`
}

// ImAuthorizeSendMsgResponse 发送消息返回值
type ImAuthorizeSendMsgResponse struct {
	ErrNo  int    `json:"err_no"`
	ErrMsg string `json:"err_msg"`
	Data   struct {
		MsgId string `json:"msg_id"`
	} `json:"data"`
	LogId string `json:"log_id"`
}

// SendPrivateMessage 主动发送私信
func (d *DouYinOpenApi) SendPrivateMessage(openId string, params ImAuthorizeSendMsgParams) (sendMsgResponse ImAuthorizeSendMsgResponse, err error) {
	token, err := d.Config.AccessToken.GetAccessToken()
	if err != nil {
		return
	}
	err = d.PostJsonWithHeader(fmt.Sprintf("%s?open_id=%s", imAuthorizeSendMsg, openId), params, &sendMsgResponse, map[string]string{
		"Content-Type": "application/json",
		"access-token": token,
	})
	if err != nil {
		return
	}
	if sendMsgResponse.ErrNo != 0 {
		err = fmt.Errorf("SendMsg error %s %d", sendMsgResponse.ErrMsg, sendMsgResponse.ErrNo)
		return
	}
	return
}

// ImAuthorizeRecallMsgParams 撤回私信消息参数
type ImAuthorizeRecallMsgParams struct {
	MsgId          string `json:"msg_id"`
	ConversationId string `json:"conversation_id"`
}

// ImAuthorizeRecallMsgResponse 撤回私信消息返回值
type ImAuthorizeRecallMsgResponse struct {
	ErrNo  int    `json:"err_no"`
	ErrMsg string `json:"err_msg"`
	LogId  string `json:"log_id"`
}

// RecallPrivateMessage 撤回私信消息
func (d *DouYinOpenApi) RecallPrivateMessage(openId string, params ImAuthorizeRecallMsgParams) (recallMsgResponse ImAuthorizeRecallMsgResponse, err error) {
	token, err := d.Config.AccessToken.GetAccessToken()
	if err != nil {
		return
	}
	err = d.PostJsonWithHeader(fmt.Sprintf("%s?open_id=%s", imAuthorizeRecallMsg, openId), params, &recallMsgResponse, map[string]string{
		"Content-Type": "application/json",
		"access-token": token,
	})
	if err != nil {
		return
	}
	if recallMsgResponse.ErrNo != 0 {
		err = fmt.Errorf("RecallMsg error %s %d", recallMsgResponse.ErrMsg, recallMsgResponse.ErrNo)
		return
	}
	return
}

// imGroupFansListResponse 查询群信息返回值
type imGroupFansListResponse struct {
}

// WebHookResponse webHook回调返回值
type WebHookResponse struct {
	ErrNo   int    `json:"err_no"`
	ErrTips string `json:"err_tips"`
}

// WebHookBody webHook回调body
type WebHookBody struct {
	Event                string                          `json:"event"`
	ClientKey            string                          `json:"client_key"`
	FromUserId           string                          `json:"from_user_id"`
	ToUserId             string                          `json:"to_user_id"`
	Content              string                          `json:"content"`
	VerifyWebhookContent VerifyWebhook                   `json:"verify_webhook_content,omitempty"`
	ImAuthorizeContent   EventImAuthorize                `json:"im_authorize_content,omitempty"`
	MsgCallbackContent   EventImAuthorizeMessageCallback `json:"msg_callback_content,omitempty"`
}

// EventImAuthorize 主动授权事件订阅
type EventImAuthorize struct {
	CreateTime    int    `json:"create_time"`
	ExpireTime    int    `json:"expire_time"`
	UpdateTime    int    `json:"update_time"`
	OperationType int    `json:"operation_type"`
	AuthStatus    int    `json:"auth_status"`
	Source        string `json:"source"`
	Extra         string `json:"extra"`
	LogId         string `json:"log_id"`
}

// EventImAuthorizeMessageCallback 主动私信发送回调
type EventImAuthorizeMessageCallback struct {
	ConversationShortId string    `json:"conversation_short_id"`
	ServerMessageId     string    `json:"server_message_id"`
	ConversationType    int       `json:"conversation_type"`
	CreateTime          int64     `json:"create_time"`
	MessageType         string    `json:"message_type"`
	Text                string    `json:"text"`
	Image               string    `json:"image"`
	ItemId              string    `json:"item_id"`
	Content             string    `json:"content"`
	Title               string    `json:"title"`
	Actions             string    `json:"actions"`
	UserInfos           UserInfos `json:"user_infos"`
}

// UserInfos 用户信息
type UserInfos struct {
	OpenId   string `json:"open_id"`
	NickName string `json:"nick_name"`
	Avatar   string `json:"avatar"`
}

// VerifyWebhook 验证接口
type VerifyWebhook struct {
	Challenge int `json:"challenge"`
}

// generateSignature
func (d *DouYinOpenApi) generateSignature(body string) (signature string) {
	h := sha1.New()
	h.Write([]byte(d.Config.AppSecret))
	h.Write([]byte(body))
	return hex.EncodeToString(h.Sum(nil))
}

// HandleWebHook 延签 + 解析返回值
func (d *DouYinOpenApi) HandleWebHook(xDouyinSignature, body string, checkSign bool) (webHookBody WebHookBody, err error) {
	if checkSign == true && d.generateSignature(body) != xDouyinSignature {
		err = fmt.Errorf("signature error new: %s old: %s", d.generateSignature(body), xDouyinSignature)
		return
	}
	err = json.Unmarshal([]byte(body), &webHookBody)
	if err != nil {
		return
	}
	switch webHookBody.Event {
	case "verify_webhook":
		var verifyWebhook VerifyWebhook
		err = json.Unmarshal([]byte(webHookBody.Content), &verifyWebhook)
		if err != nil {
			return
		}
		webHookBody.VerifyWebhookContent = verifyWebhook
	case "im_authorize":
		var eventImAuthorize EventImAuthorize
		err = json.Unmarshal([]byte(webHookBody.Content), &eventImAuthorize)
		if err != nil {
			return
		}
		webHookBody.ImAuthorizeContent = eventImAuthorize
	case "im_authorize_message_callback":
		var eventImAuthorizeMessageCallback EventImAuthorizeMessageCallback
		err = json.Unmarshal([]byte(webHookBody.Content), &eventImAuthorizeMessageCallback)
		if err != nil {
			return
		}
		webHookBody.MsgCallbackContent = eventImAuthorizeMessageCallback
	default:
		err = fmt.Errorf("event %s not support", webHookBody.Event)
		return
	}
	return
}
