// Code generated by protoc-gen-go-ttrpc. DO NOT EDIT.
// source: github.com/fuweid/benchmark-yamux/api/api.proto
package api

import (
	context "context"
	ttrpc "github.com/containerd/ttrpc"
)

type UnknownHubService interface {
	Read(ctx context.Context, req *ReadRequest) (*ReadResponse, error)
}

func RegisterUnknownHubService(srv *ttrpc.Server, svc UnknownHubService) {
	srv.Register("api.v1.UnknownHub", map[string]ttrpc.Method{
		"Read": func(ctx context.Context, unmarshal func(interface{}) error) (interface{}, error) {
			var req ReadRequest
			if err := unmarshal(&req); err != nil {
				return nil, err
			}
			return svc.Read(ctx, &req)
		},
	})
}

type unknownHubClient struct {
	client *ttrpc.Client
}

func NewUnknownHubClient(client *ttrpc.Client) UnknownHubService {
	return &unknownHubClient{
		client: client,
	}
}
func (c *unknownHubClient) Read(ctx context.Context, req *ReadRequest) (*ReadResponse, error) {
	var resp ReadResponse
	if err := c.client.Call(ctx, "api.v1.UnknownHub", "Read", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
