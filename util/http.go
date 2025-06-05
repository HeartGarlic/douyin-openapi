package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPDoer represents the minimal interface required to execute an HTTP request.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

var defaultClient HTTPDoer = &http.Client{Timeout: 10 * time.Second}

// PostForm post form 数据请求
func PostFormCtx(ctx context.Context, client HTTPDoer, uri string, obj url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, strings.NewReader(obj.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http get error : uri=%v , statusCode=%v", uri, response.StatusCode)
	}
	return ioutil.ReadAll(response.Body)
}

func PostForm(uri string, obj url.Values) ([]byte, error) {
	return PostFormCtx(context.Background(), defaultClient, uri, obj)
}

// PostJSON post json 数据请求
func PostJSONCtx(ctx context.Context, client HTTPDoer, uri string, obj interface{}) ([]byte, error) {
	marshal, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewBuffer(marshal))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http get error : uri=%v , statusCode=%v", uri, response.StatusCode)
	}
	return ioutil.ReadAll(response.Body)
}

func PostJSON(uri string, obj interface{}) ([]byte, error) {
	return PostJSONCtx(context.Background(), defaultClient, uri, obj)
}

// JsonStructToMap ...
func JsonStructToMap(content interface{}) (map[string]interface{}, error) {
	var name map[string]interface{}
	if marshalContent, err := json.Marshal(content); err != nil {
		return name, err
	} else {
		d := json.NewDecoder(bytes.NewReader(marshalContent))
		d.UseNumber() // 设置将float64转为一个number
		if err := d.Decode(&name); err != nil {
			return name, err
		} else {
			for k, v := range name {
				name[k] = v
			}
		}
	}
	return name, nil
}
