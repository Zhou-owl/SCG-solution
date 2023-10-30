package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

type TeamRequest struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

type Response struct {
	Name         string `json:"name"`
	MembersCount int    `json:"members_count"`
}

func GetTeamAndMembers(req TeamRequest) (Response, error) {
	return Response{
		Name:         req.Name,
		MembersCount: len(req.Members),
	}, nil
}

func main() {
	lambda.Start(GetTeamAndMembers)
}
