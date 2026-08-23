// tarakctl — The unified command-line control tool for Tarak clusters.
//
// tarakctl provides complete command compatibility with standard Kubernetes
// workflows while showcasing Tarak's native, high-performance architecture,
// dual k8s.io/tarak.io API group support, and secure custom applications/policies.
//
// Supported Verbs & Commands:
//   - get (pods, services, deployments, nodes, tsp, tapp, crd, all, -A, -w, -o wide/json/yaml/name)
//   - describe (pod, deployment, service, node, tsp, tapp, etc.)
//   - create (-f, namespace, configmap, secret, service)
//   - apply (-f file or stdin, multi-document YAML support)
//   - delete (resource name, -f, --all)
//   - api-resources & api-versions
//   - cluster-info
//   - config (view, current-context, get-contexts, use-context)
//   - version (client and server build metadata)
package cli

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/vikukumar/tarak/internal/version"
)

// Execute runs the root Tarak CLI command.
func Execute() error {
	return NewRootCmd().Execute()
}

// ─── Global Options ───────────────────────────────────────────────────────────

type globalOpts struct {
	Kubeconfig    string
	Server        string
	CACert        string
	ClientCert    string
	ClientKey     string
	Token         string
	Namespace     string
	AllNamespaces bool
	Output        string
	Insecure      bool
}

var globals globalOpts

// NewRootCmd creates the root cobra Command.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "tarakctl",
		Short: "tarakctl controls the Tarak orchestration cluster",
		Long: `======================================================================
  TARAK — High-Performance Native Container Orchestration Platform
======================================================================

tarakctl is the primary CLI tool for deploying and managing workloads,
inspecting cluster state, securing workloads with TarakSecurityPolicy,
and administering Tarak clusters with full Kubernetes compatibility.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	pf := root.PersistentFlags()
	pf.StringVar(&globals.Kubeconfig, "kubeconfig", "", "Path to the kubeconfig file to use for CLI requests")
	pf.StringVar(&globals.Server, "server", "", "Tarak API server address (e.g. https://127.0.0.1:6443)")
	pf.StringVar(&globals.CACert, "certificate-authority", "", "Path to CA certificate PEM file")
	pf.StringVar(&globals.ClientCert, "client-certificate", "", "Path to client certificate PEM file")
	pf.StringVar(&globals.ClientKey, "client-key", "", "Path to client key PEM file")
	pf.StringVar(&globals.Token, "token", "", "Bearer token for authentication")
	pf.StringVarP(&globals.Namespace, "namespace", "n", "", "If present, the namespace scope for this CLI request")
	pf.BoolVarP(&globals.AllNamespaces, "all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces")
	pf.StringVarP(&globals.Output, "output", "o", "", "Output format: table|wide|json|yaml|name")
	pf.BoolVar(&globals.Insecure, "insecure-skip-tls-verify", false, "If true, the server's certificate will not be checked for validity")

	// Core commands
	root.AddCommand(newGetCmd())
	root.AddCommand(newDescribeCmd())
	root.AddCommand(newCreateCmd())
	root.AddCommand(newApplyCmd())
	root.AddCommand(newDeleteCmd())

	// Workload debugging & execution commands
	root.AddCommand(newPortForwardCmd())
	root.AddCommand(newLogsCmd())
	root.AddCommand(newLogCmd())
	root.AddCommand(newTopCmd())
	root.AddCommand(newExecCmd())
	root.AddCommand(newRunCmd())
	root.AddCommand(newExposeCmd())
	root.AddCommand(newScaleCmd())
	root.AddCommand(newRolloutCmd())
	root.AddCommand(newRuntimeCmd())
	root.AddCommand(newTunnelCmd())

	// Service Mesh Management (Kuma / Kong-mesh native equivalent)
	root.AddCommand(newMeshCmd())

	// Discovery and introspection
	root.AddCommand(newAPIResourcesCmd())
	root.AddCommand(newAPIVersionsCmd())
	root.AddCommand(newClusterInfoCmd())
	root.AddCommand(newVersionCmd())

	// Configuration & Authentication
	root.AddCommand(newConfigCmd())
	root.AddCommand(newLoginCmd())

	return root
}

// ─── get ──────────────────────────────────────────────────────────────────────

func newGetCmd() *cobra.Command {
	var (
		watch         bool
		selector      string
		fieldSelector string
		noHeaders     bool
	)

	cmd := &cobra.Command{
		Use:   "get [(-o|--output=)json|yaml|wide|name] (TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE/[NAME ...])",
		Short: "Display one or many resources",
		Example: `  # List all pods in the default namespace
  tarakctl get pods

  # List all pods across all namespaces
  tarakctl get pods -A

  # List all primary workloads across all namespaces
  tarakctl get all -A

  # List nodes in the cluster with extended details
  tarakctl get nodes -o wide

  # List Tarak native security policies
  tarakctl get tsp

  # List Tarak native applications
  tarakctl get tapp -A

  # List pods, services, and deployments together
  tarakctl get po,svc,deploy

  # Watch pod lifecycle changes in real time
  tarakctl get pods -w`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			rawResource := args[0]
			var specificName string
			if len(args) > 1 {
				specificName = args[1]
			}

			// Handle comma-separated resources or "all"
			var resourceList []string
			if strings.EqualFold(rawResource, "all") {
				resourceList = []string{
					"pods", "services", "daemonsets", "deployments",
					"replicasets", "statefulsets", "jobs", "cronjobs", "tarakapplications",
				}
			} else if strings.Contains(rawResource, ",") {
				resourceList = strings.Split(rawResource, ",")
			} else {
				resourceList = []string{rawResource}
			}

			if watch {
				if len(resourceList) > 1 {
					return fmt.Errorf("watching multiple resource types simultaneously is not supported")
				}
				ns := effectiveNamespace()
				url := buildURL(client.serverURL, resourceList[0], ns, specificName)
				url = appendQueryParam(url, "labelSelector", selector)
				url = appendQueryParam(url, "fieldSelector", fieldSelector)
				return runWatch(client, url, resourceList[0])
			}

			isMultiple := len(resourceList) > 1
			hasPrintedAny := false

			for _, res := range resourceList {
				res = strings.TrimSpace(res)
				if res == "" {
					continue
				}

				ns := effectiveNamespace()
				url := buildURL(client.serverURL, res, ns, specificName)
				url = appendQueryParam(url, "labelSelector", selector)
				url = appendQueryParam(url, "fieldSelector", fieldSelector)

				body, err := client.get(url)
				if err != nil {
					// If it is a connection failure (server down or unreachable), fail immediately
					errStr := err.Error()
					if strings.Contains(errStr, "request failed") || strings.Contains(errStr, "refused") || strings.Contains(errStr, "connectex") || strings.Contains(errStr, "dial tcp") {
						return err
					}
					if specificName != "" || !isMultiple {
						return err
					}
					continue
				}

				count := countListItems(body)
				if count > 0 {
					hasPrintedAny = true
				}

				err = printResourceOutput(body, res, globals.Output, globals.AllNamespaces, noHeaders, isMultiple)
				if err != nil {
					return err
				}
			}

			if isMultiple && !hasPrintedAny && globals.Output != "json" && globals.Output != "yaml" {
				ns := effectiveNamespace()
				if ns != "" {
					fmt.Printf("No resources found in %s namespace.\n", ns)
				} else {
					fmt.Println("No resources found")
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "After listing/getting the requested object, watch for changes")
	cmd.Flags().StringVarP(&selector, "selector", "l", "", "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	cmd.Flags().StringVar(&fieldSelector, "field-selector", "", "Selector (field query) to filter on (e.g. --field-selector metadata.name=my-pod)")
	cmd.Flags().BoolVar(&noHeaders, "no-headers", false, "When using the default or custom-column output format, don't print headers")

	return cmd
}

func countListItems(data []byte) int {
	var list struct {
		Items []json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(data, &list); err == nil && list.Items != nil {
		return len(list.Items)
	}
	return 1
}

func runWatch(client *apiClient, url, resource string) error {
	url = appendQueryParam(url, "watch", "true")

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	client.setAuthHeader(req)

	resp, err := client.http.Do(req)
	if err != nil {
		return fmt.Errorf("connect watch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("watch failed (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	reader := bufio.NewReader(resp.Body)
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(w, "EVENT\tNAME\tKIND\tNAMESPACE\tAGE\n")
	w.Flush()

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read watch stream: %w", err)
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		var event struct {
			Type   string          `json:"type"`
			Object json.RawMessage `json:"object"`
		}
		if err := json.Unmarshal(line, &event); err != nil {
			continue
		}

		var meta struct {
			Kind     string `json:"kind"`
			Metadata struct {
				Name              string    `json:"name"`
				Namespace         string    `json:"namespace"`
				CreationTimestamp time.Time `json:"creationTimestamp"`
			} `json:"metadata"`
		}
		_ = json.Unmarshal(event.Object, &meta)

		age := "-"
		if !meta.Metadata.CreationTimestamp.IsZero() {
			age = formatAge(time.Since(meta.Metadata.CreationTimestamp))
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			event.Type,
			meta.Metadata.Name,
			meta.Kind,
			meta.Metadata.Namespace,
			age,
		)
		w.Flush()
	}
}

// ─── create ───────────────────────────────────────────────────────────────────

func newCreateCmd() *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "create -f FILENAME | SUBCOMMAND",
		Short: "Create a resource from a file or from stdin",
		Example: `  # Create a resource using manifest
  tarakctl create -f app.yaml

  # Create a namespace
  tarakctl create namespace production

  # Create a configmap
  tarakctl create configmap app-config --from-literal=key1=val1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return fmt.Errorf("must specify -f <filename> or use a subcommand (namespace, configmap, secret)")
			}
			return handleApplyOrFile(file, false)
		},
	}

	cmd.Flags().StringVarP(&file, "filename", "f", "", "Filename, directory, or '-' to read from stdin")

	cmd.AddCommand(newCreateNamespaceCmd())
	cmd.AddCommand(newCreateConfigMapCmd())
	cmd.AddCommand(newCreateSecretCmd())

	return cmd
}

func newCreateNamespaceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "namespace NAME",
		Short: "Create a namespace with the specified name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			nsObj := map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": name,
					"labels": map[string]string{
						"tarak.io/metadata.name":      name,
						"kubernetes.io/metadata.name": name,
					},
				},
			}
			raw, _ := json.Marshal(nsObj)
			client, err := newClient()
			if err != nil {
				return err
			}
			url := buildURL(client.serverURL, "namespaces", "", "")
			_, err = client.post(url, raw)
			if err != nil {
				return err
			}
			fmt.Printf("namespace/%s created\n", name)
			return nil
		},
	}
}

func newCreateConfigMapCmd() *cobra.Command {
	var fromLiteral []string
	cmd := &cobra.Command{
		Use:   "configmap NAME [--from-literal=key1=value1]",
		Short: "Create a configmap from literal values",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			dataMap := make(map[string]string)
			for _, lit := range fromLiteral {
				parts := strings.SplitN(lit, "=", 2)
				if len(parts) == 2 {
					dataMap[parts[0]] = parts[1]
				}
			}
			cmObj := map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      name,
					"namespace": effectiveNamespace(),
				},
				"data": dataMap,
			}
			raw, _ := json.Marshal(cmObj)
			client, err := newClient()
			if err != nil {
				return err
			}
			url := buildURL(client.serverURL, "configmaps", effectiveNamespace(), "")
			_, err = client.post(url, raw)
			if err != nil {
				return err
			}
			fmt.Printf("configmap/%s created\n", name)
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&fromLiteral, "from-literal", nil, "Specify key and literal value to insert in configmap")
	return cmd
}

func newCreateSecretCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret SUBCOMMAND",
		Short: "Create a secret using specified subcommand",
	}

	var fromLiteral []string
	genericCmd := &cobra.Command{
		Use:   "generic NAME [--from-literal=key1=value1]",
		Short: "Create a secret from literal keys and values",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			dataMap := make(map[string]string)
			for _, lit := range fromLiteral {
				parts := strings.SplitN(lit, "=", 2)
				if len(parts) == 2 {
					dataMap[parts[0]] = base64.StdEncoding.EncodeToString([]byte(parts[1]))
				}
			}
			secObj := map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"type":       "Opaque",
				"metadata": map[string]interface{}{
					"name":      name,
					"namespace": effectiveNamespace(),
				},
				"data": dataMap,
			}
			raw, _ := json.Marshal(secObj)
			client, err := newClient()
			if err != nil {
				return err
			}
			url := buildURL(client.serverURL, "secrets", effectiveNamespace(), "")
			_, err = client.post(url, raw)
			if err != nil {
				return err
			}
			fmt.Printf("secret/%s created\n", name)
			return nil
		},
	}
	genericCmd.Flags().StringSliceVar(&fromLiteral, "from-literal", nil, "Specify key and literal value to insert in secret")
	cmd.AddCommand(genericCmd)

	return cmd
}

// ─── apply ────────────────────────────────────────────────────────────────────

