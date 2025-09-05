package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log_cvt/dto/request"
	"log_cvt/dto/response"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/xuri/excelize/v2"
)

type ExportParam struct {
	projectFilter       string
	fileName            string
	isRandomizeDuration bool
	minDuration         int
	maxDuration         int
}

func login(baseURL, username, password string) (*response.LoginResponse, error) {
	loginReq := request.LoginRequest{
		Username: username,
		Password: password,
	}
	body, _ := json.Marshal(loginReq)

	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Post error")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed, status: %s", resp.Status)
	}

	var loginResp response.LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, err
	}

	return &loginResp, nil
}

// Fetch log activity
func fetchLogActivity(baseURL, token string, idEmployee int, month int, year int) (*response.LogActResponse, error) {

	url := baseURL + "/log-act-detail-non-aj/table?sort=date|asc&idEmployee=" + strconv.Itoa(idEmployee) + "&months=" + strconv.Itoa(month) + "&years=" + strconv.Itoa(year)

	// Create request
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch failed, status: %s, body: %s", resp.Status, string(body))
	}

	var logResp response.LogActResponse
	if err := json.Unmarshal(body, &logResp); err != nil {
		return nil, err
	}
	return &logResp, nil
}

func exportToExcel(activities []response.Activity, param ExportParam) error {
	projectFilter := param.projectFilter
	fileName := param.fileName + ".xlsx"
	isRandomizeDuration := param.isRandomizeDuration
	minDuration := param.minDuration
	maxDuration := param.maxDuration

	f := excelize.NewFile()
	sheet := "Sheet1"
	index, err := f.NewSheet(sheet)
	if err != nil {
		return err
	}

	layout := "02-01-2006"
	grouped := make(map[string][]response.Activity)

	for _, a := range activities {
		if a.ProjectName != projectFilter {
			continue
		}
		t, err := time.Parse(layout, a.DateString)
		if err != nil {
			return fmt.Errorf("invalid date format: %s", a.DateString)
		}
		key := t.Format("2006-01")
		grouped[key] = append(grouped[key], a)
	}

	var keys []string
	for k := range grouped {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	row := 1
	for _, k := range keys {
		t, _ := time.Parse("2006-01", k)
		monthName := t.Format("January 2006")

		indonesianMonths := map[time.Month]string{
			time.January:   "Januari",
			time.February:  "Februari",
			time.March:     "Maret",
			time.April:     "April",
			time.May:       "Mei",
			time.June:      "Juni",
			time.July:      "Juli",
			time.August:    "Agustus",
			time.September: "September",
			time.October:   "Oktober",
			time.November:  "November",
			time.December:  "Desember",
		}
		monthName = indonesianMonths[t.Month()]

		if row > 1 {
			row++
		}

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), monthName)
		row++

		headers := []string{"Tanggal", "Durasi (Jam)", "Kegiatan"}
		for col, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(col+1, row)
			f.SetCellValue(sheet, cell, h)
		}
		row++

		acts := grouped[k]

		for _, a := range acts {
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), a.DateString)
			duration := a.Duration
			if isRandomizeDuration {
				randomNumber := rand.Intn(maxDuration-minDuration+1) + minDuration
				duration = randomNumber
			}

			f.SetCellValue(sheet, fmt.Sprintf("B%d", row), duration)
			f.SetCellValue(sheet, fmt.Sprintf("C%d", row), a.ActivityDetail)
			row++
		}
	}

	f.SetActiveSheet(index)

	if err := f.SaveAs(fileName); err != nil {
		return err
	}
	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	baseURL := os.Getenv("BASE_URL")
	username := os.Getenv("USERNAME_ESS")
	password := os.Getenv("PASSWORD_ESS")
	monthsExported := os.Getenv("MONTHS_EXPORTED")
	var months []int
	if err := json.Unmarshal([]byte(monthsExported), &months); err != nil {
		log.Fatal(err)
	}
	year, err := strconv.Atoi(os.Getenv("YEAR_EXPORTED"))
	if err != nil {
		log.Fatal(err)
	}
	projectName := os.Getenv("PROJECT_NAME_EXPORTED")
	resultFileName := os.Getenv("RESULT_FILE_NAME")
	randomizeDuration, err := strconv.ParseBool(os.Getenv("RANDOMIZE_DURATION"))
	if err != nil {
		log.Fatal(err)
	}
	minDuration, err := strconv.Atoi(os.Getenv("MIN_DURATION"))
	if err != nil {
		log.Fatal(err)
	}

	// Step 1: Login
	loginResp, err := login(baseURL, username, password)
	if err != nil {
		log.Fatal("Login error:", err)
	}
	fmt.Println("Login successful, token:", loginResp.IdToken)
	fmt.Printf("User: %+v\n", loginResp.UserInfo)

	var acts []response.Activity
	for _, month := range months {
		logResp, err := fetchLogActivity(baseURL, loginResp.IdToken, loginResp.UserInfo.EmployeeID, month, year)
		if err != nil {
			log.Fatal("Fetch error:", err)
		}

		acts = append(acts, logResp.Data...)
	}

	maxDuration := 0
	for _, a := range acts {
		maxDuration = a.Duration
		fmt.Printf("Tanggal: %s, Project: %s, Duration: %d, Activity: %s\n",
			a.DateString, a.ProjectName, a.Duration, a.ActivityDetail)
	}
	envMaxDuration := os.Getenv("MAX_DURATION")
	if envMaxDuration != "" {

		maxDuration, err = strconv.Atoi(envMaxDuration)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("maxdur", maxDuration)
	exportParam := ExportParam{
		projectFilter:       projectName,
		fileName:            resultFileName,
		isRandomizeDuration: randomizeDuration,
		minDuration:         minDuration,
		maxDuration:         maxDuration,
	}

	if err := exportToExcel(acts, exportParam); err != nil {
		log.Fatal("Excel export error:", err)
	}
	fmt.Println("Excel file created: " + resultFileName + ".xlsx")
}
