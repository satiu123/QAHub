package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8080/api"

// --- Structs for API communication ---

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
}

type UserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Bio      string `json:"bio,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type UpdateProfileRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Bio      string `json:"bio"`
}

var client = &http.Client{Timeout: 10 * time.Second}

// --- Helper Functions ---

func printStep(step int, description string) {
	fmt.Printf("\n--- 步骤 %d: %s ---\n", step, description)
}

func makeRequest(method, url, token string, body interface{}) (*http.Response, []byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	fmt.Printf("请求: %s %s\n", method, url)
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return resp, respBody, nil
}

func checkResult(step int, resp *http.Response, body []byte, expectedStatus int) {
	fmt.Printf("响应状态码: %d (预期: %d)\n", resp.StatusCode, expectedStatus)
	fmt.Printf("响应内容: %s\n", string(body))
	if resp.StatusCode == expectedStatus {
		log.Printf("[步骤 %d] ✅ 成功\n", step)
	} else {
		log.Fatalf("[步骤 %d] ❌ 失败: 状态码不匹配\n", step)
	}
}

// --- Main Test Logic ---

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	// --- Test Data ---
	uniqueUsername := fmt.Sprintf("testuser_%d", rand.Intn(100000))
	uniqueEmail := fmt.Sprintf("test_%d@example.com", rand.Intn(100000))
	password := "strongpassword123"
	var registeredUser UserResponse
	var loginToken string

	// --- 步骤 1: 注册新用户 ---
	printStep(1, "注册新用户")
	registerPayload := RegisterRequest{
		Username: uniqueUsername,
		Email:    uniqueEmail,
		Password: password,
		Bio:      "Hello, I'm a test user.",
	}
	resp, body, err := makeRequest("POST", baseURL+"/users/register", "", registerPayload)
	if err != nil {
		log.Fatalf("步骤 1 ❌ 失败: 请求错误: %v", err)
	}
	checkResult(1, resp, body, http.StatusCreated)
	if err := json.Unmarshal(body, &registeredUser); err != nil {
		log.Fatalf("步骤 1 ❌ 失败: 解析响应失败: %v", err)
	}

	// --- 步骤 2: 登录用户 ---
	printStep(2, "登录用户以获取Token")
	loginPayload := LoginRequest{Username: uniqueUsername, Password: password}
	resp, body, err = makeRequest("POST", baseURL+"/users/login", "", loginPayload)
	if err != nil {
		log.Fatalf("步骤 2 ❌ 失败: 请求错误: %v", err)
	}
	checkResult(2, resp, body, http.StatusOK)
	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		log.Fatalf("步骤 2 ❌ 失败: 解析响应失败: %v", err)
	}
	loginToken = loginResp.Token
	if loginToken == "" {
		log.Fatalf("步骤 2 ❌ 失败: 未能获取Token")
	}

	// --- 步骤 3: 更新个人信息 (成功场景) ---
	printStep(3, "更新个人信息 (成功场景)")
	updatePayload := UpdateProfileRequest{
		Username: uniqueUsername, // Keep username same for simplicity, or can be updated
		Email:    uniqueEmail,    // Keep email same
		Bio:      "My bio has been updated!",
	}
	updateURL := fmt.Sprintf("%s/users/%d", baseURL, registeredUser.ID)
	resp, body, err = makeRequest("PUT", updateURL, loginToken, updatePayload)
	if err != nil {
		log.Fatalf("步骤 3 ❌ 失败: 请求错误: %v", err)
	}
	checkResult(3, resp, body, http.StatusOK)

	// --- 步骤 4: 更新他人信息 (失败场景) ---
	printStep(4, "更新他人信息 (失败场景)")
	otherUserID := 99999
	updateOtherURL := fmt.Sprintf("%s/users/%d", baseURL, otherUserID)
	resp, body, err = makeRequest("PUT", updateOtherURL, loginToken, updatePayload)
	if err != nil {
		log.Fatalf("步骤 4 ❌ 失败: 请求错误: %v", err)
	}
	checkResult(4, resp, body, http.StatusForbidden)

	// --- 步骤 5: 删除用户 (成功场景) ---
	printStep(5, "删除用户 (成功场景)")
	deleteURL := fmt.Sprintf("%s/users/%d", baseURL, registeredUser.ID)
	resp, body, err = makeRequest("DELETE", deleteURL, loginToken, nil)
	if err != nil {
		log.Fatalf("步骤 5 ❌ 失败: 请求错误: %v", err)
	}
	checkResult(5, resp, body, http.StatusOK)

	// --- 步骤 6: 验证删除 ---
	printStep(6, "验证删除 (使用已删除账户登录)")
	resp, body, err = makeRequest("POST", baseURL+"/users/login", "", loginPayload)
	if err != nil {
		log.Fatalf("步骤 6 ❌ 失败: 请求错误: %v", err)
	}
	checkResult(6, resp, body, http.StatusUnauthorized)

	fmt.Println("\n🎉🎉🎉 所有测试步骤均按预期完成! 🎉🎉🎉")
}