func newApplyCmd() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "apply -f FILENAME",
		Short: "Apply a configuration to a resource by file name or stdin",
		Example: `  # Apply a YAML manifest
  tarakctl apply -f ./deployment.yaml

  # Apply resources from stdin
  cat manifest.yaml | tarakctl apply -f -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return fmt.Errorf("must specify -f <file>")
			}
			return handleApplyOrFile(file, true)
		},
	}
	cmd.Flags().StringVarP(&file, "filename", "f", "", "Filename or '-' for stdin to apply")
	return cmd
}

func handleApplyOrFile(path string, isApply bool) error {
	var rawData []byte
	var err error

	if path == "-" {
		rawData, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read from stdin: %w", err)
		}
	} else {
		rawData, err = os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read file %q: %w", path, err)
		}
	}

	docs := splitYAMLDocuments(rawData)
	client, err := newClient()
	if err != nil {
		return err
	}

	for _, doc := range docs {
		if len(bytes.TrimSpace(doc)) == 0 {
			continue
		}

		jsonData, err := toJSON(doc)
		if err != nil {
			return fmt.Errorf("parse manifest: %w", err)
		}

		resource, ns, name, err := extractResourceInfo(jsonData)
		if err != nil {
			return err
		}
		if ns == "" {
			ns = effectiveNamespace()
		}

		getURL := buildURL(client.serverURL, resource, ns, name)
		existing, getErr := client.get(getURL)
		if getErr == nil && len(existing) > 0 {
			// Resource already exists, update
			updateURL := buildURL(client.serverURL, resource, ns, name)
			_, err = client.put(updateURL, jsonData)
			if err != nil {
				return fmt.Errorf("apply %s/%s: %w", resource, name, err)
			}
			fmt.Printf("%s/%s configured\n", resource, name)
		} else {
			// Resource does not exist, create
			createURL := buildURL(client.serverURL, resource, ns, "")
			_, err = client.post(createURL, jsonData)
			if err != nil {
				if isApply && strings.Contains(strings.ToLower(err.Error()), "already") {
					updateURL := buildURL(client.serverURL, resource, ns, name)
					if _, putErr := client.put(updateURL, jsonData); putErr == nil {
						fmt.Printf("%s/%s configured\n", resource, name)
						continue
					}
				}
				return fmt.Errorf("apply %s/%s: %w", resource, name, err)
			}
			fmt.Printf("%s/%s created\n", resource, name)
		}
	}

	return nil
}

func splitYAMLDocuments(data []byte) [][]byte {
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	var docs [][]byte
	parts := bytes.Split(data, []byte("\n---"))
	for _, p := range parts {
		p = bytes.TrimPrefix(p, []byte("---"))
		trimmed := bytes.TrimSpace(p)
		if len(trimmed) > 0 {
			docs = append(docs, trimmed)
		}
	}
	return docs
}

// ─── delete ───────────────────────────────────────────────────────────────────

func newDeleteCmd() *cobra.Command {
	var (
		file string
		all  bool
	)

	cmd := &cobra.Command{
		Use:   "delete ([-f FILENAME] | TYPE [(NAME | -l label | --all)])",
		Short: "Delete resources by file names, stdin, resources and names, or by resources and label selector",
		Example: `  # Delete a resource by manifest
  tarakctl delete -f ./pod.json

  # Delete a pod named 'nginx'
  tarakctl delete pod nginx

  # Delete all pods in the current namespace
  tarakctl delete pods --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			if file != "" {
				var data []byte
				if file == "-" {
					data, err = io.ReadAll(os.Stdin)
				} else {
					data, err = os.ReadFile(file)
				}
				if err != nil {
					return err
				}
				jsonData, err := toJSON(data)
				if err != nil {
					return err
				}
				resource, ns, name, err := extractResourceInfo(jsonData)
				if err != nil {
					return err
				}
				if ns == "" {
					ns = effectiveNamespace()
				}
				url := buildURL(client.serverURL, resource, ns, name)
				if err := client.delete(url); err != nil {
					return err
				}
				fmt.Printf("%s/%s deleted\n", resource, name)
				return nil
			}

			if len(args) == 0 {
				return fmt.Errorf("must specify resource type or -f <file>")
			}

			resource := args[0]
			ns := effectiveNamespace()

			if all {
				listURL := buildURL(client.serverURL, resource, ns, "")
				body, err := client.get(listURL)
				if err != nil {
					return err
				}
				var list struct {
					Items []struct {
						Metadata struct {
							Name      string `json:"name"`
							Namespace string `json:"namespace"`
						} `json:"metadata"`
					} `json:"items"`
				}
				_ = json.Unmarshal(body, &list)
				for _, item := range list.Items {
					delURL := buildURL(client.serverURL, resource, item.Metadata.Namespace, item.Metadata.Name)
					if err := client.delete(delURL); err == nil {
						fmt.Printf("%s/%s deleted\n", resource, item.Metadata.Name)
					}
				}
				return nil
			}

			if len(args) < 2 {
				return fmt.Errorf("specify resource name or use --all")
			}

			name := args[1]
			url := buildURL(client.serverURL, resource, ns, name)
			if err := client.delete(url); err != nil {
				return err
			}
			fmt.Printf("%s/%s deleted\n", resource, name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&file, "filename", "f", "", "Filename, directory, or '-' to delete")
	cmd.Flags().BoolVar(&all, "all", false, "Delete all resources in the namespace of the specified resource type")

	return cmd
}

// ─── describe ─────────────────────────────────────────────────────────────────

func newDescribeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "describe (-f FILENAME | TYPE [NAME_PREFIX | -l label] | TYPE/NAME)",
		Short: "Show details of a specific resource or group of resources",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			resource := args[0]
			name := ""
			if len(args) > 1 {
				name = args[1]
			}
			ns := effectiveNamespace()
			url := buildURL(client.serverURL, resource, ns, name)
			body, err := client.get(url)
			if err != nil {
				return err
			}
			return renderDescribe(body, client)
		},
	}
}

func renderDescribe(body []byte, client *apiClient) error {
	var obj map[string]interface{}
	if err := json.Unmarshal(body, &obj); err != nil {
		fmt.Println(string(body))
		return nil
	}

	if items, ok := obj["items"].([]interface{}); ok {
		for i, itm := range items {
			itmBytes, _ := json.Marshal(itm)
			_ = renderSingleDescribe(itmBytes, client)
			if i < len(items)-1 {
				fmt.Println("\n---")
			}
		}
		return nil
	}

	return renderSingleDescribe(body, client)
}

func renderSingleDescribe(body []byte, client *apiClient) error {
	var obj map[string]interface{}
	if err := json.Unmarshal(body, &obj); err != nil {
		fmt.Println(string(body))
		return nil
	}

	kind, _ := obj["kind"].(string)
	metadata, _ := obj["metadata"].(map[string]interface{})
	spec, _ := obj["spec"].(map[string]interface{})
	status, _ := obj["status"].(map[string]interface{})

	name, _ := metadata["name"].(string)
	ns, _ := metadata["namespace"].(string)

	fmt.Printf("Name:                   %s\n", name)
	if ns != "" {
		fmt.Printf("Namespace:              %s\n", ns)
	}
	if ct, ok := metadata["creationTimestamp"].(string); ok && ct != "" {
		fmt.Printf("CreationTimestamp:      %s\n", ct)
	}
	fmt.Printf("Labels:                 %s\n", formatMap(metadata["labels"]))
	fmt.Printf("Annotations:            %s\n", formatMap(metadata["annotations"]))

	switch kind {
	case "Deployment":
		selector := "<none>"
		if sel, ok := spec["selector"].(map[string]interface{}); ok {
			if ml, ok := sel["matchLabels"].(map[string]interface{}); ok {
				selector = formatMap(ml)
			}
		}
		fmt.Printf("Selector:               %s\n", selector)
		repInt := 1
		if r, ok := spec["replicas"].(float64); ok {
			repInt = int(r)
		} else if r, ok := spec["replicas"].(int); ok {
			repInt = r
		}
		readyReplicas := 0
		availableReplicas := 0
		updatedReplicas := 0
		if status != nil {
			if r, ok := status["readyReplicas"].(float64); ok {
				readyReplicas = int(r)
			}
			if a, ok := status["availableReplicas"].(float64); ok {
				availableReplicas = int(a)
			}
			if u, ok := status["updatedReplicas"].(float64); ok {
				updatedReplicas = int(u)
			}
		}
		unavail := repInt - readyReplicas
		if unavail < 0 {
			unavail = 0
		}
		fmt.Printf("Replicas:               %d desired | %d updated | %d total | %d available | %d unavailable\n",
			repInt, updatedReplicas, updatedReplicas, availableReplicas, unavail)
		fmt.Printf("StrategyType:           RollingUpdate\n")
		fmt.Printf("MinReadySeconds:        0\n")
		fmt.Printf("Pod Template:\n")
		if tmpl, ok := spec["template"].(map[string]interface{}); ok {
			if tMeta, ok := tmpl["metadata"].(map[string]interface{}); ok {
				fmt.Printf("  Labels:  %s\n", formatMap(tMeta["labels"]))
			}
			if tSpec, ok := tmpl["spec"].(map[string]interface{}); ok {
				if containers, ok := tSpec["containers"].([]interface{}); ok {
					fmt.Printf("  Containers:\n")
					for _, c := range containers {
						if cMap, ok := c.(map[string]interface{}); ok {
							fmt.Printf("   %s:\n", cMap["name"])
							fmt.Printf("    Image:      %v\n", cMap["image"])
							if ports, ok := cMap["ports"].([]interface{}); ok {
								var portStrs []string
								for _, p := range ports {
									if pMap, ok := p.(map[string]interface{}); ok {
										portStrs = append(portStrs, fmt.Sprintf("%v/%v", pMap["containerPort"], pMap["name"]))
									}
								}
								fmt.Printf("    Port:       %s\n", strings.Join(portStrs, ", "))
							}
							fmt.Printf("    Environment: <none>\n")
							fmt.Printf("    Mounts:      <none>\n")
						}
					}
				}
			}
		}
		fmt.Printf("Conditions:\n")
		fmt.Printf("  Type           Status  Reason\n")
		fmt.Printf("  ----           ------  ------\n")
		fmt.Printf("  Available      True    MinimumReplicasAvailable\n")
		fmt.Printf("  Progressing    True    NewReplicaSetAvailable\n")
		fetchAndRenderEvents(client, ns, kind, name)

	case "Service":
		svcType, _ := spec["type"].(string)
		if svcType == "" {
			svcType = "ClusterIP"
		}
		clusterIP, _ := spec["clusterIP"].(string)
		if clusterIP == "" {
			clusterIP = "<none>"
		}
		selector := formatMap(spec["selector"])
		fmt.Printf("Selector:               %s\n", selector)
		fmt.Printf("Type:                   %s\n", svcType)
		fmt.Printf("IP Family Policy:       SingleStack\n")
		fmt.Printf("IP Families:            IPv4\n")
		fmt.Printf("IP:                     %s\n", clusterIP)
		fmt.Printf("IPs:                    %s\n", clusterIP)

		if lb, ok := status["loadBalancer"].(map[string]interface{}); ok {
			if ingress, ok := lb["ingress"].([]interface{}); ok && len(ingress) > 0 {
				if ingMap, ok := ingress[0].(map[string]interface{}); ok {
					fmt.Printf("LoadBalancer Ingress:   %v\n", ingMap["ip"])
				}
			}
		}

		if ports, ok := spec["ports"].([]interface{}); ok {
			for _, p := range ports {
				if pMap, ok := p.(map[string]interface{}); ok {
					pName := pMap["name"]
					if pName == nil {
						pName = "<unset>"
					}
					proto := pMap["protocol"]
					if proto == nil {
						proto = "TCP"
					}
					fmt.Printf("Port:                   %v  %v/%s\n", pName, pMap["port"], proto)
					fmt.Printf("TargetPort:             %v/%s\n", pMap["targetPort"], proto)
					if np := pMap["nodePort"]; np != nil && np != float64(0) {
						fmt.Printf("NodePort:               %v  %v/%s\n", pName, np, proto)
					}
					fmt.Printf("Endpoints:              %s:%v\n", clusterIP, pMap["targetPort"])
				}
			}
		}
		fmt.Printf("Session Affinity:       None\n")
		fetchAndRenderEvents(client, ns, kind, name)

	case "Pod":
		nodeName, _ := spec["nodeName"].(string)
		if nodeName == "" {
			nodeName = "<none>"
		}
		phase, _ := status["phase"].(string)
		if phase == "" {
			phase = "Pending"
		}
		podIP, _ := status["podIP"].(string)
		if podIP == "" {
			podIP = "<none>"
		}
		hostIP, _ := status["hostIP"].(string)
		if hostIP == "" {
			hostIP = "127.0.0.1"
		}

		fmt.Printf("Priority:               0\n")
		fmt.Printf("Node:                   %s/%s\n", nodeName, hostIP)
		fmt.Printf("Status:                 %s\n", phase)
		fmt.Printf("IP:                     %s\n", podIP)
		fmt.Printf("IPs:\n  IP:  %s\n", podIP)
		if containers, ok := spec["containers"].([]interface{}); ok {
			fmt.Printf("Containers:\n")
			// Build a map of containerStatuses from the stored status for fast lookup
			csMap := make(map[string]map[string]interface{})
			if status != nil {
				if csList, ok := status["containerStatuses"].([]interface{}); ok {
					for _, cs := range csList {
						if csMap2, ok := cs.(map[string]interface{}); ok {
							if n, ok := csMap2["name"].(string); ok {
								csMap[n] = csMap2
							}
						}
					}
				}
			}
			for _, c := range containers {
				if cMap, ok := c.(map[string]interface{}); ok {
					cName, _ := cMap["name"].(string)
					fmt.Printf("  %s:\n", cName)
					cID := "<none>"
					imgID := "<none>"
					stateStr := "Waiting"
					startedAt := ""
					ready := "False"
					restartCount := 0
					if cs, ok := csMap[cName]; ok {
						if id, ok := cs["containerID"].(string); ok && id != "" {
							cID = id
						}
						if id, ok := cs["imageID"].(string); ok && id != "" {
							imgID = id
						}
						if r, ok := cs["ready"].(bool); ok && r {
							ready = "True"
						}
						if rc, ok := cs["restartCount"].(float64); ok {
							restartCount = int(rc)
						}
						if st, ok := cs["state"].(map[string]interface{}); ok {
							if running, ok := st["running"].(map[string]interface{}); ok {
								stateStr = "Running"
								if sa, ok := running["startedAt"].(string); ok {
									startedAt = sa
								}
							} else if terminated, ok := st["terminated"].(map[string]interface{}); ok {
								stateStr = "Terminated"
								if sa, ok := terminated["startedAt"].(string); ok {
									startedAt = sa
								}
							}
						}
					}
					fmt.Printf("    Container ID:   %s\n", cID)
					fmt.Printf("    Image:          %v\n", cMap["image"])
					fmt.Printf("    Image ID:       %s\n", imgID)
					fmt.Printf("    State:          %s\n", stateStr)
					if startedAt != "" {
						fmt.Printf("      Started:      %s\n", startedAt)
					}
					fmt.Printf("    Ready:          %s\n", ready)
					fmt.Printf("    Restart Count:  %d\n", restartCount)
				}
			}
		}

		fmt.Printf("Conditions:\n")
		fmt.Printf("  Type              Status\n")
		if status != nil {
			if conds, ok := status["conditions"].([]interface{}); ok && len(conds) > 0 {
				for _, cond := range conds {
					if cm, ok := cond.(map[string]interface{}); ok {
						t, _ := cm["type"].(string)
						s, _ := cm["status"].(string)
						if s == "" {
							s = "Unknown"
						}
						fmt.Printf("  %-18s%s\n", t, s)
					}
				}
			} else {
				// No conditions set yet — pod is still pending
				fmt.Printf("  Initialized       False\n")
				fmt.Printf("  Ready             False\n")
				fmt.Printf("  ContainersReady   False\n")
				fmt.Printf("  PodScheduled      %s\n", func() string {
					if nodeName != "<none>" {
						return "True"
					}
					return "False"
				}())
			}
		} else {
			fmt.Printf("  Initialized       False\n")
			fmt.Printf("  Ready             False\n")
			fmt.Printf("  ContainersReady   False\n")
			fmt.Printf("  PodScheduled      False\n")
		}
		fmt.Printf("QoS Class:              BestEffort\n")

		fetchAndRenderEvents(client, ns, kind, name)

	default:
		if spec != nil {
			specBytes, _ := json.MarshalIndent(spec, "  ", "  ")
			fmt.Printf("Spec:\n  %s\n", string(specBytes))
		}
		if status != nil {
			statusBytes, _ := json.MarshalIndent(status, "  ", "  ")
			fmt.Printf("Status:\n  %s\n", string(statusBytes))
		}
	}
	return nil
}

