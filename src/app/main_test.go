package main

import (
	"github.com/aws/aws-lambda-go/events"
	"testing"

	"app/utils/logger"
)

func TestHandlerSuccess(t *testing.T) {
	logger.LogInitialize(logger.Error, 25)
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "normal",
			args: args{key: "type", value: "main_summary:5ab47071-3651-457c-ae2b-bfb8fdbe1af1,main_summary:92f9ebcd-a3f1-4d5d-899b-d69214294a45,patients:5ab47071-3651-457c-ae2b-bfb8fdbe1af1,patients_summary:5ab47071-3651-457c-ae2b-bfb8fdbe1af1,inspection_persons:d4827176-d887-412a-9344-f84f161786a2,contacts:1b57f2c0-081e-4664-ba28-9cce56d0b314"},
		},
	}
	for _, tt := range tests {
		req := events.APIGatewayProxyRequest{
			QueryStringParameters: map[string]string{tt.args.key: tt.args.value},
		}
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler(req)
			if err != nil {
				t.Errorf(tt.name)
			}
		})
	}
}
