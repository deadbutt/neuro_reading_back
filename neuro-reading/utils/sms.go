package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type AliyunSMS struct {
	AccessKeyID     string
	AccessKeySecret string
	SignName        string
	TemplateCode    string
}

func NewAliyunSMS(accessKeyID, accessKeySecret, signName, templateCode string) *AliyunSMS {
	return &AliyunSMS{
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		SignName:        signName,
		TemplateCode:    templateCode,
	}
}

func (s *AliyunSMS) SendSMS(phoneNumber, code string) error {
	params := map[string]string{
		"AccessKeyId":      s.AccessKeyID,
		"Action":           "SendSms",
		"PhoneNumbers":     phoneNumber,
		"SignName":         s.SignName,
		"TemplateCode":     s.TemplateCode,
		"TemplateParam":    fmt.Sprintf(`{"code":"%s"}`, code),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":          "2017-05-25",
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
		"SignatureNonce":   strconv.FormatInt(time.Now().UnixNano(), 10),
	}

	params["Signature"] = s.sign(params)

	data := url.Values{}
	for k, v := range params {
		data.Set(k, v)
	}

	resp, err := http.Post("https://dysmsapi.aliyuncs.com/?"+data.Encode(), "application/x-www-form-urlencoded", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("SMS Response: %s\n", string(body))

	if resp.StatusCode != 200 {
		return fmt.Errorf("sms api error: %s", string(body))
	}
	return nil
}

func (s *AliyunSMS) sign(params map[string]string) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var queryParts []string
	for _, k := range keys {
		queryParts = append(queryParts, url.QueryEscape(k)+"="+url.QueryEscape(params[k]))
	}

	signString := "POST&%2F&" + url.QueryEscape(strings.Join(queryParts, "&"))
	key := s.AccessKeySecret + "&"

	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(signString))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}