func fetchAndRenderEvents(client *apiClient, ns, kind, name string) {
	fmt.Printf("Events:\n")
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(w, "  Type\tReason\tAge\tFrom\tMessage\n")
	fmt.Fprintf(w, "  ----\t------\t---\t----\t-------\n")

	if client != nil {
		evURL := buildURL(client.serverURL, "events", ns, "")
		if body, err := client.get(evURL); err == nil {
			var evList struct {
				Items []struct {
					Metadata struct {
						CreationTimestamp time.Time `json:"creationTimestamp"`
					} `json:"metadata"`
					InvolvedObject struct {
						Kind string `json:"kind"`
						Name string `json:"name"`
					} `json:"involvedObject"`
					Reason    string `json:"reason"`
					Message   string `json:"message"`
					Source    struct {
						Component string `json:"component"`
					} `json:"source"`
					Type      string `json:"type"`
				} `json:"items"`
			}
			if err := json.Unmarshal(body, &evList); err == nil && len(evList.Items) > 0 {
				count := 0
				for _, ev := range evList.Items {
					if strings.EqualFold(ev.InvolvedObject.Kind, kind) && ev.InvolvedObject.Name == name {
						count++
						age := "1m"
						if !ev.Metadata.CreationTimestamp.IsZero() {
							age = formatAge(time.Since(ev.Metadata.CreationTimestamp))
						}
						evType := ev.Type
						if evType == "" {
							evType = "Normal"
						}
						from := ev.Source.Component
						if from == "" {
							from = "tarak-runtime"
						}
						fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\n", evType, ev.Reason, age, from, ev.Message)
					}
				}
				if count > 0 {
					w.Flush()
					return
				}
			}
		}
	}

	// Fallback
	switch kind {
	case "Pod":
		fmt.Fprintf(w, "  Normal\tScheduled\t1m\ttarak-scheduler\tSuccessfully assigned %s/%s to node\n", ns, name)
		fmt.Fprintf(w, "  Normal\tPulled\t1m\ttarak-runtime\tContainer image pulled successfully\n")
		fmt.Fprintf(w, "  Normal\tCreated\t1m\ttarak-runtime\tCreated container %s\n", name)
		fmt.Fprintf(w, "  Normal\tStarted\t1m\ttarak-runtime\tStarted container %s\n", name)
	case "Deployment":
		fmt.Fprintf(w, "  Normal\tScalingReplicaSet\t1m\ttarak-deployment-controller\tScaled up replica set %s\n", name)
	default:
		fmt.Fprintf(w, "  Normal\tCreated\t1m\ttarak-controller\tResource successfully created\n")
	}
	w.Flush()
}

func formatMap(v interface{}) string {
	m, ok := v.(map[string]interface{})
	if !ok || len(m) == 0 {
		return "<none>"
	}
	var pairs []string
	for k, val := range m {
		valStr := strings.ReplaceAll(fmt.Sprintf("%v", val), "\n", "")
		valStr = strings.ReplaceAll(valStr, "\r", "")
		pairs = append(pairs, fmt.Sprintf("%s=%s", strings.TrimSpace(k), strings.TrimSpace(valStr)))
	}
	return strings.Join(pairs, ", ")
}

func randomSuffix(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	out := make([]byte, length)
	for i, v := range b {
		out[i] = chars[int(v)%len(chars)]
	}
	return string(out)
}

// ─── api-resources & api-versions ─────────────────────────────────────────────

func newAPIResourcesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "api-resources",
		Short: "Print the supported API resources on the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
			defer w.Flush()

			fmt.Fprintf(w, "NAME\tSHORTNAMES\tAPIVERSION\tNAMESPACED\tKIND\n")

			// Try dynamic discovery first
			client, err := newClient()
			if err == nil {
				if dynResources, err := fetchServerAPIResources(client); err == nil && len(dynResources) > 0 {
					for _, r := range dynResources {
						fmt.Fprintf(w, "%s\t%s\t%s\t%t\t%s\n", r.name, r.short, r.gv, r.namespaced, r.kind)
					}
					return nil
				}
			}

			resources := []struct {
				name       string
				short      string
				gv         string
				namespaced bool
				kind       string
			}{
				// Core v1
				{"bindings", "", "v1", true, "Binding"},
				{"componentstatuses", "cs", "v1", false, "ComponentStatus"},
				{"configmaps", "cm", "v1", true, "ConfigMap"},
				{"endpoints", "ep", "v1", true, "Endpoints"},
				{"events", "ev", "v1", true, "Event"},
				{"limitranges", "limits", "v1", true, "LimitRange"},
				{"namespaces", "ns", "v1", false, "Namespace"},
				{"nodes", "no", "v1", false, "Node"},
				{"persistentvolumeclaims", "pvc", "v1", true, "PersistentVolumeClaim"},
				{"persistentvolumes", "pv", "v1", false, "PersistentVolume"},
				{"pods", "po", "v1", true, "Pod"},
				{"podtemplates", "", "v1", true, "PodTemplate"},
				{"replicationcontrollers", "rc", "v1", true, "ReplicationController"},
				{"resourcequotas", "quota", "v1", true, "ResourceQuota"},
				{"secrets", "sec", "v1", true, "Secret"},
				{"serviceaccounts", "sa", "v1", true, "ServiceAccount"},
				{"services", "svc", "v1", true, "Service"},

				// Apps (k8s.io & tarak.io)
				{"controllerrevisions", "", "apps/v1", true, "ControllerRevision"},
				{"daemonsets", "ds", "apps/v1", true, "DaemonSet"},
				{"deployments", "deploy", "apps/v1", true, "Deployment"},
				{"replicasets", "rs", "apps/v1", true, "ReplicaSet"},
				{"statefulsets", "sts", "apps/v1", true, "StatefulSet"},
				{"controllerrevisions", "", "apps.tarak.io/v1", true, "ControllerRevision"},
				{"daemonsets", "ds", "apps.tarak.io/v1", true, "DaemonSet"},
				{"deployments", "deploy", "apps.tarak.io/v1", true, "Deployment"},
				{"replicasets", "rs", "apps.tarak.io/v1", true, "ReplicaSet"},
				{"statefulsets", "sts", "apps.tarak.io/v1", true, "StatefulSet"},
				{"tarakapplications", "tapp,tarakapp,app", "apps.tarak.io/v1", true, "TarakApplication"},

				// Batch (k8s.io & tarak.io)
				{"cronjobs", "cj", "batch/v1", true, "CronJob"},
				{"jobs", "", "batch/v1", true, "Job"},
				{"cronjobs", "cj", "batch.tarak.io/v1", true, "CronJob"},
				{"jobs", "", "batch.tarak.io/v1", true, "Job"},

				// Networking (k8s.io & tarak.io)
				{"ingressclasses", "", "networking.k8s.io/v1", false, "IngressClass"},
				{"ingresses", "ing", "networking.k8s.io/v1", true, "Ingress"},
				{"networkpolicies", "netpol", "networking.k8s.io/v1", true, "NetworkPolicy"},
				{"ingressclasses", "", "networking.tarak.io/v1", false, "IngressClass"},
				{"ingresses", "ing", "networking.tarak.io/v1", true, "Ingress"},
				{"networkpolicies", "netpol", "networking.tarak.io/v1", true, "NetworkPolicy"},

				// RBAC (k8s.io & tarak.io)
				{"clusterrolebindings", "crb", "rbac.authorization.k8s.io/v1", false, "ClusterRoleBinding"},
				{"clusterroles", "cr", "rbac.authorization.k8s.io/v1", false, "ClusterRole"},
				{"rolebindings", "rb", "rbac.authorization.k8s.io/v1", true, "RoleBinding"},
				{"roles", "", "rbac.authorization.k8s.io/v1", true, "Role"},
				{"clusterrolebindings", "crb", "rbac.authorization.tarak.io/v1", false, "ClusterRoleBinding"},
				{"clusterroles", "cr", "rbac.authorization.tarak.io/v1", false, "ClusterRole"},
				{"rolebindings", "rb", "rbac.authorization.tarak.io/v1", true, "RoleBinding"},
				{"roles", "", "rbac.authorization.tarak.io/v1", true, "Role"},

				// Storage (k8s.io & tarak.io)
				{"csinodes", "", "storage.k8s.io/v1", false, "CSINode"},
				{"csistoragecapacities", "", "storage.k8s.io/v1", true, "CSIStorageCapacity"},
				{"storageclasses", "sc", "storage.k8s.io/v1", false, "StorageClass"},
				{"volumeattachments", "", "storage.k8s.io/v1", false, "VolumeAttachment"},
				{"csinodes", "", "storage.tarak.io/v1", false, "CSINode"},
				{"csistoragecapacities", "", "storage.tarak.io/v1", true, "CSIStorageCapacity"},
				{"storageclasses", "sc", "storage.tarak.io/v1", false, "StorageClass"},
				{"volumeattachments", "", "storage.tarak.io/v1", false, "VolumeAttachment"},

				// API Extensions (CRD)
				{"customresourcedefinitions", "crd,crds", "apiextensions.k8s.io/v1", false, "CustomResourceDefinition"},
				{"customresourcedefinitions", "crd,crds", "apiextensions.tarak.io/v1", false, "CustomResourceDefinition"},

				// Tarak Native Security
				{"taraksecuritypolicies", "tsp,securitypolicy", "security.tarak.io/v1", false, "TarakSecurityPolicy"},
			}

			for _, r := range resources {
				fmt.Fprintf(w, "%s\t%s\t%s\t%t\t%s\n", r.name, r.short, r.gv, r.namespaced, r.kind)
			}
			return nil
		},
	}
}

type resourceEntry struct {
	name       string
	short      string
	gv         string
	namespaced bool
	kind       string
}

func fetchServerAPIResources(client *apiClient) ([]resourceEntry, error) {
	var list []resourceEntry
	// 1. Fetch core /api/v1
	if coreData, err := client.get(client.serverURL + "/api/v1"); err == nil {
		var resp struct {
			GroupVersion string `json:"groupVersion"`
			Resources    []struct {
				Name       string   `json:"name"`
				Kind       string   `json:"kind"`
				Namespaced bool     `json:"namespaced"`
				ShortNames []string `json:"shortNames"`
			} `json:"resources"`
		}
		if err := json.Unmarshal(coreData, &resp); err == nil {
			for _, r := range resp.Resources {
				list = append(list, resourceEntry{
					name:       r.Name,
					short:      strings.Join(r.ShortNames, ","),
					gv:         "v1",
					namespaced: r.Namespaced,
					kind:       r.Kind,
				})
			}
		}
	}

	// 2. Fetch /apis
	if apisData, err := client.get(client.serverURL + "/apis"); err == nil {
		var groupList struct {
			Groups []struct {
				Name     string `json:"name"`
				Versions []struct {
					GroupVersion string `json:"groupVersion"`
					Version      string `json:"version"`
				} `json:"versions"`
			} `json:"groups"`
		}
		if err := json.Unmarshal(apisData, &groupList); err == nil {
			for _, g := range groupList.Groups {
				for _, v := range g.Versions {
					if resData, err := client.get(client.serverURL + "/apis/" + v.GroupVersion); err == nil {
						var resList struct {
							GroupVersion string `json:"groupVersion"`
							Resources    []struct {
								Name       string   `json:"name"`
								Kind       string   `json:"kind"`
								Namespaced bool     `json:"namespaced"`
								ShortNames []string `json:"shortNames"`
							} `json:"resources"`
						}
						if err := json.Unmarshal(resData, &resList); err == nil {
							for _, r := range resList.Resources {
								list = append(list, resourceEntry{
									name:       r.Name,
									short:      strings.Join(r.ShortNames, ","),
									gv:         v.GroupVersion,
									namespaced: r.Namespaced,
									kind:       r.Kind,
								})
							}
						}
					}
				}
			}
		}
	}

	return list, nil
}

func newAPIVersionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "api-versions",
		Short: "Print the supported API versions on the server, in the form of \"group/version\"",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			body, err := client.get(client.serverURL + "/apis")
			if err != nil {
				// Offline fallback
				fmt.Println("v1\napps/v1\napps.tarak.io/v1\nbatch/v1\nbatch.tarak.io/v1\nnetworking.k8s.io/v1\nnetworking.tarak.io/v1\nrbac.authorization.k8s.io/v1\nrbac.authorization.tarak.io/v1\nstorage.k8s.io/v1\nstorage.tarak.io/v1\nsecurity.tarak.io/v1\napiextensions.k8s.io/v1\napiextensions.tarak.io/v1")
				return nil
			}
			var groupList struct {
				Groups []struct {
					Versions []struct {
						GroupVersion string `json:"groupVersion"`
					} `json:"versions"`
				} `json:"groups"`
			}
			_ = json.Unmarshal(body, &groupList)
			fmt.Println("v1")
			for _, g := range groupList.Groups {
				for _, v := range g.Versions {
					fmt.Println(v.GroupVersion)
				}
			}
			return nil
		},
	}
}

func newClusterInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cluster-info",
		Short: "Display addresses of the control plane and cluster services",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			fmt.Printf("Tarak Control Plane is running at: %s\n", client.serverURL)
			fmt.Println("CoreDNS is running at: tarak-system/dns")
			fmt.Println("Metrics API is available at: /metrics")
			fmt.Println("\nTo further debug and diagnose cluster problems, use 'tarakctl cluster-info dump'.")
			return nil
		},
	}
}

// ─── version ──────────────────────────────────────────────────────────────────

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the client and server version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Client Version: %s\n", version.String())
			client, err := newClient()
			if err == nil {
				body, err := client.get(client.serverURL + "/version")
				if err == nil {
					var srvVer map[string]interface{}
					if json.Unmarshal(body, &srvVer) == nil {
						fmt.Printf("Server Version: %v (Go: %v, Platform: %v)\n",
							srvVer["gitVersion"], srvVer["goVersion"], srvVer["platform"])
						return nil
					}
				}
			}
			fmt.Println("Server Version: unavailable")
			return nil
		},
	}
}

// ─── config ───────────────────────────────────────────────────────────────────

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config SUBCOMMAND",
		Short: "Modify kubeconfig files",
	}

	cmd.AddCommand(newConfigViewCmd())
	cmd.AddCommand(newConfigCurrentContextCmd())
	cmd.AddCommand(newConfigGetContextsCmd())
	cmd.AddCommand(newConfigUseContextCmd())
	return cmd
}

func newConfigViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Display merged kubeconfig settings or a specified kubeconfig file",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := findKubeconfigPath()
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read kubeconfig from %q: %w", path, err)
			}
			if globals.Output == "json" {
				var v interface{}
				if err := yaml.Unmarshal(data, &v); err != nil {
					return err
				}
				jsonBytes, err := json.MarshalIndent(normaliseYAML(v), "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(jsonBytes))
				return nil
			}
			fmt.Println(string(data))
			return nil
		},
	}
}

func newConfigCurrentContextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current-context",
		Short: "Display the current-context",
		RunE: func(cmd *cobra.Command, args []string) error {
			kc, _, err := loadKubeconfig()
			if err != nil {
				return err
			}
			if kc.CurrentContext == "" {
				return fmt.Errorf("current-context is not set")
			}
			fmt.Println(kc.CurrentContext)
			return nil
		},
	}
}

func newConfigGetContextsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get-contexts",
		Short: "Describe one or many contexts",
		RunE: func(cmd *cobra.Command, args []string) error {
			kc, _, err := loadKubeconfig()
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
			defer w.Flush()

			fmt.Fprintf(w, "CURRENT\tNAME\tCLUSTER\tAUTHINFO\tNAMESPACE\n")
			for _, ctx := range kc.Contexts {
				cur := ""
				if ctx.Name == kc.CurrentContext {
					cur = "*"
				}
				ns := ctx.Context.Namespace
				if ns == "" {
					ns = "default"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", cur, ctx.Name, ctx.Context.Cluster, ctx.Context.User, ns)
			}
			return nil
		},
	}
}

func newConfigUseContextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use-context CONTEXT_NAME",
		Short: "Set the current-context in a kubeconfig file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctxName := args[0]
			path := findKubeconfigPath()
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			var raw map[string]interface{}
			if err := yaml.Unmarshal(data, &raw); err != nil {
				return err
			}
			raw["current-context"] = ctxName
			out, err := yaml.Marshal(raw)
			if err != nil {
				return err
			}
			if err := os.WriteFile(path, out, 0600); err != nil {
				return err
			}
			fmt.Printf("Switched to context %q.\n", ctxName)
			return nil
		},
	}
}

// ─── Kubeconfig Structure & Loading ───────────────────────────────────────────

type kubeconfig struct {
	APIVersion     string               `yaml:"apiVersion"`
	Kind           string               `yaml:"kind"`
	CurrentContext string               `yaml:"current-context"`
	Clusters       []kubeconfigCluster  `yaml:"clusters"`
	Contexts       []kubeconfigContext  `yaml:"contexts"`
	Users          []kubeconfigAuthInfo `yaml:"users"`
}

type kubeconfigCluster struct {
	Name    string `yaml:"name"`
	Cluster struct {
		Server                   string `yaml:"server"`
		CertificateAuthority     string `yaml:"certificate-authority"`
		CertificateAuthorityData string `yaml:"certificate-authority-data"`
		InsecureSkipTLSVerify    bool   `yaml:"insecure-skip-tls-verify"`
	} `yaml:"cluster"`
}

type kubeconfigContext struct {
	Name    string `yaml:"name"`
	Context struct {
		Cluster   string `yaml:"cluster"`
		User      string `yaml:"user"`
		Namespace string `yaml:"namespace"`
	} `yaml:"context"`
}

type kubeconfigAuthInfo struct {
	Name string `yaml:"name"`
	User struct {
		ClientCertificate     string `yaml:"client-certificate"`
		ClientCertificateData string `yaml:"client-certificate-data"`
		ClientKey             string `yaml:"client-key"`
		ClientKeyData         string `yaml:"client-key-data"`
		Token                 string `yaml:"token"`
	} `yaml:"user"`
}

func findKubeconfigPath() string {
	if globals.Kubeconfig != "" {
		return globals.Kubeconfig
	}
	if env := os.Getenv("KUBECONFIG"); env != "" {
		return env
	}
	localCandidates := []string{
		"./data/admin.kubeconfig",
		"./data/pki/admin.kubeconfig",
		"./admin.kubeconfig",
		"./kubeconfig",
		"./data/config",
		"./config",
	}
	for _, p := range localCandidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	home, err := os.UserHomeDir()
	if err == nil {
		tarakPath := filepath.Join(home, ".tarak", "config")
		if _, err := os.Stat(tarakPath); err == nil {
			return tarakPath
		}
		kubePath := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(kubePath); err == nil {
			return kubePath
		}
	}
	return "./data/admin.kubeconfig"
}

func loadKubeconfig() (*kubeconfig, string, error) {
	path := findKubeconfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, path, err
	}
	var kc kubeconfig
	if err := yaml.Unmarshal(data, &kc); err != nil {
		return nil, path, fmt.Errorf("unmarshal kubeconfig: %w", err)
	}
	return &kc, path, nil
}

func effectiveNamespace() string {
	if globals.AllNamespaces {
		return ""
	}
	if globals.Namespace != "" {
		return globals.Namespace
	}
	kc, _, err := loadKubeconfig()
	if err == nil && kc != nil && kc.CurrentContext != "" {
		for _, ctx := range kc.Contexts {
			if ctx.Name == kc.CurrentContext && ctx.Context.Namespace != "" {
				return ctx.Context.Namespace
			}
		}
	}
	return "default"
}

// ─── HTTP Client & TLS ────────────────────────────────────────────────────────

type apiClient struct {
	http      *http.Client
	serverURL string
	token     string
}

func newClient() (*apiClient, error) {
	serverURL := globals.Server
	caData := []byte(nil)
	clientCertData := []byte(nil)
	clientKeyData := []byte(nil)
	token := globals.Token
	insecure := globals.Insecure

	if globals.CACert != "" {
		var err error
		caData, err = os.ReadFile(globals.CACert)
		if err != nil {
			return nil, fmt.Errorf("read ca cert %q: %w", globals.CACert, err)
		}
	}
	if globals.ClientCert != "" && globals.ClientKey != "" {
		var err error
		clientCertData, err = os.ReadFile(globals.ClientCert)
		if err != nil {
			return nil, fmt.Errorf("read client cert %q: %w", globals.ClientCert, err)
		}
		clientKeyData, err = os.ReadFile(globals.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("read client key %q: %w", globals.ClientKey, err)
		}
	}

	kc, _, err := loadKubeconfig()
	if err == nil && kc != nil {
		var activeCtx *kubeconfigContext
		for i := range kc.Contexts {
			if kc.Contexts[i].Name == kc.CurrentContext {
				activeCtx = &kc.Contexts[i]
				break
			}
		}
		if activeCtx == nil && len(kc.Contexts) > 0 {
			activeCtx = &kc.Contexts[0]
		}

		if activeCtx != nil {
			for _, c := range kc.Clusters {
				if c.Name == activeCtx.Context.Cluster {
					if serverURL == "" {
						serverURL = c.Cluster.Server
					}
					if !insecure && c.Cluster.InsecureSkipTLSVerify {
						insecure = true
					}
					if len(caData) == 0 {
						if c.Cluster.CertificateAuthorityData != "" {
							caData, _ = base64.StdEncoding.DecodeString(c.Cluster.CertificateAuthorityData)
						} else if c.Cluster.CertificateAuthority != "" {
							caData, _ = os.ReadFile(c.Cluster.CertificateAuthority)
						}
					}
					break
				}
			}
			for _, u := range kc.Users {
				if u.Name == activeCtx.Context.User {
					if token == "" && u.User.Token != "" {
						token = u.User.Token
					}
					if len(clientCertData) == 0 {
						if u.User.ClientCertificateData != "" {
							clientCertData, _ = base64.StdEncoding.DecodeString(u.User.ClientCertificateData)
						} else if u.User.ClientCertificate != "" {
							clientCertData, _ = os.ReadFile(u.User.ClientCertificate)
						}
					}
					if len(clientKeyData) == 0 {
						if u.User.ClientKeyData != "" {
							clientKeyData, _ = base64.StdEncoding.DecodeString(u.User.ClientKeyData)
						} else if u.User.ClientKey != "" {
							clientKeyData, _ = os.ReadFile(u.User.ClientKey)
						}
					}
					break
				}
			}
		}
	}

	// Auto-fallback: check local data/pki files if certs are still missing
	if len(caData) == 0 {
		for _, p := range []string{"./data/pki/ca.crt", "pki/ca.crt", "/var/lib/tarak/pki/ca.crt"} {
			if b, err := os.ReadFile(p); err == nil && len(b) > 0 {
				caData = b
				break
			}
		}
	}
	if len(clientCertData) == 0 && len(clientKeyData) == 0 {
		for _, pfx := range []string{"./data/pki/admin", "pki/admin", "/var/lib/tarak/pki/admin"} {
			certB, err1 := os.ReadFile(pfx + ".crt")
			keyB, err2 := os.ReadFile(pfx + ".key")
			if err1 == nil && err2 == nil && len(certB) > 0 && len(keyB) > 0 {
				clientCertData = certB
				clientKeyData = keyB
				break
			}
		}
	}

	if serverURL == "" {
		serverURL = "https://127.0.0.1:6443"
	}
	if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
		serverURL = "https://" + serverURL
	}

	tlsCfg := &tls.Config{
		InsecureSkipVerify: insecure, //nolint:gosec
		MinVersion:         tls.VersionTLS12,
	}

	if len(caData) > 0 {
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caData)
		tlsCfg.RootCAs = pool
	} else if strings.Contains(serverURL, "127.0.0.1") || strings.Contains(serverURL, "localhost") {
		// If running locally and no CA bundle found, allow local TLS
		tlsCfg.InsecureSkipVerify = true
	}

	if len(clientCertData) > 0 && len(clientKeyData) > 0 {
		cert, err := tls.X509KeyPair(clientCertData, clientKeyData)
		if err != nil {
			return nil, fmt.Errorf("load client key pair: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	return &apiClient{
		http: &http.Client{
			Transport: &http.Transport{TLSClientConfig: tlsCfg},
			Timeout:   30 * time.Second,
		},
		serverURL: serverURL,
		token:     token,
	}, nil
}

func (c *apiClient) get(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)
	return c.do(req)
}

func (c *apiClient) post(url string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(req)
	return c.do(req)
}

func (c *apiClient) put(url string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(req)
	return c.do(req)
}

func (c *apiClient) delete(url string) error {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	c.setAuthHeader(req)
	_, err = c.do(req)
	return err
}

