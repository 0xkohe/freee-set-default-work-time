# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A command-line tool written in Go that automatically registers daily work attendance records to the freee HR API for all days of a specified month and year. It sets fixed clock-in/out times and break times via API calls to freee人事労務 (freee HR system).

## Core Workflow

1. Accepts `-year` and `-month` command-line arguments to specify target month
2. Loads credentials (access token, company ID) from `.env` file or environment variables
3. Fetches the employee ID for the authenticated user via `/api/v1/users/me` endpoint
4. Iterates through each day of the specified month
5. For each day, sends a PUT request to `/api/v1/employees/{employee_id}/work_records/{date}` with:
   - Work record segments (clock-in and clock-out times)
   - Break records (break start and end times)

## Commands

### List companies and their IDs
```bash
go run main.go -list-companies
```
Displays all companies associated with your account, including their Company IDs and Employee IDs.

### Run the tool
```bash
go run main.go -year=2025 -month=3
```

### Build executable
```bash
go build -o freee-attendance-registrar main.go
```

### Run built executable
```bash
./freee-attendance-registrar -year=2025 -month=3
```

## Configuration

### Required Environment Variables
Set via `.env` file or shell environment:
- `TOKEN` - freee API access token with `write_company_employee_work_records` permission
- `COMPANY_ID` - Target company ID (numerical identifier in freee)

### Hardcoded Work Times
To change default work times, modify these variables in `main.go:65-68`:
```go
dailyClockInTime  = "11:00:00"
dailyClockOutTime = "21:00:00"
dailyBreakStart   = "12:00:00"
dailyBreakEnd     = "13:00:00"
```

## Architecture Notes

- **Single file architecture**: All logic contained in `main.go`
- **API Version**: Uses freee HR API version `2022-02-01` (set via `FREEE-VERSION` header)
- **API Base URL**: `https://api.freee.co.jp/hr`
- **HTTP Client**: Uses standard `net/http` with 10-15 second timeouts
- **Error Handling**: Fatal errors stop execution; continues on successful registration (200/201 status codes)
- **Dependencies**: Only external dependency is `github.com/joho/godotenv` for `.env` file loading
