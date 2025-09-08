package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log_converter/dto/request"
	"log_converter/dto/response"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/xuri/excelize/v2"
)

type ExportParam struct {
	projectFilter       string
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

func fetchLogActivity(baseURL, token string, idEmployee int, month int, year int) (*response.LogActResponse, error) {
	url := baseURL + "/log-act-detail-non-aj/table?sort=date|asc&idEmployee=" + strconv.Itoa(idEmployee) + "&months=" + strconv.Itoa(month) + "&years=" + strconv.Itoa(year)

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

func convertToExcel(activities []response.Activity, param ExportParam) (*excelize.File, error) {
	projectFilter := param.projectFilter
	isRandomizeDuration := param.isRandomizeDuration
	minDuration := param.minDuration
	maxDuration := param.maxDuration

	f := excelize.NewFile()
	sheet := "Sheet1"
	index, err := f.NewSheet(sheet)
	if err != nil {
		return f, err
	}
	defaultFont := &excelize.Font{
		Family: "Times New Roman",
		Size:   12,
	}
	defaultStyle, err := f.NewStyle(&excelize.Style{
		Font: defaultFont,
	})
	if err != nil {
		return f, err
	}

	f.SetCellStyle(sheet, "A1", "Z1000", defaultStyle)

	layout := "02-01-2006"
	grouped := make(map[string][]response.Activity)

	for _, a := range activities {
		if a.ProjectName != projectFilter {
			continue
		}
		t, err := time.Parse(layout, a.DateString)
		if err != nil {
			return f, fmt.Errorf("invalid date format: %s", a.DateString)
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
		indonesianMonths := map[time.Month]string{
			time.January:   "JANUARI",
			time.February:  "FEBRUARI",
			time.March:     "MARET",
			time.April:     "APRIL",
			time.May:       "MEI",
			time.June:      "JUNI",
			time.July:      "JULI",
			time.August:    "AGUSTUS",
			time.September: "SEPTEMBER",
			time.October:   "OKTOBER",
			time.November:  "NOVEMBER",
			time.December:  "DESEMBER",
		}
		monthName := indonesianMonths[t.Month()]

		if row > 1 {
			row++
		}
		boldStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Bold:   true,
				Family: defaultFont.Family,
				Size:   defaultFont.Size,
			},
		})

		monthCell := fmt.Sprintf("A%d", row)
		f.SetCellValue(sheet, monthCell, monthName)
		f.SetCellStyle(sheet, monthCell, monthCell, boldStyle)
		row++

		headers := []string{"Tanggal", "Durasi (Jam)", "Kegiatan"}
		headerStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Bold:   true,
				Family: defaultFont.Family,
				Size:   defaultFont.Size,
			},
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
			},
			Alignment: &excelize.Alignment{
				Horizontal: "center",
				Vertical:   "center",
			},
		})
		for col, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(col+1, row)
			f.SetCellValue(sheet, cell, h)
			f.SetCellStyle(sheet, cell, cell, headerStyle)

			f.SetColWidth(sheet, cell[:1], cell[:1], 15)
		}
		row++

		acts := grouped[k]
		durationStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Family: defaultFont.Family,
				Size:   defaultFont.Size,
			},
			Alignment: &excelize.Alignment{
				Horizontal: "center",
				Vertical:   "center",
			},
		})

		for _, a := range acts {
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), a.DateString)
			duration := a.Duration
			if isRandomizeDuration {
				randomNumber := rand.Intn(maxDuration-minDuration+1) + minDuration
				duration = randomNumber
			}
			durationCell := fmt.Sprintf("B%d", row)

			f.SetCellValue(sheet, durationCell, duration)
			f.SetCellStyle(sheet, durationCell, durationCell, durationStyle)
			f.SetCellValue(sheet, fmt.Sprintf("C%d", row), a.ActivityDetail)
			row++
		}
	}

	f.SetActiveSheet(index)
	return f, nil
}

func handleConvertToExcel(c *gin.Context) {
	var req request.LogConverterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	baseURL := os.Getenv("BASE_URL")
	loginResp, err := login(baseURL, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var acts []response.Activity
	for _, month := range req.Months {
		logResp, err := fetchLogActivity(baseURL, loginResp.IdToken, loginResp.UserInfo.EmployeeID, month, req.Year)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		acts = append(acts, logResp.Data...)
	}

	maxDuration := 0
	found := false
	for _, a := range acts {
		maxDuration = a.Duration
		if a.ProjectName == req.ProjectName {
			found = true
			break
		}
	}
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project name " + req.ProjectName + " not found"})
		return
	}
	envMaxDuration := req.RandomizeLog.MaxDuration
	if envMaxDuration != 0 {
		maxDuration = envMaxDuration
	}
	exportParam := ExportParam{
		projectFilter:       req.ProjectName,
		isRandomizeDuration: req.RandomizeLog.IsRandom,
		minDuration:         req.RandomizeLog.MinDuration,
		maxDuration:         maxDuration,
	}

	f, err := convertToExcel(acts, exportParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error convert to excel": err.Error()})
		return
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate excel"})
		return
	}
	currentTime := time.Now()
	timeString := currentTime.Format("2006-01-02_15-04-05")
	fileName := `timesheet_` + timeString + `.xlsx`

	c.Header("Content-Disposition", `attachment; filename=`+fileName)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("error getting env", err)
		return
	}
	router := gin.Default()
	router.POST("/convert", handleConvertToExcel)

	router.Run(":" + os.Getenv("PORT"))
}