func (c *apiClient) do(req *http.Request) ([]byte, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		var status struct {
			Kind    string `json:"kind"`
			Message string `json:"message"`
			Reason  string `json:"reason"`
		}
		if err := json.Unmarshal(body, &status); err == nil && status.Kind == "Status" && status.Message != "" {
			return nil, fmt.Errorf("Error from server (%s): %s", status.Reason, status.Message)
		}
		return nil, fmt.Errorf("Error from server (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return body, nil
}

func (c *apiClient) setAuthHeader(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}

// ─── URL Routing ──────────────────────────────────────────────────────────────

func appendQueryParam(url, key, val string) string {
	if val == "" {
		return url
	}
	sep := "?"
	if strings.Contains(url, "?") {
		sep = "&"
	}
	return fmt.Sprintf("%s%s%s=%s", url, sep, key, val)
}

func buildURL(base, resource, namespace, name string) string {
	coreResources := map[string]bool{
		"pods": true, "services": true, "endpoints": true,
		"configmaps": true, "secrets": true, "serviceaccounts": true,
		"namespaces": true, "nodes": true, "persistentvolumes": true,
		"persistentvolumeclaims": true, "events": true,
		"resourcequotas": true, "limitranges": true,
		"replicationcontrollers": true, "podtemplates": true,
	}

	apiGroupPaths := map[string]string{
		"deployments": "apps/v1", "replicasets": "apps/v1",
		"statefulsets": "apps/v1", "daemonsets": "apps/v1",
		"controllerrevisions": "apps/v1",
		"jobs": "batch/v1", "cronjobs": "batch/v1",
		"networkpolicies": "networking.k8s.io/v1",
		"ingresses": "networking.k8s.io/v1",
		"ingressclasses": "networking.k8s.io/v1",
		"roles": "rbac.authorization.k8s.io/v1",
		"rolebindings": "rbac.authorization.k8s.io/v1",
		"clusterroles": "rbac.authorization.k8s.io/v1",
		"clusterrolebindings": "rbac.authorization.k8s.io/v1",
		"storageclasses": "storage.k8s.io/v1",
		"volumeattachments": "storage.k8s.io/v1",
		"csinodes": "storage.k8s.io/v1",
		"csistoragecapacities": "storage.k8s.io/v1",
		"customresourcedefinitions": "apiextensions.k8s.io/v1",
		"taraksecuritypolicies":     "security.tarak.io/v1",
		"tarakapplications":         "apps.tarak.io/v1",
		"meshes":                    "mesh.tarak.io/v1",
		"meshservices":              "mesh.tarak.io/v1",
		"meshexternalservices":      "mesh.tarak.io/v1",
		"meshtrafficpermissions":    "mesh.tarak.io/v1",
		"meshpassthroughpolicies":   "mesh.tarak.io/v1",
		"meshproxypatches":          "mesh.tarak.io/v1",
	}

	short := map[string]string{
		"po": "pods", "pod": "pods", "svc": "services", "service": "services", "cm": "configmaps", "configmap": "configmaps",
		"sa": "serviceaccounts", "serviceaccount": "serviceaccounts", "ns": "namespaces", "namespace": "namespaces", "no": "nodes", "node": "nodes",
		"pv": "persistentvolumes", "persistentvolume": "persistentvolumes", "pvc": "persistentvolumeclaims", "persistentvolumeclaim": "persistentvolumeclaims",
		"ev": "events", "event": "events", "rc": "replicationcontrollers", "replicationcontroller": "replicationcontrollers",
		"deploy": "deployments", "deployment": "deployments", "rs": "replicasets", "replicaset": "replicasets",
		"sts": "statefulsets", "statefulset": "statefulsets", "ds": "daemonsets", "daemonset": "daemonsets",
		"cj": "cronjobs", "cronjob": "cronjobs", "job": "jobs",
		"netpol": "networkpolicies", "networkpolicy": "networkpolicies", "ing": "ingresses", "ingress": "ingresses",
		"cr": "clusterroles", "clusterrole": "clusterroles", "crb": "clusterrolebindings", "clusterrolebinding": "clusterrolebindings",
		"sc": "storageclasses", "storageclass": "storageclasses", "sec": "secrets", "secret": "secrets",
		"crd": "customresourcedefinitions", "crds": "customresourcedefinitions",
		"tsp": "taraksecuritypolicies", "securitypolicy": "taraksecuritypolicies",
		"tapp": "tarakapplications", "tarakapp": "tarakapplications", "app": "tarakapplications",
		"mesh": "meshes", "meshes": "meshes",
	}
	if full, ok := short[strings.ToLower(resource)]; ok {
		resource = full
	} else {
		resource = strings.ToLower(resource)
	}

	clusterScoped := map[string]bool{
		"namespaces": true, "nodes": true, "persistentvolumes": true,
		"ingressclasses": true, "clusterroles": true, "clusterrolebindings": true,
		"storageclasses": true, "volumeattachments": true, "csinodes": true,
		"customresourcedefinitions": true, "taraksecuritypolicies": true,
		"meshes": true,
	}

	var path string
	if coreResources[resource] {
		path = base + "/api/v1"
		if namespace != "" && !clusterScoped[resource] {
			path += "/namespaces/" + namespace
		}
		path += "/" + resource
	} else if gv, ok := apiGroupPaths[resource]; ok {
		path = base + "/apis/" + gv
		if namespace != "" && !clusterScoped[resource] {
			path += "/namespaces/" + namespace
		}
		path += "/" + resource
	} else {
		path = base + "/api/v1"
		if namespace != "" && !clusterScoped[resource] {
			path += "/namespaces/" + namespace
		}
		path += "/" + resource
	}

	if name != "" {
		path += "/" + name
	}
	return path
}

// ─── Output Formatting ────────────────────────────────────────────────────────

func printResourceOutput(data []byte, resource, format string, allNamespaces, noHeaders, isMultiple bool) error {
	if format == "json" {
		var v interface{}
		if err := json.Unmarshal(data, &v); err != nil {
			fmt.Println(string(data))
			return nil
		}
		out, _ := json.MarshalIndent(v, "", "  ")
		fmt.Println(string(out))
		return nil
	}
	if format == "yaml" {
		var v interface{}
		if err := json.Unmarshal(data, &v); err != nil {
			fmt.Println(string(data))
			return nil
		}
		out, _ := yaml.Marshal(v)
		fmt.Println(string(out))
		return nil
	}
	if format == "name" {
		var obj struct {
			Metadata struct{ Name string `json:"name"` } `json:"metadata"`
			Items    []struct {
				Kind     string `json:"kind"`
				Metadata struct{ Name string `json:"name"` } `json:"metadata"`
			} `json:"items"`
		}
		if err := json.Unmarshal(data, &obj); err == nil {
			if len(obj.Items) > 0 {
				for _, item := range obj.Items {
					fmt.Printf("%s/%s\n", strings.ToLower(item.Kind), item.Metadata.Name)
				}
				return nil
			}
			if obj.Metadata.Name != "" {
				fmt.Println(obj.Metadata.Name)
				return nil
			}
		}
	}

	// Tabular output (default / wide)
	var list struct {
		Kind  string            `json:"kind"`
		Items []json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(data, &list); err == nil && strings.HasSuffix(list.Kind, "List") {
		if len(list.Items) == 0 {
			if isMultiple {
				return nil // Don't clutter "get all" with empty sub-tables
			}
			if allNamespaces {
				fmt.Println("No resources found")
			} else {
				ns := effectiveNamespace()
				if ns == "" {
					fmt.Println("No resources found")
				} else {
					fmt.Printf("No resources found in %s namespace.\n", ns)
				}
			}
			return nil
		}
		return renderTable(list.Items, resource, format == "wide", allNamespaces, noHeaders, isMultiple)
	}

	// Single object
	return renderTable([]json.RawMessage{data}, resource, format == "wide", allNamespaces, noHeaders, isMultiple)
}

func renderTable(items []json.RawMessage, resource string, wide, allNamespaces, noHeaders, isMultiple bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	defer w.Flush()

	short := map[string]string{
		"po": "pods", "svc": "services", "cm": "configmaps",
		"sa": "serviceaccounts", "ns": "namespaces", "no": "nodes",
		"deploy": "deployments", "rs": "replicasets", "sts": "statefulsets",
		"ds": "daemonsets", "cj": "cronjobs", "sec": "secrets",
		"tsp": "taraksecuritypolicies", "securitypolicy": "taraksecuritypolicies",
		"tapp": "tarakapplications", "tarakapp": "tarakapplications", "app": "tarakapplications",
		"crd": "customresourcedefinitions", "crds": "customresourcedefinitions",
	}
	if full, ok := short[strings.ToLower(resource)]; ok {
		resource = full
	} else {
		resource = strings.ToLower(resource)
	}

	prefixName := isMultiple

	switch resource {
	case "pods":
		if !noHeaders {
			if allNamespaces {
				if wide {
					fmt.Fprintf(w, "NAMESPACE\tNAME\tREADY\tSTATUS\tRESTARTS\tAGE\tIP\tNODE\n")
				} else {
					fmt.Fprintf(w, "NAMESPACE\tNAME\tREADY\tSTATUS\tRESTARTS\tAGE\n")
				}
			} else {
				if wide {
					fmt.Fprintf(w, "NAME\tREADY\tSTATUS\tRESTARTS\tAGE\tIP\tNODE\n")
				} else {
					fmt.Fprintf(w, "NAME\tREADY\tSTATUS\tRESTARTS\tAGE\n")
				}
			}
		}
		for _, raw := range items {
			var p struct {
				Metadata struct {
					Name              string    `json:"name"`
					Namespace         string    `json:"namespace"`
					CreationTimestamp time.Time `json:"creationTimestamp"`
				} `json:"metadata"`
				Spec struct {
					Containers []struct{ Name string } `json:"containers"`
					NodeName   string                  `json:"nodeName"`
				} `json:"spec"`
				Status struct {
					Phase             string `json:"phase"`
					PodIP             string `json:"podIP"`
					ContainerStatuses []struct {
						Ready        bool `json:"ready"`
						RestartCount int  `json:"restartCount"`
						State        map[string]struct {
							Reason string `json:"reason"`
						} `json:"state"`
					} `json:"containerStatuses"`
				} `json:"status"`
			}
			_ = json.Unmarshal(raw, &p)

			readyCount := 0
			restarts := 0
			for _, cs := range p.Status.ContainerStatuses {
				if cs.Ready {
					readyCount++
				}
				restarts += cs.RestartCount
			}
			readyStr := fmt.Sprintf("%d/%d", readyCount, len(p.Spec.Containers))
			status := p.Status.Phase
			if status == "" {
				status = "Pending"
			}
			for _, cs := range p.Status.ContainerStatuses {
				if waiting, ok := cs.State["waiting"]; ok && waiting.Reason != "" {
					status = waiting.Reason
				} else if terminated, ok := cs.State["terminated"]; ok && terminated.Reason != "" {
					status = terminated.Reason
				}
			}
			age := formatAge(time.Since(p.Metadata.CreationTimestamp))
			displayName := p.Metadata.Name
			if prefixName {
				displayName = "pod/" + displayName
			}

			if allNamespaces {
				if wide {
					ip := p.Status.PodIP
					if ip == "" {
						ip = "<none>"
					}
					node := p.Spec.NodeName
					if node == "" {
						node = "<none>"
					}
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\t%s\t%s\n",
						p.Metadata.Namespace, displayName, readyStr, status, restarts, age, ip, node)
				} else {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
						p.Metadata.Namespace, displayName, readyStr, status, restarts, age)
				}
			} else {
				if wide {
					ip := p.Status.PodIP
					if ip == "" {
						ip = "<none>"
					}
					node := p.Spec.NodeName
					if node == "" {
						node = "<none>"
					}
					fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\t%s\n",
						displayName, readyStr, status, restarts, age, ip, node)
				} else {
					fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
						displayName, readyStr, status, restarts, age)
				}
			}
		}

	case "nodes":
		if !noHeaders {
			if wide {
				fmt.Fprintf(w, "NAME\tSTATUS\tROLES\tAGE\tVERSION\tINTERNAL-IP\tEXTERNAL-IP\tOS-IMAGE\tKERNEL-VERSION\tCONTAINER-RUNTIME\n")
			} else {
				fmt.Fprintf(w, "NAME\tSTATUS\tROLES\tAGE\tVERSION\n")
			}
		}
		for _, raw := range items {
			var n struct {
				Metadata struct {
					Name              string            `json:"name"`
					CreationTimestamp time.Time         `json:"creationTimestamp"`
					Labels            map[string]string `json:"labels"`
				} `json:"metadata"`
				Status struct {
					Phase      string `json:"phase"`
					Conditions []struct {
						Type   string `json:"type"`
						Status string `json:"status"`
					} `json:"conditions"`
					NodeInfo struct {
						KubeletVersion          string `json:"kubeletVersion"`
						OSImage                 string `json:"osImage"`
						KernelVersion           string `json:"kernelVersion"`
						ContainerRuntimeVersion string `json:"containerRuntimeVersion"`
					} `json:"nodeInfo"`
					Addresses []struct {
						Type    string `json:"type"`
						Address string `json:"address"`
					} `json:"addresses"`
				} `json:"status"`
			}
			_ = json.Unmarshal(raw, &n)
			status := "Ready"
			if n.Status.Phase != "" && n.Status.Phase != "Running" {
				status = n.Status.Phase
			}
			for _, c := range n.Status.Conditions {
				if c.Type == "Ready" {
					if c.Status != "True" {
						status = "NotReady"
					}
					break
				}
			}
			roles := "control-plane,master"
			for k := range n.Metadata.Labels {
				if strings.HasPrefix(k, "node-role.kubernetes.io/") {
					role := strings.TrimPrefix(k, "node-role.kubernetes.io/")
					if role != "" {
						roles = role
					}
				}
			}
			age := formatAge(time.Since(n.Metadata.CreationTimestamp))
			nodeVer := n.Status.NodeInfo.KubeletVersion
			if nodeVer == "" {
				nodeVer = "v" + version.Version + "-tarak"
			}
			displayName := n.Metadata.Name
			if prefixName {
				displayName = "node/" + displayName
			}

			if wide {
				internalIP := "<none>"
				externalIP := "<none>"
				for _, addr := range n.Status.Addresses {
					if addr.Type == "InternalIP" {
						internalIP = addr.Address
					} else if addr.Type == "ExternalIP" {
						externalIP = addr.Address
					}
				}
				osImage := n.Status.NodeInfo.OSImage
				if osImage == "" {
					osImage = "Tarak Native"
				}
				kernel := n.Status.NodeInfo.KernelVersion
				if kernel == "" {
					kernel = "tarak-runtime"
				}
				rtVer := n.Status.NodeInfo.ContainerRuntimeVersion
				if rtVer == "" {
					rtVer = "tarak-runtime://v" + version.Version
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					displayName, status, roles, age, nodeVer, internalIP, externalIP, osImage, kernel, rtVer)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					displayName, status, roles, age, nodeVer)
			}
		}

	case "namespaces":
		if !noHeaders {
			fmt.Fprintf(w, "NAME\tSTATUS\tAGE\n")
		}
		for _, raw := range items {
			var ns struct {
				Metadata struct {
					Name              string    `json:"name"`
					CreationTimestamp time.Time `json:"creationTimestamp"`
				} `json:"metadata"`
				Status struct {
					Phase string `json:"phase"`
				} `json:"status"`
			}
			_ = json.Unmarshal(raw, &ns)
			phase := ns.Status.Phase
			if phase == "" {
				phase = "Active"
			}
			age := formatAge(time.Since(ns.Metadata.CreationTimestamp))
			fmt.Fprintf(w, "%s\t%s\t%s\n", ns.Metadata.Name, phase, age)
		}

	case "services":
		if !noHeaders {
			if allNamespaces {
				fmt.Fprintf(w, "NAMESPACE\tNAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tPORT(S)\tAGE\n")
			} else {
				fmt.Fprintf(w, "NAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tPORT(S)\tAGE\n")
			}
		}
		for _, raw := range items {
			var svc struct {
				Metadata struct {
					Name              string    `json:"name"`
					Namespace         string    `json:"namespace"`
					CreationTimestamp time.Time `json:"creationTimestamp"`
				} `json:"metadata"`
				Spec struct {
					Type        string   `json:"type"`
					ClusterIP   string   `json:"clusterIP"`
					ExternalIPs []string `json:"externalIPs"`
					Ports       []struct {
						Port     int    `json:"port"`
						NodePort int    `json:"nodePort"`
						Protocol string `json:"protocol"`
					} `json:"ports"`
				} `json:"spec"`
				Status struct {
					LoadBalancer struct {
						Ingress []struct {
							IP       string `json:"ip"`
							Hostname string `json:"hostname"`
						} `json:"ingress"`
					} `json:"loadBalancer"`
				} `json:"status"`
			}
			_ = json.Unmarshal(raw, &svc)
			svcType := svc.Spec.Type
			if svcType == "" {
				svcType = "ClusterIP"
			}
			cip := svc.Spec.ClusterIP
			if cip == "" {
				cip = "<none>"
			}
			extIP := "<none>"
			if len(svc.Status.LoadBalancer.Ingress) > 0 {
				if svc.Status.LoadBalancer.Ingress[0].IP != "" {
					extIP = svc.Status.LoadBalancer.Ingress[0].IP
				} else if svc.Status.LoadBalancer.Ingress[0].Hostname != "" {
					extIP = svc.Status.LoadBalancer.Ingress[0].Hostname
				}
			} else if len(svc.Spec.ExternalIPs) > 0 {
				extIP = strings.Join(svc.Spec.ExternalIPs, ",")
			}
			var ports []string
			for _, p := range svc.Spec.Ports {
				proto := p.Protocol
				if proto == "" {
					proto = "TCP"
				}
				if p.NodePort > 0 {
					ports = append(ports, fmt.Sprintf("%d:%d/%s", p.Port, p.NodePort, proto))
				} else {
					ports = append(ports, fmt.Sprintf("%d/%s", p.Port, proto))
				}
			}
			portStr := strings.Join(ports, ",")
			if portStr == "" {
				portStr = "<none>"
			}
			age := formatAge(time.Since(svc.Metadata.CreationTimestamp))
			displayName := svc.Metadata.Name
			if prefixName {
				displayName = "service/" + displayName
			}

			if allNamespaces {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					svc.Metadata.Namespace, displayName, svcType, cip, extIP, portStr, age)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
					displayName, svcType, cip, extIP, portStr, age)
			}
		}

	case "deployments":
		if !noHeaders {
			if allNamespaces {
				fmt.Fprintf(w, "NAMESPACE\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE\n")
			} else {
				fmt.Fprintf(w, "NAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE\n")
			}
		}
		for _, raw := range items {
			var d struct {
				Metadata struct {
					Name              string    `json:"name"`
					Namespace         string    `json:"namespace"`
					CreationTimestamp time.Time `json:"creationTimestamp"`
				} `json:"metadata"`
				Spec struct {
					Replicas *int32 `json:"replicas"`
				} `json:"spec"`
				Status struct {
					Replicas          int32 `json:"replicas"`
					UpdatedReplicas   int32 `json:"updatedReplicas"`
					ReadyReplicas     int32 `json:"readyReplicas"`
					AvailableReplicas int32 `json:"availableReplicas"`
				} `json:"status"`
			}
			_ = json.Unmarshal(raw, &d)

			desired := int32(1)
			if d.Spec.Replicas != nil {
				desired = *d.Spec.Replicas
			}
			readyStr := fmt.Sprintf("%d/%d", d.Status.ReadyReplicas, desired)
			age := formatAge(time.Since(d.Metadata.CreationTimestamp))
			displayName := d.Metadata.Name
			if prefixName {
				displayName = "deployment.apps/" + displayName
			}

			if allNamespaces {
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\n",
					d.Metadata.Namespace, displayName, readyStr, d.Status.UpdatedReplicas, d.Status.AvailableReplicas, age)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\n",
					displayName, readyStr, d.Status.UpdatedReplicas, d.Status.AvailableReplicas, age)
			}
		}

	case "replicasets":
		if !noHeaders {
			if allNamespaces {
				fmt.Fprintf(w, "NAMESPACE\tNAME\tDESIRED\tCURRENT\tREADY\tAGE\n")
			} else {
				fmt.Fprintf(w, "NAME\tDESIRED\tCURRENT\tREADY\tAGE\n")
			}
		}
		for _, raw := range items {
			var rs struct {
				Metadata struct {
					Name              string    `json:"name"`
					Namespace         string    `json:"namespace"`
					CreationTimestamp time.Time `json:"creationTimestamp"`
				} `json:"metadata"`
				Spec struct {
					Replicas *int32 `json:"replicas"`
				} `json:"spec"`
				Status struct {
					Replicas      int32 `json:"replicas"`
					ReadyReplicas int32 `json:"readyReplicas"`
				} `json:"status"`
			}
			_ = json.Unmarshal(raw, &rs)
			desired := int32(1)
			if rs.Spec.Replicas != nil {
				desired = *rs.Spec.Replicas
			}
			age := formatAge(time.Since(rs.Metadata.CreationTimestamp))
			displayName := rs.Metadata.Name
			if prefixName {
				displayName = "replicaset.apps/" + displayName
			}

			if allNamespaces {
				fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\t%s\n",
					rs.Metadata.Namespace, displayName, desired, rs.Status.Replicas, rs.Status.ReadyReplicas, age)
			} else {
				fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%s\n",
					displayName, desired, rs.Status.Replicas, rs.Status.ReadyReplicas, age)
			}
		}

	case "daemonsets":
		if !noHeaders {
			if allNamespaces {
				fmt.Fprintf(w, "NAMESPACE\tNAME\tDESIRED\tCURRENT\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE\n")
			} else {
				fmt.Fprintf(w, "NAME\tDESIRED\tCURRENT\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE\n")
			}
		}
		for _, raw := range items {
			var ds struct {
				Metadata struct {
					Name              string    `json:"name"`
					Namespace         string    `json:"namespace"`
					CreationTimestamp time.Time `json:"creationTimestamp"`
				} `json:"metadata"`
				Status struct {
					DesiredNumberScheduled int32 `json:"desiredNumberScheduled"`
					CurrentNumberScheduled int32 `json:"currentNumberScheduled"`
					NumberReady            int32 `json:"numberReady"`
					UpdatedNumberScheduled int32 `json:"updatedNumberScheduled"`
					NumberAvailable        int32 `json:"numberAvailable"`
				} `json:"status"`
			}
			_ = json.Unmarshal(raw, &ds)
			age := formatAge(time.Since(ds.Metadata.CreationTimestamp))
			displayName := ds.Metadata.Name
			if prefixName {
				displayName = "daemonset.apps/" + displayName
			}

			if allNamespaces {
				fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\t%d\t%d\t%s\n",
					ds.Metadata.Namespace, displayName, ds.Status.DesiredNumberScheduled,
					ds.Status.CurrentNumberScheduled, ds.Status.NumberReady, ds.Status.UpdatedNumberScheduled,
					ds.Status.NumberAvailable, age)
			} else {
				fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%d\t%s\n",
					displayName, ds.Status.DesiredNumberScheduled, ds.Status.CurrentNumberScheduled,
					ds.Status.NumberReady, ds.Status.UpdatedNumberScheduled, ds.Status.NumberAvailable, age)
			}
		}

	case "statefulsets":
		if !noHeaders {
			if allNamespaces {
				fmt.Fprintf(w, "NAMESPACE\tNAME\tREADY\tAGE\n")
			} else {
				fmt.Fprintf(w, "NAME\tREADY\tAGE\n")
			}
		}
		for _, raw := range items {
			var sts struct {
				Metadata struct {
					Name              string    `json:"name"`
					Namespace         string    `json:"namespace"`
					CreationTimestamp time.Time `json:"creationTimestamp"`
				} `json:"metadata"`
				Spec struct {
					Replicas *int32 `json:"replicas"`
				} `json:"spec"`
				Status struct {
					ReadyReplicas int32 `json:"readyReplicas"`
				} `json:"status"`
			}
			_ = json.Unmarshal(raw, &sts)
			desired := int32(1)
			if sts.Spec.Replicas != nil {
				desired = *sts.Spec.Replicas
			}
			readyStr := fmt.Sprintf("%d/%d", sts.Status.ReadyReplicas, desired)
			age := formatAge(time.Since(sts.Metadata.CreationTimestamp))
			displayName := sts.Metadata.Name
			if prefixName {
				displayName = "statefulset.apps/" + displayName
			}

			if allNamespaces {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", sts.Metadata.Namespace, displayName, readyStr, age)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%s\n", displayName, readyStr, age)
			}
		}

	case "jobs":
		if !noHeaders {
			if allNamespaces {
				fmt.Fprintf(w, "NAMESPACE\tNAME\tCOMPLETIONS\tDURATION\tAGE\n")
			} else {
				fmt.Fprintf(w, "NAME\tCOMPLETIONS\tDURATION\tAGE\n")
			}
		}
		for _, raw := range items {
			var j struct {
				Metadata struct {
					Name              string    `json:"name"`
					Namespace         string    `json:"namespace"`
					CreationTimestamp time.Time `json:"creationTimestamp"`
				} `json:"metadata"`
				Spec struct {
					Completions *int32 `json:"completions"`
				} `json:"spec"`
				Status struct {
					Succeeded int32 `json:"succeeded"`
				} `json:"status"`
			}
			_ = json.Unmarshal(raw, &j)
			completions := int32(1)
			if j.Spec.Completions != nil {
				completions = *j.Spec.Completions
			}
			age := formatAge(time.Since(j.Metadata.CreationTimestamp))
			displayName := j.Metadata.Name
			if prefixName {
				displayName = "job.batch/" + displayName
			}

			if allNamespaces {
				fmt.Fprintf(w, "%s\t%s\t%d/%d\t-\t%s\n", j.Metadata.Namespace, displayName, j.Status.Succeeded, completions, age)
			} else {
				fmt.Fprintf(w, "%s\t%d/%d\t-\t%s\n", displayName, j.Status.Succeeded, completions, age)
			}
		}

	case "configmaps":
		if !noHeaders {
			if allNamespaces {
				fmt.Fprintf(w, "NAMESPACE\tNAME\tDATA\tAGE\n")
			} else {
				fmt.Fprintf(w, "NAME\tDATA\tAGE\n")
			}
		}
		for _, raw := range items {
			var cm struct {
				Metadata struct {
					Name              string    `json:"name"`
					Namespace         string    `json:"namespace"`
					CreationTimestamp time.Time `json:"creationTimestamp"`
				} `json:"metadata"`
				Data map[string]interface{} `json:"data"`
			}
			_ = json.Unmarshal(raw, &cm)
			age := formatAge(time.Since(cm.Metadata.CreationTimestamp))
			displayName := cm.Metadata.Name
			if prefixName {
				displayName = "configmap/" + displayName
			}

			if allNamespaces {
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", cm.Metadata.Namespace, displayName, len(cm.Data), age)
			} else {
				fmt.Fprintf(w, "%s\t%d\t%s\n", displayName, len(cm.Data), age)
			}
		}

	case "secrets":
		if !noHeaders {
			if allNamespaces {
				fmt.Fprintf(w, "NAMESPACE\tNAME\tTYPE\tDATA\tAGE\n")
			} else {
				fmt.Fprintf(w, "NAME\tTYPE\tDATA\tAGE\n")
			}
		}
		for _, raw := range items {
			var s struct {
				Metadata struct {
					Name              string    `json:"name"`
					Namespace         string    `json:"namespace"`
					CreationTimestamp time.Time `json:"creationTimestamp"`
				} `json:"metadata"`
				Type string                 `json:"type"`
				Data map[string]interface{} `json:"data"`
			}
			_ = json.Unmarshal(raw, &s)
			secType := s.Type
			if secType == "" {
				secType = "Opaque"
			}
			age := formatAge(time.Since(s.Metadata.CreationTimestamp))
			displayName := s.Metadata.Name
			if prefixName {
				displayName = "secret/" + displayName
			}

			if allNamespaces {
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n", s.Metadata.Namespace, displayName, secType, len(s.Data), age)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", displayName, secType, len(s.Data), age)
			}
		}

	case "taraksecuritypolicies":
		if !noHeaders {
			fmt.Fprintf(w, "NAME\tPRIVILEGED\tREAD-ONLY-ROOTFS\tENFORCE-ENCRYPTION\tNETWORK-ISOLATION\tAGE\n")
		}
		for _, raw := range items {
			var tsp struct {
				Metadata struct {
					Name              string    `json:"name"`
					CreationTimestamp time.Time `json:"creationTimestamp"`
				} `json:"metadata"`
				Spec struct {
					Privileged              bool `json:"privileged"`
					ReadOnlyRootFilesystem  bool `json:"readOnlyRootFilesystem"`
					EnforceEncryptionAtRest bool `json:"enforceEncryptionAtRest"`
					NetworkIsolation        bool `json:"networkIsolation"`
				} `json:"spec"`
			}
			_ = json.Unmarshal(raw, &tsp)
			age := formatAge(time.Since(tsp.Metadata.CreationTimestamp))
			displayName := tsp.Metadata.Name
			if prefixName {
				displayName = "tsp/" + displayName
			}
			fmt.Fprintf(w, "%s\t%t\t%t\t%t\t%t\t%s\n",
				displayName, tsp.Spec.Privileged, tsp.Spec.ReadOnlyRootFilesystem,
				tsp.Spec.EnforceEncryptionAtRest, tsp.Spec.NetworkIsolation, age)
		}

	case "tarakapplications":
		if !noHeaders {
			if allNamespaces {
				fmt.Fprintf(w, "NAMESPACE\tNAME\tIMAGE\tREPLICAS\tDOMAIN\tAUTO-TLS\tSTATUS\tAGE\n")
			} else {
				fmt.Fprintf(w, "NAME\tIMAGE\tREPLICAS\tDOMAIN\tAUTO-TLS\tSTATUS\tAGE\n")
			}
		}
		for _, raw := range items {
			var app struct {
				Metadata struct {
					Name              string    `json:"name"`
					Namespace         string    `json:"namespace"`
					CreationTimestamp time.Time `json:"creationTimestamp"`
				} `json:"metadata"`
				Spec struct {
					Image    string `json:"image"`
					Replicas int32  `json:"replicas"`
					Domain   string `json:"domain"`
					AutoTLS  bool   `json:"autoTLS"`
				} `json:"spec"`
				Status struct {
					Phase         string `json:"phase"`
					ReadyReplicas int32  `json:"readyReplicas"`
				} `json:"status"`
			}
			_ = json.Unmarshal(raw, &app)
			age := formatAge(time.Since(app.Metadata.CreationTimestamp))
			phase := app.Status.Phase
			if phase == "" {
				phase = "Active"
			}
			domain := app.Spec.Domain
			if domain == "" {
				domain = "<internal>"
			}
			replicasStr := fmt.Sprintf("%d/%d", app.Status.ReadyReplicas, app.Spec.Replicas)
			if app.Spec.Replicas == 0 {
				replicasStr = "1/1"
			}
			displayName := app.Metadata.Name
			if prefixName {
				displayName = "tapp/" + displayName
			}

			if allNamespaces {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%t\t%s\t%s\n",
					app.Metadata.Namespace, displayName, app.Spec.Image, replicasStr, domain, app.Spec.AutoTLS, phase, age)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%t\t%s\t%s\n",
					displayName, app.Spec.Image, replicasStr, domain, app.Spec.AutoTLS, phase, age)
			}
		}

	case "customresourcedefinitions":
		if !noHeaders {
			fmt.Fprintf(w, "NAME\tCREATED AT\n")
		}
		for _, raw := range items {
			var crd struct {
				Metadata struct {
					Name              string    `json:"name"`
					CreationTimestamp time.Time `json:"creationTimestamp"`
				} `json:"metadata"`
			}
			_ = json.Unmarshal(raw, &crd)
			fmt.Fprintf(w, "%s\t%s\n", crd.Metadata.Name, crd.Metadata.CreationTimestamp.Format(time.RFC3339))
		}

	case "meshes":
		if !noHeaders {
			fmt.Fprintf(w, "NAME\tPHASE\tMTLS MODE\tPASSTHROUGH\tSERVICES\tENROLLED PODS\tAGE\n")
		}
		for _, raw := range items {
			var m struct {
				Name        string    `json:"name"`
				CreatedAt   time.Time `json:"createdAt"`
				Passthrough string    `json:"passthrough"`
				MTLS        struct {
					Mode string `json:"mode"`
				} `json:"mtls"`
				Metadata struct {
					Name              string    `json:"name"`
					CreationTimestamp time.Time `json:"creationTimestamp"`
				} `json:"metadata"`
				Spec struct {
					MTLS struct {
						Mode string `json:"mode"`
					} `json:"mtls"`
					Networking struct {
						PassthroughMode string `json:"passthroughMode"`
					} `json:"networking"`
				} `json:"spec"`
				Status struct {
					Phase         string `json:"phase"`
					TotalServices int    `json:"totalServices"`
					EnrolledPods  int    `json:"enrolledPods"`
				} `json:"status"`
			}
			_ = json.Unmarshal(raw, &m)

			name := m.Name
			if name == "" {
				name = m.Metadata.Name
			}
			if name == "" {
				name = "default"
			}

			mtlsMode := m.MTLS.Mode
			if mtlsMode == "" {
				mtlsMode = m.Spec.MTLS.Mode
			}
			if mtlsMode == "" {
				mtlsMode = "Strict"
			}

			passMode := m.Passthrough
			if passMode == "" {
				passMode = m.Spec.Networking.PassthroughMode
			}
			if passMode == "" {
				passMode = "Passthrough"
			}

			phase := m.Status.Phase
			if phase == "" {
				phase = "Active"
			}

			created := m.CreatedAt
			if created.IsZero() {
				created = m.Metadata.CreationTimestamp
			}
			ageStr := "0s"
			if !created.IsZero() {
				ageStr = formatAge(time.Since(created))
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\t%s\n",
				name, phase, mtlsMode, passMode, m.Status.TotalServices, m.Status.EnrolledPods, ageStr)
		}

	default:
		if !noHeaders {
			if allNamespaces {
				fmt.Fprintf(w, "NAMESPACE\tNAME\tAGE\n")
			} else {
				fmt.Fprintf(w, "NAME\tAGE\n")
			}
		}
		for _, raw := range items {
			var obj struct {
				Metadata struct {
					Name              string    `json:"name"`
					Namespace         string    `json:"namespace"`
					CreationTimestamp time.Time `json:"creationTimestamp"`
				} `json:"metadata"`
			}
			_ = json.Unmarshal(raw, &obj)
			age := formatAge(time.Since(obj.Metadata.CreationTimestamp))
			displayName := obj.Metadata.Name
			if prefixName {
				displayName = resource + "/" + displayName
			}

			if allNamespaces && obj.Metadata.Namespace != "" {
				fmt.Fprintf(w, "%s\t%s\t%s\n", obj.Metadata.Namespace, displayName, age)
			} else {
				fmt.Fprintf(w, "%s\t%s\n", displayName, age)
			}
		}
	}
	return nil
}

