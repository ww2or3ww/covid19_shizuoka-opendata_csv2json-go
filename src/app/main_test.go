package main

import (
	"os"
	"strconv"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/joho/godotenv"

	"app/utils/logger"
)

func init() {
	if os.Getenv("LOG_LEVEL") == "" {
		godotenv.Load(".env")
	}
	logLv := logger.Error
	envLogLv := os.Getenv("LOG_LEVEL")
	if envLogLv != "" {
		n, _ := strconv.Atoi(envLogLv)
		logLv = logger.LogLv(n)
	}
	logger.LogInitialize(logLv, 25)
}

func TestHandlerSuccess(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name       string
		args       args
		statusCode int
	}{
		{
			name:       "normal",
			args:       args{key: "type", value: "main_summary:5ab47071-3651-457c-ae2b-bfb8fdbe1af1,main_summary:92f9ebcd-a3f1-4d5d-899b-d69214294a45,patients:5ab47071-3651-457c-ae2b-bfb8fdbe1af1,patients_summary:5ab47071-3651-457c-ae2b-bfb8fdbe1af1,inspection_persons:d4827176-d887-412a-9344-f84f161786a2,contacts:1b57f2c0-081e-4664-ba28-9cce56d0b314"},
			statusCode: 200,
		},
	}
	for _, tt := range tests {
		req := events.APIGatewayProxyRequest{
			QueryStringParameters: map[string]string{tt.args.key: tt.args.value},
		}
		t.Run(tt.name, func(t *testing.T) {
			ret, err := handler(req)
			if err != nil {
				t.Errorf(tt.name)
			} else if ret.StatusCode != tt.statusCode {
				t.Errorf(tt.name)
			}
		})
	}
}
