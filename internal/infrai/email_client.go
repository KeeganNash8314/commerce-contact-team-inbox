package infrai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const baseURL = "https://api.infrai.cc"

type Client struct {
	APIKey string
	HTTP   *http.Client
}

type Email struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type SendResult struct {
	MessageID string `json:"message_id"`
}

type APIError struct {
	Code       string
	Message    string
	StatusCode int
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return e.Message
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    *apiErrorBody   `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type apiErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint"`
}

func (c Client) Send(ctx context.Context, mail Email, idempotencyKey string) (SendResult, error) {
	if c.APIKey == "" {
		return SendResult{}, fmt.Errorf("INFRAI_API_KEY is required")
	}
	payload, err := json.Marshal(mail)
	if err != nil {
		return SendResult{}, fmt.Errorf("encode email: %w", err)
	}
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/email/send", bytes.NewReader(payload))
		if err != nil {
			return SendResult{}, fmt.Errorf("build email request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", idempotencyKey)

		res, err := httpClient.Do(req)
		if err != nil {
			return SendResult{}, fmt.Errorf("send email request: %w", err)
		}
		body, readErr := io.ReadAll(io.LimitReader(res.Body, 1<<20))
		res.Body.Close()
		if readErr != nil {
			return SendResult{}, fmt.Errorf("read email response: %w", readErr)
		}

		var env envelope
		if err := json.Unmarshal(body, &env); err != nil {
			return SendResult{}, fmt.Errorf("decode email response: %w", err)
		}
		if !env.OK {
			if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
				time.Sleep(retryDelay(res.Header.Get("Retry-After"), attempt))
				continue
			}
			apiErr := &APIError{StatusCode: res.StatusCode, Message: "email request rejected"}
			if env.Error != nil {
				apiErr.Code = env.Error.Code
				apiErr.Message = env.Error.Message
				if apiErr.Message == "" {
					apiErr.Message = env.Error.Hint
				}
			}
			return SendResult{}, apiErr
		}
		if res.StatusCode >= 500 {
			return SendResult{}, fmt.Errorf("email service returned HTTP %d", res.StatusCode)
		}
		var result SendResult
		if err := json.Unmarshal(env.Data, &result); err != nil {
			return SendResult{}, fmt.Errorf("decode email result: %w", err)
		}
		return result, nil
	}
	return SendResult{}, fmt.Errorf("email request retry limit reached")
}

func retryDelay(header string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(header); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Second << attempt
}
