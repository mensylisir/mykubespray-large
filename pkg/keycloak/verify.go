package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
)

func main1() {
	//token := "eyJhbGciOiJSUzI1NiIsImtpZCI6IkJNaExxcHB2SFFJX2JDQUtRT3c0SkZIY1psd3IxbXhoSnN5c2FhaW00djAiLCJ0eXAiOiJKV1QifQ.eyJhY3IiOiIxIiwiYXRfaGFzaCI6InJGM2F0dXMwZVF2TXZVeEQ1YmFuWXciLCJhdWQiOiJrdWJlcm5ldGVzIiwiYXV0aF90aW1lIjowLCJhenAiOiJrdWJlcm5ldGVzIiwiZW1haWxfdmVyaWZpZWQiOmZhbHNlLCJleHAiOjE3MjMxODU4MTksImdyb3VwcyI6WyJrdWJlcm5ldGVzLXZpZXdlciJdLCJpYXQiOjE3MjIzMjU0MTksImlzcyI6Imh0dHBzOi8va2V5Y2xvYWsua21wcC5pby9hdXRoL3JlYWxtcy9jYXJzIiwianRpIjoiMjliMTBjMTEtYjZhNi00NDZmLTk3MzUtNDU0ODNmMjUzZDMwIiwibmFtZSI6Indhbmd6aGVuZG9uZyIsInByZWZlcnJlZF91c2VybmFtZSI6Indhbmd6aGVuZG9uZyIsInNlc3Npb25fc3RhdGUiOiI5ODA3YTdiZi05ODhmLTQ5YTYtODhiMC0xNjEzZWJkMjcwYmIiLCJzaWQiOiI5ODA3YTdiZi05ODhmLTQ5YTYtODhiMC0xNjEzZWJkMjcwYWEiLCJzdWIiOiI0MzkxYzI4OC05NjVmLTRkNTAtYWRjYS00OTliMGNjZWNiMzciLCJ0eXAiOiJJRCJ9.urNVZryOHVFcDn5kqKV3T8Z08gXftE-MiaFcjwXGj9AbfanscAR6jt2N-vWaL2sQAyvrYIOJWMngcx3DcVeD2w" // 替换为实际生成的 token

	// Keycloak 认证端点 URL
	introspectURL := "https://172.30.1.12:31147/api/v1/namespaces"

	// 准备 POST 请求的 body 数据
	//data := strings.NewReader(fmt.Sprintf("token=%s", token))

	// 创建 HTTP POST 请求
	req, err := http.NewRequest("GET", introspectURL, nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IkJNaExxcHB2SFFJX2JDQUtRT3c0SkZIY1psd3IxbXhoSnN5c2FhaW00djAiLCJ0eXAiOiJKV1QifQ.eyJhY3IiOiIxIiwiYXRfaGFzaCI6InJGM2F0dXMwZVF2TXZVeEQ1YmFuWXciLCJhdWQiOiJrdWJlcm5ldGVzIiwiYXV0aF90aW1lIjowLCJhenAiOiJrdWJlcm5ldGVzIiwiZW1haWxfdmVyaWZpZWQiOmZhbHNlLCJleHAiOjE3MjMxODY1OTYsImdyb3VwcyI6WyJrdWJlcm5ldGVzLXZpZXdlciJdLCJpYXQiOjE3MjIzMjYxOTYsImlzcyI6Imh0dHBzOi8va2V5Y2xvYWsua21wcC5pby9hdXRoL3JlYWxtcy9jYXJzIiwianRpIjoiMjliMTBjMTEtYjZhNi00NDZmLTk3MzUtNDU0ODNmMjUzZDMwIiwibmFtZSI6Indhbmd6aGVuZG9uZzEiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJ3YW5nemhlbmRvbmcxIiwic2Vzc2lvbl9zdGF0ZSI6Ijk4MDdhN2JmLTk4OGYtNDlhNi04OGIwLTE2MTNlYmQyNzBiYiIsInNpZCI6Ijk4MDdhN2JmLTk4OGYtNDlhNi04OGIwLTE2MTNlYmQyNzBhYSIsInN1YiI6IjQzOTFjMjg4LTk2NWYtNGQ1MC1hZGNhLTQ5OWIwY2NlY2IzNyIsInR5cCI6IklEIn0.u61ejOWL-lYnzVhNnJUQ6PcyYrq5flropA2ZY7VF-yO9vO-01S31kATPJmxBzcpaz7cCxqrDlGJQuNAqtDzAiQ")
	if err != nil {
		fmt.Println("Failed to create HTTP request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 忽略 SSL 证书验证
	}
	// 发送请求并获取响应
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send HTTP request:", err)
		return
	}
	defer resp.Body.Close()

	// 解析响应
	var introspectionResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&introspectionResponse)
	if err != nil {
		fmt.Println("Failed to parse introspection response:", err)
		return
	}
	data, _ := json.Marshal(introspectionResponse)
	// 输出响应
	fmt.Println("Introspection Response:", string(data))
}
