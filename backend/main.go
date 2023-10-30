package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

const (
	host     = "scg-chr.cjwqh3bcqjqu.ap-southeast-1.rds.amazonaws.com"
	port     = "5432"
	user     = "master"
	password = "password"
	dbname   = "request"
)

type NewRequest struct {
	EmployeeID  string `json:"employee_id"`
	RequestType string `json:"request_type"`
}

type RequestSchema struct {
	ID           int    `json:"id"`
	RequestType  string `json:"request_type"`
	StepName     string `json:"step_name"`
	StepSequence int    `json:"step_sequence"`
}

type PendingRequest struct {
	RequestID   int       `json:"request_id"`
	EmployeeID  string    `json:"employee_id"`
	RequestType string    `json:"request_type"`
	StartTime   time.Time `json:"start_time"`
	CurrentStep string    `json:"current_step"`
}

func main() {
	e := echo.New()

	// Connect to PostgreSQL database
	connectionString := "user=" + user + " password=" + password + " dbname=" + dbname + " host=" + host + " port=" + port + " sslmode=disable"
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// API endpoints
	e.POST("/api/request/new", func(c echo.Context) error {
		// Parse request data
		var requestData NewRequest
		if err := c.Bind(&requestData); err != nil {
			return c.JSON(http.StatusBadRequest, "Invalid input")
		}

		// Fetch the matching request type and sequence 1 step from request_schema
		var requestSchema RequestSchema
		err := db.QueryRow("SELECT * FROM request_schema WHERE request_type=$1 AND step_sequence=1", requestData.RequestType).Scan(&requestSchema.ID, &requestSchema.RequestType, &requestSchema.StepName, &requestSchema.StepSequence)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusBadRequest, "Invalid request type")
			}
			return c.JSON(http.StatusInternalServerError, "Database error")
		}

		// Insert the new request into employee_requests
		startTime := time.Now()
		_, err = db.Exec("INSERT INTO employee_requests (employee_id, request_type, start_time, current_step) VALUES ($1, $2, $3, $4)", requestData.EmployeeID, requestData.RequestType, startTime, requestSchema.StepName)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to create new request")
		}

		return c.JSON(http.StatusCreated, "New request created")
	})

	e.GET("/api/request/pending", func(c echo.Context) error {

		// Get the employee ID from the query parameter
		currentEmployeeID := c.QueryParam("employee_id")
		if currentEmployeeID == "" {
			return c.JSON(http.StatusBadRequest, "Missing employee ID")
		}

		// Fetch all pending requests for the current user
		rows, err := db.Query("SELECT id, employee_id, request_type, start_time, current_step FROM employee_requests WHERE employee_id=$1 AND status='pending'", currentEmployeeID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Database error")
		}
		defer rows.Close()

		// Store the fetched requests in a slice
		var pendingRequests []PendingRequest
		for rows.Next() {
			var request PendingRequest
			err := rows.Scan(&request.RequestID, &request.EmployeeID, &request.RequestType, &request.StartTime, &request.CurrentStep)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, "Database error")
			}
			pendingRequests = append(pendingRequests, request)
		}

		return c.JSON(http.StatusOK, pendingRequests)
	})

	// Other API endpoints...

	// Start the server
	e.Start(":80")
}
