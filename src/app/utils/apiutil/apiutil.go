package apiutil

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"app/utils/logger"
)

func GetBytesFromResponseBody(address string) ([]byte, error) {
	// get request
	logger.Infos(address)
	response, err := http.Get(address)
	if err != nil {
		fmt.Printf("error %+v\n", err)
		logger.Errors(err)
		return nil, err
	}
	defer response.Body.Close()

	// check status code
	logger.Infof("Status=%v", response.Status)
	if response.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("response.Status : %v", response.Status)
		logger.Errorf(errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	// read body bytes
	bytesBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logger.Errors(err)
	}

	return bytesBody, err
}

func GetJsonMapFromResponseBody(address string) (*map[string]interface{}, error) {
	// get bytes data from api response body
	bytesBody, err := GetBytesFromResponseBody(address)
	if err != nil {
		logger.Errors(err)
		return nil, err
	}

	// convert to map from body bytes data
	var mapBody map[string]interface{}
	err = json.Unmarshal([]byte(bytesBody), &mapBody)
	if err != nil {
		logger.Errors(err)
		return nil, err
	}

	return &mapBody, nil
}
