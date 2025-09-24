package main

import (
	"bytes"
	"fmt"
	"log_converter/dto/request"
	"log_converter/dto/response"
	"log_converter/service"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func authenticate(c *gin.Context) {
	var req request.AuthenticateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	baseURL := os.Getenv("BASE_URL")
	loginResp, err := service.Login(baseURL, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	projects, err := service.ProjectList(baseURL, strconv.Itoa(loginResp.UserInfo.EmployeeID), loginResp.IdToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"projects": projects, "token": loginResp.IdToken, "employeeId": loginResp.UserInfo.EmployeeID})

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
	var acts []response.Activity
	for _, month := range req.Months {
		logResp, err := service.FetchLogActivity(baseURL, req.Token, req.EmployeeID, month, req.Year)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project: '" + req.ProjectName + "' not found on selected month and year"})
		return
	}
	reqMaxDuration := req.RandomizeLog.MaxDuration
	if reqMaxDuration != 0 {
		maxDuration = reqMaxDuration
	}
	exportParam := request.ExportParam{
		ProjectFilter:       req.ProjectName,
		IsRandomizeDuration: req.RandomizeLog.IsRandom,
		MinDuration:         req.RandomizeLog.MinDuration,
		MaxDuration:         maxDuration,
	}

	f, err := service.ConvertToExcel(acts, exportParam)
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
		// return
	}
	router := gin.Default()
	router.Static("/static", "./static")
	router.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})
	router.POST("/api/authenticate", authenticate)
	router.POST("/api/convert", handleConvertToExcel)

	router.Run(":" + os.Getenv("PORT"))
}