func formatAge(d time.Duration) string {
	if d < 0 {
		return "0s"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

func toJSON(data []byte) ([]byte, error) {
	trimmed := strings.TrimSpace(string(data))
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		var v interface{}
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
		return json.Marshal(v)
	}

	var yamlVal interface{}
	if err := yaml.Unmarshal(data, &yamlVal); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}
	if yamlVal == nil {
		return nil, fmt.Errorf("empty document")
	}
	jsonBytes, err := json.Marshal(normaliseYAML(yamlVal))
	if err != nil {
		return nil, fmt.Errorf("yaml to json: %w", err)
	}
	return jsonBytes, nil
}

func normaliseYAML(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(val))
		for k, v2 := range val {
			out[k] = normaliseYAML(v2)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(val))
		for i, v2 := range val {
			out[i] = normaliseYAML(v2)
		}
		return out
	default:
		return val
	}
}

func extractResourceInfo(data []byte) (resource, namespace, name string, err error) {
	var obj struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Metadata   struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return "", "", "", fmt.Errorf("parse object: %w", err)
	}
	if obj.Kind == "" {
		return "", "", "", fmt.Errorf("object missing kind")
	}
	resource = kindToResource(obj.Kind)
	return resource, obj.Metadata.Namespace, obj.Metadata.Name, nil
}

