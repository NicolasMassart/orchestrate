package client

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/infra/api"
	"github.com/traefik/traefik/v2/pkg/log"
)

const (
	cannotReadResponseBody = "failed to read response body"
	invalidResponseBody    = "failed to decode response body"
)

type HTTPErr struct {
	statusCode int
	err        error
}

func NewHTTPErr(err error, code int) *HTTPErr {
	return &HTTPErr{
		err:        err,
		statusCode: code,
	}
}

func (m *HTTPErr) Error() string {
	return m.err.Error()
}

func (m *HTTPErr) Code() int {
	return m.statusCode
}

func parseResponse(ctx context.Context, response *http.Response, resp interface{}) error {
	if response.StatusCode == http.StatusAccepted || response.StatusCode == http.StatusOK {
		if resp == nil {
			return nil
		}

		if err := json.NewDecoder(response.Body).Decode(resp); err != nil {
			log.FromContext(ctx).WithError(err).Error(invalidResponseBody)
			return errors.ServiceConnectionError(invalidResponseBody)
		}

		return nil
	}

	// Read body
	respMsg, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.FromContext(ctx).WithError(err).Error(cannotReadResponseBody)
		return errors.ServiceConnectionError(cannotReadResponseBody)
	}

	if string(respMsg) != "" {
		errResp := api.ErrorResponse{}
		if err = json.Unmarshal(respMsg, &errResp); err == nil {
			return NewHTTPErr(errors.Errorf(errResp.Code, errResp.Message), response.StatusCode)
		}
	}

	return parseResponseError(response.StatusCode, string(respMsg))
}

func parseResponseError(statusCode int, errMsg string) error {
	switch statusCode {
	case http.StatusBadRequest:
		if errMsg == "" {
			errMsg = "invalid request data"
		}
		return NewHTTPErr(errors.InvalidFormatError(errMsg), statusCode)
	case http.StatusConflict:
		if errMsg == "" {
			errMsg = "invalid data message"
		}
		return NewHTTPErr(errors.ConflictedError(errMsg), statusCode)
	case http.StatusNotFound:
		if errMsg == "" {
			errMsg = "cannot find entity"
		}
		return NewHTTPErr(errors.NotFoundError(errMsg), statusCode)
	case http.StatusUnauthorized:
		if errMsg == "" {
			errMsg = "not authorized"
		}
		return NewHTTPErr(errors.UnauthorizedError(errMsg), statusCode)
	case http.StatusUnprocessableEntity:
		if errMsg == "" {
			errMsg = "invalid request format"
		}
		return NewHTTPErr(errors.InvalidParameterError(errMsg), statusCode)
	default:
		if errMsg == "" {
			errMsg = "server error"
		}
		return NewHTTPErr(errors.ServiceConnectionError(errMsg), statusCode)
	}
}

func parseStringResponse(ctx context.Context, response *http.Response) (string, error) {
	if response.StatusCode != http.StatusOK {
		errResp := api.ErrorResponse{}
		if err := json.NewDecoder(response.Body).Decode(&errResp); err != nil {
			log.FromContext(ctx).WithError(err).Error(invalidResponseBody)
			return "", errors.ServiceConnectionError(invalidResponseBody)
		}

		return "", NewHTTPErr(errors.Errorf(errResp.Code, errResp.Message), response.StatusCode)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.FromContext(ctx).WithError(err).Error(invalidResponseBody)
		return "", errors.ServiceConnectionError(invalidResponseBody)
	}

	return string(responseData), nil
}

func ParseEmptyBodyResponse(ctx context.Context, response *http.Response) error {
	if response.StatusCode != http.StatusNoContent && response.StatusCode != http.StatusAccepted {
		errResp := api.ErrorResponse{}
		if err := json.NewDecoder(response.Body).Decode(&errResp); err != nil {
			log.FromContext(ctx).WithError(err).Error(invalidResponseBody)
			return errors.ServiceConnectionError(invalidResponseBody)
		}

		return NewHTTPErr(errors.Errorf(errResp.Code, errResp.Message), response.StatusCode)
	}

	return nil
}
