package main

import (
	"bytes"
	"encoding/json"
	"flag" // コマンドライン引数解析用パッケージをインポート
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"io"
	"log"
	"net/http"

	"time"
)

const (
	freeeBaseURL = "https://api.freee.co.jp/hr"
)

// --- Structs for API Responses --- (変更なし)
type UserInfoResponse struct {
	ID        int           `json:"id"`
	Companies []CompanyInfo `json:"companies"`
}

type CompanyInfo struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Role        string  `json:"role"`
	ExternalID  string  `json:"external_id"`
	EmployeeID  *int    `json:"employee_id"` // Use pointer to handle potential null
	DisplayName *string `json:"display_name"`
}

// --- Structs for API Request Body --- (変更なし)
type WorkRecordPayload struct {
	CompanyID          int                 `json:"company_id"`
	BreakRecords       []TimeRecord        `json:"break_records"`
	WorkRecordSegments []WorkRecordSegment `json:"work_record_segments"`
	// Note string `json:"note,omitempty"` // Optional: Add if you need notes
}

type TimeRecord struct {
	ClockInAt  string `json:"clock_in_at"`  // Format: "YYYY-MM-DD HH:MM:SS"
	ClockOutAt string `json:"clock_out_at"` // Format: "YYYY-MM-DD HH:MM:SS"
}

type WorkRecordSegment struct {
	ClockInAt  string `json:"clock_in_at"`  // Format: "YYYY-MM-DD HH:MM:SS"
	ClockOutAt string `json:"clock_out_at"` // Format: "YYYY-MM-DD HH:MM:SS"
}

// --- Configuration ---

var (
	// targetYear, targetMonth は引数から取得するため削除
	accessToken     = "YOUR_ACCESS_TOKEN" // .env または環境変数で上書きされます
	targetCompanyId = 0                   // .env または環境変数で上書きされます

	// Define the daily times (変更なし)
	dailyClockInTime  = "11:00:00"
	dailyClockOutTime = "21:00:00"
	dailyBreakStart   = "12:00:00"
	dailyBreakEnd     = "13:00:00"
)

// --- Main Logic ---

