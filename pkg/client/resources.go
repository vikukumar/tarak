package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// Pod represents a simplified typed pod object.
type Pod struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Phase     string            `json:"phase"`
	IP        string            `json:"ip"`
	Node      string            `json:"node"`
	Labels    map[string]string `json:"labels"`
}

// PodsInterface exposes Pod CRUD operations for a namespace.
type PodsInterface interface {
	List(ctx context.Context) ([]Pod, error)
	Get(ctx context.Context, name string) (map[string]interface{}, error)
	Create(ctx context.Context, obj map[string]interface{}) (map[string]interface{}, error)
	Delete(ctx context.Context, name string) error
}

type podClient struct {
	c  *Client
	ns string
}

// Pods returns a PodsInterface for the given namespace.
func (c *Client) Pods(namespace string) PodsInterface {
	return &podClient{c: c, ns: namespace}
}

func (p *podClient) List(ctx context.Context) ([]Pod, error) {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods", p.ns)
	data, err := p.c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var list struct {
		Items []struct {
			Metadata struct {
				Name      string            `json:"name"`
				Namespace string            `json:"namespace"`
				Labels    map[string]string `json:"labels"`
			} `json:"metadata"`
			Spec struct {
				NodeName string `json:"nodeName"`
			} `json:"spec"`
			Status struct {
				Phase string `json:"phase"`
				PodIP string `json:"podIP"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}

	pods := make([]Pod, 0, len(list.Items))
	for _, it := range list.Items {
		phase := it.Status.Phase
		if phase == "" {
			phase = "Pending"
		}
		pods = append(pods, Pod{
			Name:      it.Metadata.Name,
			Namespace: it.Metadata.Namespace,
			Phase:     phase,
			IP:        it.Status.PodIP,
			Node:      it.Spec.NodeName,
			Labels:    it.Metadata.Labels,
		})
	}
	return pods, nil
}

func (p *podClient) Get(ctx context.Context, name string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s", p.ns, name)
	data, err := p.c.get(ctx, path)
	if err != nil {
		return nil, err
	}
	var res map[string]interface{}
	return res, json.Unmarshal(data, &res)
}

func (p *podClient) Create(ctx context.Context, obj map[string]interface{}) (map[string]interface{}, error) {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods", p.ns)
	data, err := p.c.post(ctx, path, obj)
	if err != nil {
		return nil, err
	}
	var res map[string]interface{}
	return res, json.Unmarshal(data, &res)
}

func (p *podClient) Delete(ctx context.Context, name string) error {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s", p.ns, name)
	return p.c.delete(ctx, path)
}

// ─── Namespaces ──────────────────────────────────────────────────────────────

// NamespacesInterface provides namespace management.
type NamespacesInterface interface {
	List(ctx context.Context) ([]string, error)
	Create(ctx context.Context, name string) error
	Delete(ctx context.Context, name string) error
}

type nsClient struct {
	c *Client
}

// Namespaces returns the NamespacesInterface.
func (c *Client) Namespaces() NamespacesInterface {
	return &nsClient{c: c}
}

func (n *nsClient) List(ctx context.Context) ([]string, error) {
	data, err := n.c.get(ctx, "/api/v1/namespaces")
	if err != nil {
		return nil, err
	}
	var list struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		} `json:"items"`
	}
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	res := make([]string, 0, len(list.Items))
	for _, it := range list.Items {
		res = append(res, it.Metadata.Name)
	}
	return res, nil
}

func (n *nsClient) Create(ctx context.Context, name string) error {
	payload := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata": map[string]interface{}{
			"name": name,
		},
	}
	_, err := n.c.post(ctx, "/api/v1/namespaces", payload)
	return err
}

func (n *nsClient) Delete(ctx context.Context, name string) error {
	return n.c.delete(ctx, "/api/v1/namespaces/"+name)
}
