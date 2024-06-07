package smmstool

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var client = &http.Client{
	Timeout: 5 * time.Minute, // 设置全局客户端超时时间
}

type FormData interface {
	ContentType() string
	Body() (io.Reader, error)
}

// URLValuesWrapper 用于包装 url.Values
type URLValuesWrapper struct {
	data url.Values
}

func (v *URLValuesWrapper) ContentType() string {
	return "application/x-www-form-urlencoded"
}

func (v *URLValuesWrapper) Body() (io.Reader, error) {
	return strings.NewReader(v.data.Encode()), nil
}

// BufferWrapper 用于包装 bytes.Buffer
type BufferWrapper struct {
	data        *bytes.Buffer
	contentType string
}

func (b *BufferWrapper) ContentType() string {
	return b.contentType
}

func (b *BufferWrapper) Body() (io.Reader, error) {
	return b.data, nil
}

// createRequest 创建 HTTP 请求
func createRequest(ctx context.Context, urlStr, token string, formData FormData) (*http.Request, error) {
	body, err := formData.Body()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", urlStr, body)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Add("Authorization", token)
	}
	req.Header.Set("Content-Type", formData.ContentType())
	return req, nil
}

// sendRequestWithRetry 发送 HTTP 请求，并在失败时重试
func sendRequestWithRetry(ctx context.Context, urlStr, token string, formData FormData, retries int) (*http.Response, error) {
	var lastErr error

	for i := 0; i <= retries; i++ {
		req, err := createRequest(ctx, urlStr, token, formData)
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				fmt.Printf("Attempt %d: request timed out (context deadline exceeded)\n", i+1)
				lastErr = err
				continue
			} else if err, ok := err.(net.Error); ok && err.Timeout() {
				fmt.Printf("Attempt %d: request timed out (net timeout)\n", i+1)
				lastErr = err
				continue
			} else {
				return nil, fmt.Errorf("error sending request: %v", err)
			}
		}

		if resp.StatusCode == http.StatusOK {
			return resp, nil
		}

		resp.Body.Close()
		fmt.Printf("Attempt %d/%d failed: %s\n", i+1, retries, http.StatusText(resp.StatusCode))
		lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		time.Sleep(2 * time.Second) // Optional: adding sleep between retries
	}

	return nil, fmt.Errorf("all retries failed: %v", lastErr)
}