func main() {
	// --- コマンドライン引数の定義 ---
	// デフォルト値を0にすることで、指定がない場合に検出できるようにする
	yearPtr := flag.Int("year", 0, "Target year (e.g., 2025) (required)")
	monthPtr := flag.Int("month", 0, "Target month as a number (1-12) (required)")
	listCompaniesPtr := flag.Bool("list-companies", false, "List all companies associated with your account")

	// コマンドライン引数をパース
	flag.Parse()

	// --- .env/環境変数からの設定読み込み (変更なし) ---
	err := godotenv.Load()
	if err != nil {
		log.Println("Info: .env file not found or failed to load. Using environment variables or defaults.")
	}

	tokenEnv := os.Getenv("TOKEN")
	if tokenEnv != "" {
		accessToken = tokenEnv
	}
	companyIdEnv := os.Getenv("COMPANY_ID")
	if companyIdEnv != "" {
		parsedId, convErr := strconv.Atoi(companyIdEnv)
		if convErr != nil {
			log.Fatalf("Error: Invalid COMPANY_ID '%s' in environment variable: %v", companyIdEnv, convErr)
		}
		targetCompanyId = parsedId
	}

	if accessToken == "YOUR_ACCESS_TOKEN" || accessToken == "" {
		log.Fatal("Error: Please set your freee API access token either in code, .env file (TOKEN=...), or environment variable.")
	}
	// --- 設定読み込み終了 ---

	// --- 会社一覧表示モード ---
	if *listCompaniesPtr {
		err := listCompanies(accessToken)
		if err != nil {
			log.Fatalf("Error listing companies: %v", err)
		}
		return
	}

	// --- 引数の検証 ---
	if *yearPtr == 0 || *monthPtr == 0 {
		fmt.Println("Error: Both -year and -month flags are required.")
		fmt.Println("Usage:")
		flag.PrintDefaults() // ヘルプメッセージを表示
		os.Exit(1)           // エラー終了
	}
	if *monthPtr < 1 || *monthPtr > 12 {
		fmt.Printf("Error: Invalid month value %d. Month must be between 1 and 12.\n", *monthPtr)
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// 検証済みの値を代入
	targetYear := *yearPtr
	targetMonth := *monthPtr // time.Month型ではなくint型(1-12)として保持

	if targetCompanyId == 0 {
		log.Fatal("Error: Please set your targetCompanyId either in code, .env file (COMPANY_ID=...), or environment variable.")
	}

	log.Println("Fetching employee ID...")
	employeeID, err := getEmployeeID(accessToken, targetCompanyId)
	if err != nil {
		log.Fatalf("Error getting employee ID: %v", err)
	}
	log.Printf("Found Employee ID: %d for Company ID: %d\n", employeeID, targetCompanyId)

	// 引数で指定された年月で処理開始ログ
	log.Printf("Registering attendance for all days in %d-%02d...\n", targetYear, targetMonth)

	// 対象月の初日を取得 (int型の月を time.Month 型にキャスト)
	currentDate := time.Date(targetYear, time.Month(targetMonth), 1, 0, 0, 0, 0, time.UTC)

	// 対象月である間ループ (比較対象も time.Month 型にキャスト)
	targetMonthAsTimeMonth := time.Month(targetMonth)
	for currentDate.Month() == targetMonthAsTimeMonth {
		dateStr := currentDate.Format("2006-01-02")
		log.Printf("Processing date: %s...\n", dateStr)

		// その日の勤怠を登録
		err = registerAttendanceForSingleDate(accessToken, targetCompanyId, employeeID, currentDate)
		if err != nil {
			log.Fatalf("Error registering attendance for %s: %v", dateStr, err)
		}

		// 次の日に進む
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	log.Printf("Attendance registration process completed for %d-%02d.", targetYear, targetMonth)
}

// --- API Interaction Functions ---

// listCompanies は、認証されたユーザーに関連する会社の一覧を取得して表示します
func listCompanies(token string) error {
	url := freeeBaseURL + "/api/v1/users/me"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("accept", "application/json")
	req.Header.Set("FREEE-VERSION", "2022-02-01")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status %s: %s", resp.Status, string(bodyBytes))
	}

	var userInfoResp UserInfoResponse
	err = json.Unmarshal(bodyBytes, &userInfoResp)
	if err != nil {
		return fmt.Errorf("failed to parse JSON response: %w. Response body: %s", err, string(bodyBytes))
	}

	fmt.Println("\n=== 所属会社一覧 / Company List ===")
	fmt.Printf("User ID: %d\n\n", userInfoResp.ID)

	if len(userInfoResp.Companies) == 0 {
		fmt.Println("No companies found.")
		return nil
	}

	for i, company := range userInfoResp.Companies {
		fmt.Printf("[%d] Company ID: %d\n", i+1, company.ID)
		fmt.Printf("    Name: %s\n", company.Name)
		fmt.Printf("    Role: %s\n", company.Role)
		if company.DisplayName != nil {
			fmt.Printf("    Display Name: %s\n", *company.DisplayName)
		}
		if company.EmployeeID != nil {
			fmt.Printf("    Employee ID: %d\n", *company.EmployeeID)
		} else {
			fmt.Printf("    Employee ID: (Not registered as employee)\n")
		}
		if company.ExternalID != "" {
			fmt.Printf("    External ID: %s\n", company.ExternalID)
		}
		fmt.Println()
	}

	return nil
}

// getEmployeeID 関数 (変更なし)
func getEmployeeID(token string, companyID int) (int, error) {
	url := freeeBaseURL + "/api/v1/users/me"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("accept", "application/json")
	req.Header.Set("FREEE-VERSION", "2022-02-01")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API request failed with status %s: %s", resp.Status, string(bodyBytes))
	}

	var userInfoResp UserInfoResponse
	err = json.Unmarshal(bodyBytes, &userInfoResp)
	if err != nil {
		return 0, fmt.Errorf("failed to parse JSON response: %w. Response body: %s", err, string(bodyBytes))
	}

	for _, company := range userInfoResp.Companies {
		if company.ID == companyID {
			if company.EmployeeID != nil {
				return *company.EmployeeID, nil
			} else {
				return 0, fmt.Errorf("employee ID is null for company ID %d (user might not be registered as an employee in freee HR for this company)", companyID)
			}
		}
	}

	return 0, fmt.Errorf("company ID %d not found in user's associated companies", companyID)
}

// registerAttendanceForSingleDate 関数 (変更なし)
func registerAttendanceForSingleDate(token string, companyID, employeeID int, targetDate time.Time) error {
	client := &http.Client{Timeout: 15 * time.Second}

	dateStr := targetDate.Format("2006-01-02")

	payload := WorkRecordPayload{
		CompanyID: companyID,
		BreakRecords: []TimeRecord{
			{
				ClockInAt:  fmt.Sprintf("%s %s", dateStr, dailyBreakStart),
				ClockOutAt: fmt.Sprintf("%s %s", dateStr, dailyBreakEnd),
			},
		},
		WorkRecordSegments: []WorkRecordSegment{
			{
				ClockInAt:  fmt.Sprintf("%s %s", dateStr, dailyClockInTime),
				ClockOutAt: fmt.Sprintf("%s %s", dateStr, dailyClockOutTime),
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling payload for %s: %w", dateStr, err)
	}

	url := fmt.Sprintf("%s/api/v1/employees/%d/work_records/%s", freeeBaseURL, employeeID, dateStr)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("error creating PUT request for %s: %w", dateStr, err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")
	req.Header.Set("FREEE-VERSION", "2022-02-01")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending PUT request for %s: %w", dateStr, err)
	}
	defer resp.Body.Close()

	respBodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Printf("Warning: failed to read response body for %s: %v", dateStr, readErr)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API PUT request failed for date %s (Status: %s): %s", dateStr, resp.Status, string(respBodyBytes))
	}

	log.Printf("Successfully registered/updated attendance for %s (Status: %s)", dateStr, resp.Status)

	return nil
}
