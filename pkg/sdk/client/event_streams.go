package client

import (
	"context"
	"fmt"
	"strings"

	clientutils "github.com/consensys/orchestrate/pkg/toolkit/app/http/client-utils"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
)

func (c *HTTPClient) GetEventStream(ctx context.Context, uuid string) (*types.EventStreamResponse, error) {
	reqURL := fmt.Sprintf("%v/eventstreams/%s", c.config.URL, uuid)
	resp := &types.EventStreamResponse{}

	err := callWithBackOff(ctx, c.config.backOff, func() error {
		response, err := clientutils.GetRequest(ctx, c.client, reqURL)
		if err != nil {
			return err
		}

		defer clientutils.CloseResponse(response)
		return parseResponse(ctx, response, resp)
	})

	return resp, err
}

func (c *HTTPClient) SearchEventStreams(ctx context.Context, filters *entities.EventStreamFilters) ([]*types.EventStreamResponse, error) {
	reqURL := fmt.Sprintf("%v/eventstreams", c.config.URL)
	var resp []*types.EventStreamResponse

	var qParams []string
	if len(filters.Names) > 0 {
		qParams = append(qParams, "names="+strings.Join(filters.Names, ","))
	}

	if filters.TenantID != "" {
		qParams = append(qParams, "tenant_id="+filters.TenantID)
	}

	if filters.ChainUUID != "" {
		qParams = append(qParams, "chain_uuid="+filters.ChainUUID)
	}

	if len(qParams) > 0 {
		reqURL = reqURL + "?" + strings.Join(qParams, "&")
	}

	err := callWithBackOff(ctx, c.config.backOff, func() error {
		response, err := clientutils.GetRequest(ctx, c.client, reqURL)
		if err != nil {
			return err
		}
		defer clientutils.CloseResponse(response)
		return parseResponse(ctx, response, &resp)
	})

	return resp, err
}

func (c *HTTPClient) CreateEventStream(ctx context.Context, request *types.CreateEventStreamRequest) (*types.EventStreamResponse, error) {
	reqURL := fmt.Sprintf("%v/eventstreams", c.config.URL)
	resp := &types.EventStreamResponse{}

	err := callWithBackOff(ctx, c.config.backOff, func() error {
		response, err := clientutils.PostRequest(ctx, c.client, reqURL, request)
		if err != nil {
			return err
		}
		defer clientutils.CloseResponse(response)
		return parseResponse(ctx, response, resp)
	})

	return resp, err
}

func (c *HTTPClient) UpdateEventStream(ctx context.Context, uuid string, request *types.UpdateEventStreamRequest) (*types.EventStreamResponse, error) {
	reqURL := fmt.Sprintf("%v/eventstreams/%v", c.config.URL, uuid)
	resp := &types.EventStreamResponse{}

	err := callWithBackOff(ctx, c.config.backOff, func() error {
		response, err := clientutils.PatchRequest(ctx, c.client, reqURL, request)
		if err != nil {
			return err
		}

		defer clientutils.CloseResponse(response)
		return parseResponse(ctx, response, resp)
	})

	return resp, err
}

func (c *HTTPClient) DeleteEventStream(ctx context.Context, uuid string) error {
	reqURL := fmt.Sprintf("%v/eventstreams/%v", c.config.URL, uuid)

	response, err := clientutils.DeleteRequest(ctx, c.client, reqURL)
	if err != nil {
		return err
	}

	defer clientutils.CloseResponse(response)
	return ParseEmptyBodyResponse(ctx, response)
}
