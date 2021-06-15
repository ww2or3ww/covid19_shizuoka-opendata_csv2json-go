package csv2json

import (
	"testing"
)

const defApiAddress = "https://opendata.pref.shizuoka.jp/api/package_show"

func TestProcess(t *testing.T) {
	type args struct {
		apiAddress  string
		queryStrPrm string
	}
	tests := []struct {
		name     string
		args     args
		hasError bool
	}{
		{
			name: "normal",
			args: args{
				apiAddress:  defApiAddress,
				queryStrPrm: "main_summary:5ab47071-3651-457c-ae2b-bfb8fdbe1af1",
			},
			hasError: false,
		},
		{
			name: "hasError : invalid qyery param",
			args: args{
				apiAddress:  defApiAddress,
				queryStrPrm: "main_summary",
			},
			hasError: true,
		},
		{
			name: "hasError : invalid address",
			args: args{
				apiAddress:  "xxx",
				queryStrPrm: "main_summary:5ab47071-3651-457c-ae2b-bfb8fdbe1af1",
			},
			hasError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapResult := Process(tt.args.apiAddress, tt.args.queryStrPrm)
			if mapResult == nil {
				t.Errorf(tt.name)
			} else if (*mapResult)["hasError"] != tt.hasError {
				t.Errorf(tt.name)
			}
		})
	}
}
