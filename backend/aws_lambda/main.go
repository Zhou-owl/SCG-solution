package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
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

type ResponseMessage struct {
	Message string `json:"message"`
}

func HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Connect to PostgreSQL database
	connectionString := "user=" + user + " password=" + password + " dbname=" + dbname + " host=" + host + " port=" + port + " sslmode=disable"
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return serverError(err)
	}
	defer db.Close()

	var response events.APIGatewayProxyResponse

	switch req.HTTPMethod {
	case "POST":
		if req.Path == "/api/request/new" {
			response, err = createNewRequest(db, req)
		}
	case "GET":
		if req.Path == "/api/request/pending" {
			response, err = getPendingRequests(db, req)
		}
	default:
		response, err = clientError(http.StatusNotFound)
	}

	// Add CORS headers to the response
	response.Headers = map[string]string{
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Headers":     "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
		"Access-Control-Allow-Methods":     "OPTIONS,GET,POST",
		"Access-Control-Allow-Credentials": "true",
	}

	return response, err
}

func createNewRequest(db *sql.DB, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse request data
	var requestData NewRequest
	err := json.Unmarshal([]byte(req.Body), &requestData)
	if err != nil {
		return clientError(http.StatusBadRequest)
	}

	// Fetch the matching request type and sequence 1 step from request_schema
	var requestSchema RequestSchema
	err = db.QueryRow("SELECT * FROM request_schema WHERE request_type=$1 AND step_sequence=1", requestData.RequestType).Scan(&requestSchema.ID, &requestSchema.RequestType, &requestSchema.StepName, &requestSchema.StepSequence)
	if err != nil {
		if err == sql.ErrNoRows {
			return clientError(http.StatusBadRequest)
		}
		return serverError(err)
	}

	// Insert the new request into employee_requests
	startTime := time.Now()
	_, err = db.Exec("INSERT INTO employee_requests (employee_id, request_type, start_time, current_step) VALUES ($1, $2, $3, $4)", requestData.EmployeeID, requestData.RequestType, startTime, requestSchema.StepName)
	if err != nil {
		return serverError(err)
	}

	if err != nil {
		return serverError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
	}, nil
}

func getPendingRequests(db *sql.DB, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get the employee ID from the query parameter
	currentEmployeeID := req.QueryStringParameters["employee_id"]
	if currentEmployeeID == "" {
		return clientError(http.StatusBadRequest)
	}

	// Fetch all pending requests for the current user
	rows, err := db.Query("SELECT id, employee_id, request_type, start_time, current_step FROM employee_requests WHERE employee_id=$1 AND status='pending'", currentEmployeeID)
	if err != nil {
		return serverError(err)
	}
	defer rows.Close()

	// Store the fetched requests in a slice
	var pendingRequests []PendingRequest
	for rows.Next() {
		var request PendingRequest
		err := rows.Scan(&request.RequestID, &request.EmployeeID, &request.RequestType, &request.StartTime, &request.CurrentStep)
		if err != nil {
			return serverError(err)
		}
		pendingRequests = append(pendingRequests, request)
	}

	responseBody, err := json.Marshal(pendingRequests)
	if err != nil {
		return serverError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(responseBody),
	}, nil
}

func serverError(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
	}, err
}

func clientError(status int) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       http.StatusText(status),
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
