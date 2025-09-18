package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	pb "qahub/api/proto/user" // 导入生成的 protobuf 包

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const grpcAddress = "localhost:50051"

func printStep(step int, description string) {
	fmt.Printf("\n--- 步骤 %d: %s ---\n", step, description)
}

// checkResult 用于检查 gRPC 调用结果
func checkResult(step int, testName string, err error, expectedCodes ...codes.Code) {
	// Case 1: 预期没有错误，实际上也没有错误
	if err == nil {
		if len(expectedCodes) > 0 {
			log.Fatalf("[步骤 %d: %s] ❌ 失败: 预期错误码 %v, 但没有错误发生\n", step, testName, expectedCodes)
		}
		log.Printf("[步骤 %d: %s] ✅ 成功\n", step, testName)
		return
	}

	// Case 2: 发生了错误，检查是否是预期的 gRPC 错误
	st, ok := status.FromError(err)
	if !ok {
		log.Fatalf("[步骤 %d: %s] ❌ 失败: 收到非gRPC错误: %v\n", step, testName, err)
	}

	if len(expectedCodes) > 0 {
		match := false
		for _, code := range expectedCodes {
			if st.Code() == code {
				match = true
				break
			}
		}
		if match {
			log.Printf("[步骤 %d: %s] ✅ 成功 (收到预期的错误码: %s)\n", step, testName, st.Code())
		} else {
			log.Fatalf("[步骤 %d: %s] ❌ 失败: 预期错误码 %v, 但收到 %s\n", step, testName, expectedCodes, st.Code())
		}
	} else {
		// Case 3: 预期没有错误，但发生了错误
		log.Fatalf("[步骤 %d: %s] ❌ 失败: 发生未预期的gRPC错误: %v\n", step, testName, err)
	}
}

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	// --- 测试数据 ---
	uniqueUsername := fmt.Sprintf("testuser_%d", rand.Intn(100000))
	uniqueEmail := fmt.Sprintf("test_%d@example.com", rand.Intn(100000))
	password := "strongpassword123"
	var registeredUserID int64

	// --- 设置 gRPC 客户端 ---
	conn, err := grpc.NewClient(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("无法连接到 gRPC 服务器: %v", err)
	}
	defer conn.Close()
	client := pb.NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// --- 步骤 1: 注册新用户 ---
	printStep(1, "注册新用户")
	registerResp, err := client.Register(ctx, &pb.RegisterRequest{
		Username: uniqueUsername,
		Email:    uniqueEmail,
		Password: password,
		Bio:      "Hello, I'm a gRPC test user.",
	})
	checkResult(1, "注册", err)
	if registerResp != nil && registerResp.User != nil {
		registeredUserID = registerResp.User.Id
		fmt.Printf("响应内容: 用户ID %d, 用户名 %s\n", registeredUserID, registerResp.User.Username)
	} else {
		log.Fatalf("步骤 1 ❌ 失败: 注册响应为空")
	}

	// --- 步骤 2: 登录用户 ---
	printStep(2, "登录用户以获取Token")
	loginResp, err := client.Login(ctx, &pb.LoginRequest{Username: uniqueUsername, Password: password})
	checkResult(2, "登录", err)
	if loginResp == nil || loginResp.Token == "" {
		log.Fatalf("步骤 2 ❌ 失败: 未能获取Token")
	}
	loginToken := loginResp.Token
	fmt.Printf("响应内容: Token %s...\n", loginToken[:10])

	// --- 创建带有认证信息的 context ---
	md := metadata.New(map[string]string{"authorization": "Bearer " + loginToken})
	authedCtx := metadata.NewOutgoingContext(ctx, md)

	// --- 步骤 3: 更新个人信息 (成功场景) ---
	printStep(3, "更新个人信息 (成功场景)")
	_, err = client.UpdateUserProfile(authedCtx, &pb.UpdateUserProfileRequest{
		UserId:   registeredUserID,
		Username: uniqueUsername,
		Email:    uniqueEmail,
		Bio:      "My bio has been updated via gRPC!",
	})
	checkResult(3, "更新自己的信息", err)

	// --- 步骤 4: 更新他人信息 (失败场景) ---
	printStep(4, "更新他人信息 (失败场景)")
	otherUserID := int64(99999)
	_, err = client.UpdateUserProfile(authedCtx, &pb.UpdateUserProfileRequest{
		UserId:   otherUserID,
		Username: "otheruser",
		Email:    "other@example.com",
		Bio:      "Trying to update other user's bio",
	})
	checkResult(4, "更新他人信息", err, codes.PermissionDenied)

	// --- 步骤 5: 删除用户 (成功场景) ---
	printStep(5, "删除用户 (成功场景)")
	_, err = client.DeleteUser(authedCtx, &pb.DeleteUserRequest{UserId: registeredUserID})
	checkResult(5, "删除用户", err)

	// --- 步骤 6: 验证删除 ---
	printStep(6, "验证删除 (使用已删除账户登录)")
	_, err = client.Login(ctx, &pb.LoginRequest{Username: uniqueUsername, Password: password})
	// 服务层返回的是通用错误，gRPC 默认会包装成 Unknown。更完善的实现应返回如 NotFound 或 Unauthenticated
	checkResult(6, "使用已删除账户登录", err, codes.Unknown, codes.NotFound, codes.Unauthenticated)

	fmt.Println("\n🎉🎉🎉 所有 gRPC 测试步骤均按预期完成! 🎉🎉🎉")
}
