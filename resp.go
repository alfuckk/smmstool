package smmstool

import (
	"encoding/json"
	"fmt"
)

// 定义通用的响应结构体
type Response struct {
	Success   bool        `json:"success"`
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"` // Data 类型为 interface{}
	RequestID string      `json:"RequestId"`
}

// 泛型函数，处理不同类型的 Data 字段
func processResponse[T any](jsonData string, dataField *T) (code string, err error) {
	var response Response

	// 解析 JSON 数据
	err = json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		return response.Code, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	// 再次解析 Data 字段为指定的类型
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return response.Code, fmt.Errorf("error marshaling Data field: %v", err)
	}

	err = json.Unmarshal(dataBytes, &dataField)
	if err != nil {
		return response.Code, fmt.Errorf("error unmarshaling Data field: %v", err)
	}

	return response.Code, nil
}
