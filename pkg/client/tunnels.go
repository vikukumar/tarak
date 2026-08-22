package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// TunnelInfo represents the live operational status of a tunnel.
type TunnelInfo struct {
	Type      string    `json:"type"`
	Active    bool      `json:"active"`
	PublicURL string    `json:"publicURL"`
	Mode      string    `json:"mode"`
	StartedAt time.Time `json:"startedAt"`
	LastError string    `json:"lastError,omitempty"`
}

// TunnelsInterface exposes tunnel inspection operations.
type TunnelsInterface interface {
	List(ctx context.Context) ([]TunnelInfo, error)
}

type tunnelClient struct {
	c *Client
}

// Tunnels returns a TunnelsInterface for inspecting Cloudflare and Tailscale tunnels.
func (c *Client) Tunnels() TunnelsInterface {
	return &tunnelClient{c: c}
}

func (t *tunnelClient) List(ctx context.Context) ([]TunnelInfo, error) {
	data, err := t.c.get(ctx, "/apis/networking.tarak.io/v1/tunnels")
	if err != nil {
		return nil, err
	}

	var resp struct {
		Items []TunnelInfo `json:"items"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal tunnels: %w", err)
	}
	return resp.Items, nil
}

// IngressesInterface exposes Ingress operations for a namespace.
type IngressesInterface interface {
	List(ctx context.Context) ([]map[string]interface{}, error)
	Get(ctx context.Context, name string) (map[string]interface{}, error)
	Create(ctx context.Context, obj map[string]interface{}) (map[string]interface{}, error)
	Delete(ctx context.Context, name string) error
}

type ingressClient struct {
	c  *Client
	ns string
}

// Ingresses returns an IngressesInterface for the given namespace.
func (c *Client) Ingresses(namespace string) IngressesInterface {
	return &ingressClient{c: c, ns: namespace}
}

func (i *ingressClient) List(ctx context.Context) ([]map[string]interface{}, error) {
	path := fmt.Sprintf("/apis/networking.k8s.io/v1/namespaces/%s/ingresses", i.ns)
	data, err := i.c.get(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Items []map[string]interface{} `json:"items"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (i *ingressClient) Get(ctx context.Context, name string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/apis/networking.k8s.io/v1/namespaces/%s/ingresses/%s", i.ns, name)
	data, err := i.c.get(ctx, path)
	if err != nil {
		return nil, err
	}
	var obj map[string]interface{}
	err = json.Unmarshal(data, &obj)
	return obj, err
}

func (i *ingressClient) Create(ctx context.Context, obj map[string]interface{}) (map[string]interface{}, error) {
	path := fmt.Sprintf("/apis/networking.k8s.io/v1/namespaces/%s/ingresses", i.ns)
	data, err := i.c.post(ctx, path, obj)
	if err != nil {
		return nil, err
	}
	var res map[string]interface{}
	err = json.Unmarshal(data, &res)
	return res, err
}

func (i *ingressClient) Delete(ctx context.Context, name string) error {
	path := fmt.Sprintf("/apis/networking.k8s.io/v1/namespaces/%s/ingresses/%s", i.ns, name)
	return i.c.delete(ctx, path)
}