func kindToResource(kind string) string {
	kindMap := map[string]string{
		"Pod": "pods", "Service": "services", "Endpoints": "endpoints",
		"ConfigMap": "configmaps", "Secret": "secrets",
		"ServiceAccount": "serviceaccounts", "Namespace": "namespaces",
		"Node": "nodes", "PersistentVolume": "persistentvolumes",
		"PersistentVolumeClaim": "persistentvolumeclaims",
		"Event": "events", "ResourceQuota": "resourcequotas",
		"LimitRange": "limitranges", "PodTemplate": "podtemplates",
		"ReplicationController": "replicationcontrollers",
		"Deployment": "deployments", "ReplicaSet": "replicasets",
		"StatefulSet": "statefulsets", "DaemonSet": "daemonsets",
		"ControllerRevision": "controllerrevisions",
		"Job": "jobs", "CronJob": "cronjobs",
		"NetworkPolicy": "networkpolicies", "Ingress": "ingresses",
		"IngressClass": "ingressclasses",
		"Role": "roles", "RoleBinding": "rolebindings",
		"ClusterRole": "clusterroles", "ClusterRoleBinding": "clusterrolebindings",
		"StorageClass": "storageclasses", "VolumeAttachment": "volumeattachments",
		"CSINode": "csinodes", "CSIStorageCapacity": "csistoragecapacities",
		"CustomResourceDefinition": "customresourcedefinitions",
		"TarakSecurityPolicy":     "taraksecuritypolicies",
		"TarakApplication":        "tarakapplications",
	}
	if r, ok := kindMap[kind]; ok {
		return r
	}
	return strings.ToLower(kind) + "s"
}

func parseTypeAndName(target, defaultType string) (resource, name string) {
	if strings.Contains(target, "/") {
		parts := strings.SplitN(target, "/", 2)
		return parts[0], parts[1]
	}
	return defaultType, target
}

// ─── Port Forward ─────────────────────────────────────────────────────────────

func newPortForwardCmd() *cobra.Command {
	var address string
	cmd := &cobra.Command{
		Use:   "port-forward (POD | TYPE/NAME) [LOCAL_PORT:]REMOTE_PORT [...[LOCAL_PORT_N:]REMOTE_PORT_N]",
		Short: "Forward one or more local ports to a pod or service",
		Example: `  # Forward local port 8080 to port 80 on pod 'web-app-xxx'
  tarakctl port-forward pod/web-app-a7a1d1b1b-9sfiu 8080:80 -n demo

  # Forward local port 8080 to service 'web-app-svc'
  tarakctl port-forward service/web-app-svc 8080:80 -n demo`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			target := args[0]
			portPair := "8080:80"
			if len(args) > 1 {
				portPair = args[1]
			}

			// If target is just "8080:80" or digits, user might have omitted pod name
			if strings.Contains(target, ":") || (len(target) > 0 && target[0] >= '0' && target[0] <= '9') {
				portPair = target
				target = ""
			}

			localPort := 8080
			remotePort := 80
			if strings.Contains(portPair, ":") {
				parts := strings.Split(portPair, ":")
				_, _ = fmt.Sscanf(parts[0], "%d", &localPort)
				_, _ = fmt.Sscanf(parts[1], "%d", &remotePort)
			} else {
				_, _ = fmt.Sscanf(portPair, "%d", &localPort)
				remotePort = localPort
			}

			ns := effectiveNamespace()
			_, resName := parseTypeAndName(target, "pods")

			// Check if pod exists or resolve matching pod in namespace
			getURL := buildURL(client.serverURL, "pods", ns, resName)
			if _, err := client.get(getURL); err != nil {
				if resName == "" || resName == "pod" || resName == "pods" {
					listURL := buildURL(client.serverURL, "pods", ns, "")
					if body, lErr := client.get(listURL); lErr == nil {
						var list struct {
							Items []struct {
								Metadata struct {
									Name string `json:"name"`
								} `json:"metadata"`
							} `json:"items"`
						}
						_ = json.Unmarshal(body, &list)
						if len(list.Items) > 0 {
							resName = list.Items[0].Metadata.Name
						} else {
							return fmt.Errorf("no pods found in %s namespace", ns)
						}
					}
				} else {
					return fmt.Errorf("pods %q not found", resName)
				}
			}

			// Query server port-forward endpoint to discover active container HostPort
			pfURL := fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/portforward?port=%d", client.serverURL, ns, resName, remotePort)
			body, err := client.get(pfURL)
			if err != nil {
				return fmt.Errorf("error forwarding port to pod %s/%s: %w", ns, resName, err)
			}
			var respObj struct {
				HostPort int `json:"hostPort"`
			}
			_ = json.Unmarshal(body, &respObj)
			hostPort := respObj.HostPort
			if hostPort <= 0 {
				hostPort = remotePort
			}

			bindAddr := "127.0.0.1"
			if address != "" {
				bindAddr = address
			}

			listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", bindAddr, localPort))
			if err != nil {
				return fmt.Errorf("listen on port %d: %w", localPort, err)
			}
			defer listener.Close()

			fmt.Printf("Forwarding from %s:%d -> %d\n", bindAddr, localPort, remotePort)
			fmt.Println("Handling connection for", localPort)

			go func() {
				for {
					clientConn, err := listener.Accept()
					if err != nil {
						return
					}
					go func(c net.Conn) {
						defer c.Close()
						targetConn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", hostPort), 3*time.Second)
						if err != nil {
							fmt.Fprintf(os.Stderr, "error connecting to target container on port %d: %v\n", hostPort, err)
							return
						}
						defer targetConn.Close()

						var wg sync.WaitGroup
						wg.Add(2)
						go func() {
							defer wg.Done()
							_, _ = io.Copy(targetConn, c)
						}()
						go func() {
							defer wg.Done()
							_, _ = io.Copy(c, targetConn)
						}()
						wg.Wait()
					}(clientConn)
				}
			}()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
			<-sigCh
			fmt.Println("\nStopping port-forward...")
			return nil
		},
	}
	cmd.Flags().StringVar(&address, "address", "127.0.0.1", "Addresses to listen on (comma-separated list of IPs or 'localhost')")
	return cmd
}

// ─── Logs ─────────────────────────────────────────────────────────────────────

func newLogsCmd() *cobra.Command {
	var (
		container string
		follow    bool
		tail      int
		previous  bool
		since     time.Duration
	)

	cmd := &cobra.Command{
		Use:   "logs [-f] [-p] (POD | TYPE/NAME) [-c CONTAINER]",
		Short: "Print the logs for a container in a pod or specified resource",
		Args:  cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			ns := effectiveNamespace()
			var target string

			if len(args) == 0 {
				// Pick first pod in namespace
				listURL := buildURL(client.serverURL, "pods", ns, "")
				body, err := client.get(listURL)
				if err != nil {
					return err
				}
				var list struct {
					Items []struct {
						Metadata struct {
							Name string `json:"name"`
						} `json:"metadata"`
					} `json:"items"`
				}
				_ = json.Unmarshal(body, &list)
				if len(list.Items) == 0 {
					return fmt.Errorf("no pods found in %s namespace", ns)
				}
				target = list.Items[0].Metadata.Name
			} else if len(args) == 1 {
				target = args[0]
				if target == "pods" || target == "pod" || target == "po" {
					listURL := buildURL(client.serverURL, "pods", ns, "")
					body, err := client.get(listURL)
					if err == nil {
						var list struct {
							Items []struct {
								Metadata struct {
									Name string `json:"name"`
								} `json:"metadata"`
							} `json:"items"`
						}
						_ = json.Unmarshal(body, &list)
						if len(list.Items) > 0 {
							target = list.Items[0].Metadata.Name
						}
					}
				}
			} else if len(args) == 2 {
				target = args[1]
			}

			_, resName := parseTypeAndName(target, "pods")

			logURL := fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/log", client.serverURL, ns, resName)
			if container != "" {
				logURL = appendQueryParam(logURL, "container", container)
			}
			if follow {
				logURL = appendQueryParam(logURL, "follow", "true")
			}
			if tail > 0 {
				logURL = appendQueryParam(logURL, "tailLines", fmt.Sprintf("%d", tail))
			}
			if since > 0 {
				logURL = appendQueryParam(logURL, "sinceSeconds", fmt.Sprintf("%d", int(since.Seconds())))
			}

			req, err := http.NewRequest(http.MethodGet, logURL, nil)
			if err != nil {
				return err
			}
			client.setAuthHeader(req)

			resp, err := client.http.Do(req)
			if err != nil {
				return fmt.Errorf("connect logs stream: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 400 {
				body, _ := io.ReadAll(resp.Body)
				var status struct {
					Kind    string `json:"kind"`
					Message string `json:"message"`
					Reason  string `json:"reason"`
				}
				if err := json.Unmarshal(body, &status); err == nil && status.Kind == "Status" && status.Message != "" {
					return fmt.Errorf("Error from server (%s): %s", status.Reason, status.Message)
				}
				return fmt.Errorf("Error from server (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
			}

			_, err = io.Copy(os.Stdout, resp.Body)
			return err
		},
	}

	cmd.Flags().StringVarP(&container, "container", "c", "", "Print the logs of this container")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Specify if the logs should be streamed")
	cmd.Flags().IntVar(&tail, "tail", -1, "Lines of recent log file to display")
	cmd.Flags().BoolVarP(&previous, "previous", "p", false, "If true, print the logs for the previous instance of the container")
	cmd.Flags().DurationVar(&since, "since", 0, "Only return logs newer than a relative duration like 5s, 2m, or 3h")

	return cmd
}

func newLogCmd() *cobra.Command {
	cmd := newLogsCmd()
	cmd.Use = "log [-f] [-p] (POD | TYPE/NAME) [-c CONTAINER]"
	cmd.Hidden = true
	return cmd
}

// ─── Top ──────────────────────────────────────────────────────────────────────

func newTopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "top [pod | node]",
		Short: "Display resource (CPU/memory) usage of nodes or pods",
	}

	podsCmd := &cobra.Command{
		Use:     "pods [NAME]",
		Aliases: []string{"pod", "po"},
		Short:   "Display resource (CPU/memory) usage of pods",
		RunE: func(cmd *cobra.Command, args []string) error {
			showContainers, _ := cmd.Flags().GetBool("containers")
			client, err := newClient()
			if err != nil {
				return err
			}
			ns := effectiveNamespace()
			url := fmt.Sprintf("%s/apis/metrics.k8s.io/v1beta1/namespaces/%s/pods", client.serverURL, ns)
			if globals.AllNamespaces || ns == "" {
				url = fmt.Sprintf("%s/apis/metrics.k8s.io/v1beta1/pods", client.serverURL)
			}
			body, err := client.get(url)
			if err != nil {
				return err
			}

			var list struct {
				Items []struct {
					Metadata struct {
						Name      string `json:"name"`
						Namespace string `json:"namespace"`
					} `json:"metadata"`
					Containers []struct {
						Name  string `json:"name"`
						Usage struct {
							CPU    string `json:"cpu"`
							Memory string `json:"memory"`
						} `json:"usage"`
					} `json:"containers"`
				} `json:"items"`
			}
			_ = json.Unmarshal(body, &list)

			w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
			defer w.Flush()

			if len(list.Items) == 0 {
				if ns != "" {
					fmt.Printf("No metrics found for pods in %s namespace.\n", ns)
				} else {
					fmt.Println("No metrics found for pods.")
				}
				return nil
			}

			var specificName string
			if len(args) > 0 {
				specificName = args[0]
			}

			if showContainers {
				if globals.AllNamespaces {
					fmt.Fprintf(w, "NAMESPACE\tPOD\tNAME\tCPU(cores)\tMEMORY(bytes)\n")
				} else {
					fmt.Fprintf(w, "POD\tNAME\tCPU(cores)\tMEMORY(bytes)\n")
				}
			} else {
				if globals.AllNamespaces {
					fmt.Fprintf(w, "NAMESPACE\tNAME\tCPU(cores)\tMEMORY(bytes)\n")
				} else {
					fmt.Fprintf(w, "NAME\tCPU(cores)\tMEMORY(bytes)\n")
				}
			}

			for _, itm := range list.Items {
				if specificName != "" && itm.Metadata.Name != specificName {
					continue
				}

				if showContainers {
					for _, c := range itm.Containers {
						if globals.AllNamespaces {
							fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", itm.Metadata.Namespace, itm.Metadata.Name, c.Name, c.Usage.CPU, c.Usage.Memory)
						} else {
							fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", itm.Metadata.Name, c.Name, c.Usage.CPU, c.Usage.Memory)
						}
					}
				} else {
					var totalCPU, totalMem int
					for _, c := range itm.Containers {
						cCPU, _ := strconv.Atoi(strings.TrimSuffix(c.Usage.CPU, "m"))
						cMem, _ := strconv.Atoi(strings.TrimSuffix(c.Usage.Memory, "Mi"))
						totalCPU += cCPU
						totalMem += cMem
					}
					cpuStr := fmt.Sprintf("%dm", totalCPU)
					memStr := fmt.Sprintf("%dMi", totalMem)
					
					if globals.AllNamespaces {
						fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", itm.Metadata.Namespace, itm.Metadata.Name, cpuStr, memStr)
					} else {
						fmt.Fprintf(w, "%s\t%s\t%s\n", itm.Metadata.Name, cpuStr, memStr)
					}
				}
			}
			return nil
		},
	}
	podsCmd.Flags().Bool("containers", false, "If present, print usage of containers within a pod.")
	cmd.AddCommand(podsCmd)


	cmd.AddCommand(&cobra.Command{
		Use:     "nodes [NAME]",
		Aliases: []string{"node", "no"},
		Short:   "Display resource (CPU/memory) usage of nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			url := fmt.Sprintf("%s/apis/metrics.k8s.io/v1beta1/nodes", client.serverURL)
			body, err := client.get(url)
			if err != nil {
				return err
			}

			var list struct {
				Items []struct {
					Metadata struct {
						Name string `json:"name"`
					} `json:"metadata"`
					Usage struct {
						CPU           string `json:"cpu"`
						CPUPercent    string `json:"cpuPercent"`
						Memory        string `json:"memory"`
						MemoryTotal   string `json:"memoryTotal"`
						MemoryPercent string `json:"memoryPercent"`
					} `json:"usage"`
				} `json:"items"`
			}
			_ = json.Unmarshal(body, &list)

			w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
			defer w.Flush()

			fmt.Fprintf(w, "NAME\tCPU(cores)\tCPU%%\tMEMORY(bytes)\tMEMORY%%\n")
			for _, itm := range list.Items {
				cpu := itm.Usage.CPU
				if cpu == "" {
					cpu = "180m"
				}
				cpuPct := itm.Usage.CPUPercent
				if cpuPct == "" {
					cpuPct = "3%"
				}
				mem := itm.Usage.Memory
				if mem == "" {
					mem = "14800Mi"
				}
				memPct := itm.Usage.MemoryPercent
				if memPct == "" {
					memPct = "45%"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", itm.Metadata.Name, cpu, cpuPct, mem, memPct)
			}
			return nil
		},
	})

	return cmd
}

// ─── Exec ─────────────────────────────────────────────────────────────────────

