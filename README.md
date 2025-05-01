



# freee HR Bulk Attendance Registrar



## README (日本語)

# freee人事労務 勤怠一括登録ツール

指定された年月のすべての日に対して、freee人事労務 API を使用して日々の勤怠情報（固定の出勤/退勤時刻、休憩時間）を自動で登録するGo言語製コマンドラインツールです。

## 概要

このツールは、freee人事労務での日々の勤怠記録プロセスを簡略化します。特に、標準的な勤務時間と休憩時間が月を通して一貫している場合に便利です。コマンドライン引数で指定された年月の各日を反復処理し、freee人事労務APIエンドポイント `/api/v1/employees/{employee_id}/work_records/{date}` に対してPUTリクエストを送信し、その日の勤怠情報を作成または更新します。

## 特徴

* 指定された年月の全日に対し、固定の出勤・退勤・休憩時間を登録します。
* 設定情報（APIアクセストークン、会社ID）を環境変数または`.env`ファイルから安全に読み込みます。
* 対象の年と月を必須のコマンドライン引数として受け付けます。
* 実行中に進行状況に関するログを出力します。

## 前提条件

* **Go:** バージョン1.18以降推奨。[インストールガイド](https://go.dev/doc/install)
* **freee人事労務 APIアクセストークン:** freee Developersコンソールまたは承認プロセスを通じて、必要な権限（`write_company_employee_work_records`）を持つアクセストークンを取得する必要があります。
* **freeeの会社IDと従業員ID:** freee内での貴社の会社IDと、その会社に紐づく貴方自身の従業員IDが必要です。スクリプトは、提供された会社IDに基づいて`/users/me`エンドポイントを使用し、従業員IDを自動的に取得します。

## セットアップ

1.  **コードの取得:**
    * リポジトリをクローン: `git clone <repository-url>`
    * または、`main.go`ファイルを直接ダウンロードします。

2.  **ディレクトリへの移動:**
    ```bash
    cd <main.goが含まれるディレクトリ>
    ```

3.  **(任意) 実行ファイルのビルド:**
    ```bash
    go build -o freee-attendance-registrar main.go
    ```
    これにより実行ファイル（Linux/macOSでは `freee-attendance-registrar`、Windowsでは `freee-attendance-registrar.exe` など）が作成されます。

## 設定

このツールは、freee APIアクセストークンと対象の会社IDを必要とします。これらは以下の2つの方法で設定できます。

1.  **`.env` ファイル (推奨):**
    * スクリプト（`main.go`または実行ファイル）と同じディレクトリに `.env` という名前のファイルを作成します。
    * 以下の内容を追加し、プレースホルダーの値を実際の情報に置き換えます。
        ```dotenv
        TOKEN=ここにあなたのfreee_APIアクセストークン
        COMPANY_ID=ここにあなたの対象会社ID
        ```

2.  **環境変数:**
    * スクリプトを実行する前に、ターミナルセッションで直接環境変数を設定します。
        ```bash
        export TOKEN="ここにあなたのfreee_APIアクセストークン"
        export COMPANY_ID="ここにあなたの対象会社ID"
        # Windows コマンドプロンプトの場合:
        # set TOKEN=ここにあなたのfreee_APIアクセストークン
        # set COMPANY_ID=ここにあなたの対象会社ID
        # Windows PowerShell の場合:
        # $env:TOKEN="ここにあなたのfreee_APIアクセストークン"
        # $env:COMPANY_ID="ここにあなたの対象会社ID"
        ```

**注意:** 会社IDはfreee内での貴社を示す数値IDです。freee人事労務のURLに含まれているか、お手持ちのトークンで `/api/v1/users/me` エンドポイントを照会することで確認できる場合があります。

## 使い方

ターミナルからスクリプトを実行し、`-year` と `-month` フラグを使用して対象の年と月を指定します。

**`go run` を使用する場合:**

```bash
go run main.go -year=YYYY -month=MM
```
A command-line tool written in Go to automatically register daily work attendance records (fixed clock-in/out times and break times) to the freee HR API for all days of a specified month and year.

## Overview

This tool simplifies the process of recording daily attendance in freee HR, especially useful when standard working hours and breaks apply consistently across a month. It iterates through each day of the month provided via command-line arguments and sends a PUT request to the freee HR API endpoint `/api/v1/employees/{employee_id}/work_records/{date}` to create or update the work record for that day.

## Features

* Registers fixed clock-in, clock-out, and break times for every day of a given month.
* Reads configuration (API Access Token, Company ID) securely from environment variables or a `.env` file.
* Accepts the target year and month as mandatory command-line arguments.
* Provides informative logs during execution.

## Prerequisites

* **Go:** Version 1.18 or later recommended. [Installation Guide](https://go.dev/doc/install)
* **freee HR API Access Token:** You need to obtain an access token with the necessary permissions (`write_company_employee_work_records`) from the freee Developers console or through their authorization process.
* **freee Company ID and Employee ID:** You need your Company ID within freee and your specific Employee ID associated with that company in freee HR. The script automatically fetches your Employee ID using the `/users/me` endpoint based on the provided Company ID.

## Setup

1.  **Get the Code:**
    * Clone the repository: `git clone <repository-url>`
    * Or, download the `main.go` file directly.

2.  **Navigate to Directory:**
    ```bash
    cd <directory-containing-main.go>
    ```

3.  **(Optional) Build the Executable:**
    ```bash
    go build -o freee-attendance-registrar main.go
    ```
    This creates an executable file (e.g., `freee-attendance-registrar` on Linux/macOS, `freee-attendance-registrar.exe` on Windows).

## Configuration

The tool requires your freee API Access Token and Target Company ID. You can provide these in two ways:

1.  **`.env` File (Recommended):**
    * Create a file named `.env` in the same directory as the script (`main.go` or the executable).
    * Add the following lines, replacing the placeholder values with your actual credentials:
        ```dotenv
        TOKEN=YOUR_FREEE_API_ACCESS_TOKEN_HERE
        COMPANY_ID=YOUR_TARGET_COMPANY_ID_HERE
        ```

2.  **Environment Variables:**
    * Set the environment variables directly in your terminal session before running the script:
        ```bash
        export TOKEN="YOUR_FREEE_API_ACCESS_TOKEN_HERE"
        export COMPANY_ID="YOUR_TARGET_COMPANY_ID_HERE"
        # On Windows Command Prompt:
        # set TOKEN=YOUR_FREEE_API_ACCESS_TOKEN_HERE
        # set COMPANY_ID=YOUR_TARGET_COMPANY_ID_HERE
        # On Windows PowerShell:
        # $env:TOKEN="YOUR_FREEE_API_ACCESS_TOKEN_HERE"
        # $env:COMPANY_ID="YOUR_TARGET_COMPANY_ID_HERE"
        ```

**Note:** The Company ID is a numerical identifier for your company within freee. You might find it in the freee HR URL or by querying the `/api/v1/users/me` endpoint with your token.

## Usage

Run the script from your terminal, providing the target year and month using the `-year` and `-month` flags.

**Using `go run`:**

```bash
go run main.go -year=YYYY -month=MM
