package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/taking/kubemigrate/pkg/utils"
)

func main() {
	fmt.Println("=== Utils Package Usage Example ===")

	// 1. 문자열 기본값 처리
	fmt.Println("\n1. String default value handling...")

	testCases := []struct {
		value    string
		default_ string
		expected string
	}{
		{"", "default value", "default value"},
		{"actual value", "default value", "actual value"},
		{"   ", "default value", "   "}, // 공백은 빈 문자열이 아님
		{"0", "default value", "0"},
	}

	for _, tc := range testCases {
		result := utils.GetStringOrDefault(tc.value, tc.default_)
		fmt.Printf("  GetStringOrDefault(%q, %q) = %q\n", tc.value, tc.default_, result)
	}

	// 2. 불린 기본값 처리
	fmt.Println("\n2. Boolean default value handling...")

	boolTestCases := []struct {
		value    bool
		default_ bool
		expected bool
	}{
		{true, false, true},
		{false, true, false},
		{true, true, true},
		{false, false, false},
	}

	for _, tc := range boolTestCases {
		result := utils.GetBoolOrDefault(tc.value, tc.default_)
		fmt.Printf("  GetBoolOrDefault(%v, %v) = %v\n", tc.value, tc.default_, result)
	}

	// 3. 문자열을 정수로 변환
	fmt.Println("\n3. String to integer conversion...")

	intTestCases := []struct {
		value    string
		default_ int
		expected int
	}{
		{"123", 0, 123},
		{"-456", 0, -456},
		{"0", 999, 0},
		{"invalid", 999, 999},
		{"", 999, 999},
		{"12.34", 999, 999}, // 소수점은 정수가 아님
	}

	for _, tc := range intTestCases {
		result := utils.StringToIntOrDefault(tc.value, tc.default_)
		fmt.Printf("  StringToIntOrDefault(%q, %d) = %d\n", tc.value, tc.default_, result)
	}

	// 4. 문자열을 불린으로 변환
	fmt.Println("\n4. String to boolean conversion...")

	boolStringTestCases := []struct {
		value    string
		default_ bool
		expected bool
	}{
		{"true", false, true},
		{"false", true, false},
		{"TRUE", false, true},
		{"FALSE", true, false},
		{"1", false, true},
		{"0", true, false},
		{"yes", false, true},
		{"no", true, false},
		{"invalid", false, false},
		{"", true, true},
	}

	for _, tc := range boolStringTestCases {
		result := utils.StringToBoolOrDefault(tc.value, tc.default_)
		fmt.Printf("  StringToBoolOrDefault(%q, %v) = %v\n", tc.value, tc.default_, result)
	}

	// 5. 타임아웃을 사용한 함수 실행
	fmt.Println("\n5. Function execution with timeout...")

	// 빠른 함수 실행
	ctx1, cancel1 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel1()

	err := utils.RunWithTimeout(ctx1, func() error {
		time.Sleep(100 * time.Millisecond) // 100ms 대기
		fmt.Println("  ✅ Fast function execution completed")
		return nil
	})
	if err != nil {
		log.Printf("Fast function execution failed: %v", err)
	}

	// 타임아웃이 발생하는 함수 실행
	ctx2, cancel2 := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel2()

	err = utils.RunWithTimeout(ctx2, func() error {
		time.Sleep(1 * time.Second) // 1초 대기 (타임아웃 발생)
		fmt.Println("  This message will not be printed")
		return nil
	})
	if err != nil {
		fmt.Printf("  ⏰ Expected timeout: %v\n", err)
	}

	// 6. 실제 사용 시나리오 예제
	fmt.Println("\n6. Real-world usage scenario example...")

	// 환경 변수에서 설정값 읽기 시뮬레이션
	envVars := map[string]string{
		"PORT":           "8080",
		"DEBUG":          "true",
		"MAX_WORKERS":    "10",
		"TIMEOUT":        "30",
		"ENABLE_LOGGING": "false",
		"INVALID_NUMBER": "abc",
		"EMPTY_VALUE":    "",
	}

	// 설정값 파싱
	port := utils.StringToIntOrDefault(envVars["PORT"], 3000)
	debug := utils.StringToBoolOrDefault(envVars["DEBUG"], false)
	maxWorkers := utils.StringToIntOrDefault(envVars["MAX_WORKERS"], 5)
	timeout := utils.StringToIntOrDefault(envVars["TIMEOUT"], 60)
	enableLogging := utils.StringToBoolOrDefault(envVars["ENABLE_LOGGING"], true)
	invalidNumber := utils.StringToIntOrDefault(envVars["INVALID_NUMBER"], 999)
	emptyValue := utils.GetStringOrDefault(envVars["EMPTY_VALUE"], "default")

	fmt.Printf("  Parsed configuration values:\n")
	fmt.Printf("    PORT: %d\n", port)
	fmt.Printf("    DEBUG: %v\n", debug)
	fmt.Printf("    MAX_WORKERS: %d\n", maxWorkers)
	fmt.Printf("    TIMEOUT: %d seconds\n", timeout)
	fmt.Printf("    ENABLE_LOGGING: %v\n", enableLogging)
	fmt.Printf("    INVALID_NUMBER: %d (using default)\n", invalidNumber)
	fmt.Printf("    EMPTY_VALUE: %q (using default)\n", emptyValue)

	fmt.Println("\n=== Utils Package Example Completed ===")
}
