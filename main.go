package main

import (
	"bytes"
	"encoding/json"
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

// UserInfoResponse maps the relevant parts of the /users/me response
type UserInfoResponse struct {
	ID        int           `json:"id"`
	Companies []CompanyInfo `json:"companies"`
}

type UserInfo struct {
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

// WorkRecordPayload maps the PUT /work_records request body
type WorkRecordPayload struct {
	CompanyID          int                 `json:"company_id"`
	BreakRecords       []TimeRecord        `json:"break_records"`
	WorkRecordSegments []WorkRecordSegment `json:"work_record_segments"`
	// Add Note string `json:"note,omitempty"` // Optional: Add if you need notes
}

// TimeRecord maps the clock_in_at/clock_out_at structure for breaks
type TimeRecord struct {
	ClockInAt  string `json:"clock_in_at"`  // Format: "YYYY-MM-DD HH:MM:SS"
	ClockOutAt string `json:"clock_out_at"` // Format: "YYYY-MM-DD HH:MM:SS"
}

// WorkRecordSegment maps the clock_in_at/clock_out_at structure for work segments
type WorkRecordSegment struct {
	ClockInAt  string `json:"clock_in_at"`  // Format: "YYYY-MM-DD HH:MM:SS"
	ClockOutAt string `json:"clock_out_at"` // Format: "YYYY-MM-DD HH:MM:SS"
}

// --- Configuration ---

var (
	accessToken     = "3d5b3c4d18543a76994a1548f271fa4e8985b135217d87b2a988c84df17355e2"
	targetCompanyId = 1908354 // <--- REPLACE WITH YOUR TARGET COMPANY ID
	targetYear      = 2025    // <--- SET TARGET YEAR
	targetMonth     = 5       // <--- SET TARGET MONTH (May)
	targetDay       = 1       // <--- SET TARGET DAY (1st)

	// Define the daily times (変更なし)
	dailyClockInTime  = "11:00:00"
	dailyClockOutTime = "21:00:00"
	dailyBreakStart   = "12:00:00"
	dailyBreakEnd     = "13:00:00"
)

// --- Main Logic ---

func main() {
	err := godotenv.Load()

	accessToken = os.Getenv("TOKEN")
	tcids := os.Getenv("COMPANY_ID")
	targetCompanyId, err = strconv.Atoi(tcids)
	if err != nil {
		log.Fatal("Error converting COMPANY_ID to integer")
	}

	if accessToken == "YOUR_ACCESS_TOKEN" {
		log.Fatal("Error: Please replace 'YOUR_ACCESS_TOKEN' with your actual freee API access token.")
	}
	if targetCompanyId == 0 {
		log.Fatal("Error: Please set your targetCompanyId.")
	}

	log.Println("Fetching employee ID...")
	employeeID, err := getEmployeeID(accessToken, targetCompanyId)
	if err != nil {
		log.Fatalf("Error getting employee ID: %v", err)
	}
	log.Printf("Found Employee ID: %d for Company ID: %d\n", employeeID, targetCompanyId)

	// 特定の日付を指定
	dateToRegister := time.Date(targetYear, time.Month(targetMonth), targetDay, 0, 0, 0, 0, time.UTC)
	dateStr := dateToRegister.Format("2006-01-02")

	log.Printf("Registering attendance for specific date: %s...\n", dateStr)
	err = registerAttendanceForSingleDate(accessToken, targetCompanyId, employeeID, dateToRegister)
	if err != nil {
		log.Fatalf("Error registering attendance for %s: %v", dateStr, err)
	}

	log.Printf("Attendance registration process completed for %s.", dateStr)
}

// --- API Interaction Functions ---

// getEmployeeID fetches the user info and extracts the employee ID for the target company (変更なし)
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
		return 0, fmt.Errorf("failed to parse JSON response: %w", err)
	}
	fmt.Println("%v", userInfoResp)

	// Find the employee ID for the specified company
	for _, company := range userInfoResp.Companies {
		fmt.Println("kohe" + company.Name)
		fmt.Println("Company ID:", company.ID, "Employee ID:", company.EmployeeID)
		if company.ID == companyID {
			if company.EmployeeID != nil {
				return *company.EmployeeID, nil
			} else {
				return 0, fmt.Errorf("employee ID not found for company ID %d (user might not be an employee)", companyID)
			}
		}
	}

	return 0, fmt.Errorf("company ID %d not found in user's associated companies", companyID)
}

// registerAttendanceForSingleDate registers attendance for the specified single date
func registerAttendanceForSingleDate(token string, companyID, employeeID int, targetDate time.Time) error {
	client := &http.Client{Timeout: 15 * time.Second} // Timeout for PUT

	dateStr := targetDate.Format("2006-01-02") // Format YYYY-MM-DD
	log.Printf("Processing date: %s\n", dateStr)

	// Construct the payload for this specific day
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

	fmt.Println("Payload for %s: %+v\n", dateStr, payload)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling payload for %s: %w", dateStr, err)
	}

	// Construct the API endpoint URL
	// URL: /api/v1/employees/{employee_id}/work_records/{date}
	url := fmt.Sprintf("%s/api/v1/employees/%d/work_records/%s", freeeBaseURL, employeeID, dateStr)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("error creating PUT request for %s: %w", dateStr, err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending PUT request for %s: %w", dateStr, err)
	}
	defer resp.Body.Close()

	respBodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Printf("Warning: failed to read response body for %s: %v", dateStr, readErr)
		// Continue even if reading body fails, but check status code
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated { // Check for success (200 OK or 201 Created)
		return fmt.Errorf("API PUT request failed for date %s (Status: %s): %s", dateStr, resp.Status, string(respBodyBytes))
	}

	log.Printf("Successfully registered/updated attendance for %s (Status: %s)", dateStr, resp.Status)
	// Optional: Parse the response body if needed: log.Printf("Response body: %s", string(respBodyBytes))

	return nil // Indicate success for this date
}