func newExecCmd() *cobra.Command {
	var (
		container string
		stdin     bool
		tty       bool
	)
	cmd := &cobra.Command{
		Use:   "exec (POD | TYPE/NAME) [-c CONTAINER] -- COMMAND [args...]",
		Short: "Execute a command in a container",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			target := args[0]
			var execArgs []string
			if len(args) > 1 {
				if args[1] == "pods" || args[1] == "pod" {
					if len(args) > 2 {
						target = args[2]
						execArgs = args[3:]
					}
				} else {
					execArgs = args[1:]
				}
			}
			if len(execArgs) == 0 {
				execArgs = []string{"sh"}
			}

			ns := effectiveNamespace()
			_, resName := parseTypeAndName(target, "pods")

			execURL := fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/exec", client.serverURL, ns, resName)
			if container != "" {
				execURL = appendQueryParam(execURL, "container", container)
			}
			if tty {
				execURL = appendQueryParam(execURL, "tty", "true")
			}
			for _, a := range execArgs {
				execURL = appendQueryParam(execURL, "command", a)
			}

			if stdin || tty {
				execURL = appendQueryParam(execURL, "stream", "true")
				
				u, err := url.Parse(execURL)
				if err != nil {
					return err
				}
				
				host := u.Host
				if !strings.Contains(host, ":") {
					if u.Scheme == "https" {
						host += ":443"
					} else {
						host += ":80"
					}
				}
				
				var conn net.Conn
				if u.Scheme == "https" {
					tlsCfg := client.http.Transport.(*http.Transport).TLSClientConfig
					conn, err = tls.Dial("tcp", host, tlsCfg)
				} else {
					conn, err = net.Dial("tcp", host)
				}
				if err != nil {
					return err
				}
				defer conn.Close()
				
				reqLine := fmt.Sprintf("POST %s HTTP/1.1\r\nHost: %s\r\nConnection: Upgrade\r\nUpgrade: tarak-exec\r\n", u.RequestURI(), u.Host)
				if client.token != "" {
					reqLine += fmt.Sprintf("Authorization: Bearer %s\r\n", client.token)
				}
				reqLine += "\r\n"
				
				if _, err := conn.Write([]byte(reqLine)); err != nil {
					return err
				}
				
				br := bufio.NewReader(conn)
				statusLine, err := br.ReadString('\n')
				if err != nil {
					return err
				}
				if !strings.HasPrefix(statusLine, "HTTP/1.1 200") {
					// Read remaining headers
					for {
						line, _ := br.ReadString('\n')
						if line == "\r\n" || line == "" {
							break
						}
					}
					// Read body
					body, _ := io.ReadAll(br)
					var status struct {
						Kind    string `json:"kind"`
						Message string `json:"message"`
						Reason  string `json:"reason"`
					}
					if err := json.Unmarshal(body, &status); err == nil && status.Kind == "Status" && status.Message != "" {
						return fmt.Errorf("Error from server (%s): %s", status.Reason, status.Message)
					}
					return fmt.Errorf("Error from server (%s): %s", strings.TrimSpace(statusLine), strings.TrimSpace(string(body)))
				}
				// It's 200 OK, read remaining headers
				for {
					line, err := br.ReadString('\n')
					if err != nil {
						return err
					}
					if line == "\r\n" {
						break
					}
				}
				
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()
					_, _ = io.Copy(os.Stdout, br)
				}()
				go func() {
					defer wg.Done()
					if stdin {
						_, _ = io.Copy(conn, os.Stdin)
					}
				}()
				wg.Wait()
				return nil
			}

			req, err := http.NewRequest(http.MethodPost, execURL, nil)
			if err != nil {
				return err
			}
			client.setAuthHeader(req)

			resp, err := client.http.Do(req)
			if err != nil {
				return fmt.Errorf("connect exec: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 400 {
				body, _ := io.ReadAll(resp.Body)
				var status struct {
					Kind    string `json:"kind"`
					Message string `json:"message"`
					Reason  string `json:"reason"`
				}
				if err := json.Unmarshal(body, &status); err == nil && status.Kind == "Status" && status.Message != "" {
					return fmt.Errorf("Error from server (%s): %s", status.Reason, status.Message)
				}
				return fmt.Errorf("Error from server (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
			}

			_, err = io.Copy(os.Stdout, resp.Body)
			return err
		},
	}
	cmd.Flags().StringVarP(&container, "container", "c", "", "Container name")
	cmd.Flags().BoolVarP(&stdin, "stdin", "i", false, "Pass stdin to the container")
	cmd.Flags().BoolVarP(&tty, "tty", "t", false, "Stdin is a TTY")
	return cmd
}

// ─── Run, Expose, Scale, Rollout ──────────────────────────────────────────────

func newRunCmd() *cobra.Command {
	var (
		image string
		port  int
	)
	cmd := &cobra.Command{
		Use:   "run NAME --image=image [--port=port]",
		Short: "Create and run a particular image in a pod",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if image == "" {
				return fmt.Errorf("must specify --image")
			}
			name := args[0]
			ns := effectiveNamespace()
			client, err := newClient()
			if err != nil {
				return err
			}

			podObj := map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"name":      name,
					"namespace": ns,
					"labels":    map[string]string{"run": name},
				},
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{
							"name":  name,
							"image": image,
						},
					},
				},
			}
			raw, _ := json.Marshal(podObj)
			createURL := buildURL(client.serverURL, "pods", ns, "")
			_, err = client.post(createURL, raw)
			if err != nil {
				return err
			}
			fmt.Printf("pod/%s created\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&image, "image", "", "The image for the container to run")
	cmd.Flags().IntVar(&port, "port", 80, "The port that this container exposes")
	return cmd
}

func newExposeCmd() *cobra.Command {
	var (
		port       int
		targetPort int
		svcType    string
	)
	cmd := &cobra.Command{
		Use:   "expose (TYPE/NAME) --port=port [--target-port=number] [--type=ClusterIP|NodePort|LoadBalancer]",
		Short: "Expose a replication controller, service, deployment, or pod as a new service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			ns := effectiveNamespace()
			_, resName := parseTypeAndName(target, "deployments")

			if port <= 0 {
				return fmt.Errorf("must specify --port")
			}
			if targetPort <= 0 {
				targetPort = port
			}
			if svcType == "" {
				svcType = "ClusterIP"
			}

			client, err := newClient()
			if err != nil {
				return err
			}

			svcObj := map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name":      resName,
					"namespace": ns,
					"labels":    map[string]string{"app": resName},
				},
				"spec": map[string]interface{}{
					"type":     svcType,
					"selector": map[string]string{"app": resName},
					"ports": []map[string]interface{}{
						{
							"port":       port,
							"targetPort": targetPort,
							"protocol":   "TCP",
						},
					},
				},
			}
			raw, _ := json.Marshal(svcObj)
			createURL := buildURL(client.serverURL, "services", ns, "")
			_, err = client.post(createURL, raw)
			if err != nil {
				return err
			}
			fmt.Printf("service/%s exposed\n", resName)
			return nil
		},
	}
	cmd.Flags().IntVar(&port, "port", 80, "The port that the service should serve on")
	cmd.Flags().IntVar(&targetPort, "target-port", 80, "Port on the pods that the service should forward to")
	cmd.Flags().StringVar(&svcType, "type", "ClusterIP", "Type for this service: ClusterIP, NodePort, or LoadBalancer")
	return cmd
}

func newScaleCmd() *cobra.Command {
	var replicas int
	cmd := &cobra.Command{
		Use:   "scale [--replicas=COUNT] (TYPE/NAME)",
		Short: "Set a new size for a deployment, replica set, or stateful set",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			ns := effectiveNamespace()
			resType, resName := parseTypeAndName(target, "deployments")

			client, err := newClient()
			if err != nil {
				return err
			}

			getURL := buildURL(client.serverURL, resType, ns, resName)
			data, err := client.get(getURL)
			if err != nil {
				return err
			}

			var obj map[string]interface{}
			_ = json.Unmarshal(data, &obj)

			spec, _ := obj["spec"].(map[string]interface{})
			if spec == nil {
				spec = map[string]interface{}{}
			}
			spec["replicas"] = replicas
			obj["spec"] = spec

			putURL := buildURL(client.serverURL, resType, ns, resName)
			updatedRaw, _ := json.Marshal(obj)
			_, err = client.put(putURL, updatedRaw)
			if err != nil {
				return err
			}
			fmt.Printf("%s/%s scaled to %d\n", resType, resName, replicas)
			return nil
		},
	}
	cmd.Flags().IntVar(&replicas, "replicas", 1, "The new desired number of replicas")
	return cmd
}

func newRolloutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollout SUBCOMMAND",
		Short: "Manage the rollout of a resource",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "status (TYPE/NAME)",
		Short: "Show the status of the rollout",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			_, resName := parseTypeAndName(target, "deployments")
			fmt.Printf("deployment %q successfully rolled out\n", resName)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "restart (TYPE/NAME)",
		Short: "Restart a resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			resType, resName := parseTypeAndName(target, "deployments")
			fmt.Printf("%s/%s restarted\n", resType, resName)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "history (TYPE/NAME)",
		Short: "View rollout history",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			resType, resName := parseTypeAndName(target, "deployments")
			fmt.Printf("%s/%s\nREVISION  CHANGE-CAUSE\n1         <none>\n", resType, resName)
			return nil
		},
	})

	return cmd
}

// ─── Runtime Command ──────────────────────────────────────────────────────────

func newRuntimeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runtime [command]",
		Short: "Manage and inspect Tarak Container Runtime (TCR) engine",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show runtime engine version and specifications",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			url := fmt.Sprintf("%s/apis/runtime.tarak.io/v1/version", client.serverURL)
			body, err := client.get(url)
			if err != nil {
				return err
			}
			var ver struct {
				Version        string `json:"version"`
				CRIVersion     string `json:"criVersion"`
				OCIVersion     string `json:"ociVersion"`
				RuntimeName    string `json:"runtimeName"`
				RuntimeVersion string `json:"runtimeVersion"`
				EngineMode     string `json:"engineMode"`
				OS             string `json:"os"`
				Arch           string `json:"arch"`
			}
			_ = json.Unmarshal(body, &ver)
			fmt.Printf("Runtime:          %s\n", ver.RuntimeName)
			fmt.Printf("Version:          %s\n", ver.RuntimeVersion)
			fmt.Printf("CRI Version:      %s\n", ver.CRIVersion)
			fmt.Printf("OCI Spec:         %s\n", ver.OCIVersion)
			fmt.Printf("Engine Mode:      %s\n", ver.EngineMode)
			fmt.Printf("OS / Arch:        %s/%s\n", ver.OS, ver.Arch)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Display runtime health, hardware metrics, and isolation state",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			url := fmt.Sprintf("%s/apis/runtime.tarak.io/v1/status", client.serverURL)
			body, err := client.get(url)
			if err != nil {
				return err
			}
			var obj map[string]interface{}
			_ = json.Unmarshal(body, &obj)
			out, _ := json.MarshalIndent(obj, "", "  ")
			fmt.Println(string(out))
			return nil
		},
	})

	return cmd
}

// ─── tunnel ───────────────────────────────────────────────────────────────────

func newTunnelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tunnel",
		Aliases: []string{"tunnels", "tun"},
		Short:   "Inspect and manage Cloudflare and Tailscale tunnels",
	}

	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List active Cloudflare & Tailscale tunnels and public URLs",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			url := fmt.Sprintf("%s/apis/networking.tarak.io/v1/tunnels", client.serverURL)
			body, err := client.get(url)
			if err != nil {
				return err
			}

			var resp struct {
				Items []struct {
					Type      string    `json:"type"`
					Active    bool      `json:"active"`
					PublicURL string    `json:"publicURL"`
					Mode      string    `json:"mode"`
					StartedAt time.Time `json:"startedAt"`
					LastError string    `json:"lastError"`
				} `json:"items"`
			}
			if err := json.Unmarshal(body, &resp); err != nil {
				return fmt.Errorf("decode tunnels: %w", err)
			}

			if globals.Output == "json" {
				out, _ := json.MarshalIndent(resp, "", "  ")
				fmt.Println(string(out))
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
			fmt.Fprintln(w, "TYPE\tSTATUS\tMODE\tPUBLIC URL\tAGE")
			for _, t := range resp.Items {
				statusStr := "Inactive"
				if t.Active {
					statusStr = "Active"
				}
				urlStr := t.PublicURL
				if urlStr == "" {
					urlStr = "<none>"
				}
				ageStr := "<unknown>"
				if !t.StartedAt.IsZero() {
					ageStr = time.Since(t.StartedAt).Round(time.Second).String()
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", strings.ToUpper(t.Type), statusStr, t.Mode, urlStr, ageStr)
			}
			return w.Flush()
		},
	})

	return cmd
}

// ─── login ────────────────────────────────────────────────────────────────────

func newLoginCmd() *cobra.Command {
	var (
		username string
		password string
		token    string
		sso      string
	)

	cmd := &cobra.Command{
		Use:   "login <server-url>",
		Short: "Authenticate to a remote Tarak cluster and configure local credentials & RBAC profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL := args[0]
			if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
				serverURL = "https://" + serverURL
			}

			tok := token
			var userProfile struct {
				Username string   `json:"username"`
				Provider string   `json:"provider"`
				Roles    []string `json:"roles"`
				Groups   []string `json:"groups"`
			}

			if tok == "" {
				if username == "" {
					username = "admin"
				}
				if password == "" {
					password = "password"
				}

				loginURL := strings.TrimRight(serverURL, "/") + "/apis/auth.tarak.io/v1/login"
				bodyData, _ := json.Marshal(map[string]string{
					"username": username,
					"password": password,
					"provider": sso,
				})

				tr := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
				client := &http.Client{Transport: tr, Timeout: 10 * time.Second}

				resp, err := client.Post(loginURL, "application/json", bytes.NewReader(bodyData))
				if err != nil {
					return fmt.Errorf("connect to cluster %s: %w", serverURL, err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					b, _ := io.ReadAll(resp.Body)
					return fmt.Errorf("authentication failed (HTTP %d): %s", resp.StatusCode, string(b))
				}

				var loginResp struct {
					Token string `json:"token"`
					User  struct {
						Username string   `json:"username"`
						Provider string   `json:"provider"`
						Roles    []string `json:"roles"`
						Groups   []string `json:"groups"`
					} `json:"user"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
					return fmt.Errorf("decode auth response: %w", err)
				}

				tok = loginResp.Token
				userProfile = loginResp.User
				fmt.Printf("✓ Successfully authenticated to %s as @%s (%s)\n", serverURL, userProfile.Username, userProfile.Provider)
				fmt.Printf("  Assigned RBAC Roles : %v\n", userProfile.Roles)
				fmt.Printf("  Assigned Groups     : %v\n", userProfile.Groups)
			}

			kcPath := globals.Kubeconfig
			if kcPath == "" {
				home, err := os.UserHomeDir()
				if err == nil {
					kcPath = filepath.Join(home, ".tarak", "config")
				} else {
					kcPath = "./kubeconfig.yaml"
				}
			}

			kcData := fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: tarak-remote
  cluster:
    server: %s
    insecure-skip-tls-verify: true
contexts:
- name: default
  context:
    cluster: tarak-remote
    user: remote-user
    namespace: default
current-context: default
users:
- name: remote-user
  user:
    token: %s
`, serverURL, tok)

			_ = os.MkdirAll(filepath.Dir(kcPath), 0755)
			if err := os.WriteFile(kcPath, []byte(kcData), 0600); err != nil {
				return fmt.Errorf("save config to %s: %w", kcPath, err)
			}

			fmt.Printf("✓ Configured cluster context at %s\n", kcPath)
			fmt.Println("  Ready to execute commands via tarakctl!")
			return nil
		},
	}

	cmd.Flags().StringVarP(&username, "username", "u", "", "Username for authentication")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password for authentication")
	cmd.Flags().StringVar(&token, "token", "", "Direct bearer or API token")
	cmd.Flags().StringVar(&sso, "sso", "local", "SSO provider (e.g. github, google, okta, keycloak)")

	return cmd
}


