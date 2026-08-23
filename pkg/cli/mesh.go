package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// newMeshCmd creates the 'mesh' command family for managing native Tarak Multi-Mesh.
func newMeshCmd() *cobra.Command {
	meshCmd := &cobra.Command{
		Use:     "mesh",
		Aliases: []string{"meshes"},
		Short:   "Manage native Tarak multi-mesh, zero-trust mTLS, and virtual DNS",
		Long: `======================================================================
  TARAK NATIVE SERVICE MESH (Kuma / Kong-Mesh Universal Equivalent)
======================================================================

The 'mesh' command suite allows managing isolated multi-tenant meshes,
inspecting discovered workloads, zero-trust traffic permissions,
egress passthrough policies, external services, and canary traffic routes.`,
	}

	meshCmd.AddCommand(newMeshListCmd())
	meshCmd.AddCommand(newMeshCreateCmd())
	meshCmd.AddCommand(newMeshServicesCmd())
	meshCmd.AddCommand(newMeshExternalCmd())
	meshCmd.AddCommand(newMeshPermissionsCmd())
	meshCmd.AddCommand(newMeshPassthroughCmd())
	meshCmd.AddCommand(newMeshPatchesCmd())
	meshCmd.AddCommand(newMeshRoutesCmd())

	return meshCmd
}

// ─── mesh list ───────────────────────────────────────────────────────────────

func newMeshListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "get"},
		Short:   "List all active Service Meshes in the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newClient()
			if err != nil {
				return err
			}

			reqURL := fmt.Sprintf("%s/apis/mesh.tarak.io/v1/meshes", c.serverURL)
			data, err := c.get(reqURL)
			if err != nil {
				return fmt.Errorf("connect to mesh API: %w", err)
			}

			var listResp struct {
				Items []struct {
					Metadata struct {
						Name string `json:"name"`
					} `json:"metadata"`
					Spec struct {
						MTLS struct {
							Enabled bool   `json:"enabled"`
							Mode    string `json:"mode"`
						} `json:"mtls"`
						Networking struct {
							PassthroughMode string `json:"passthroughMode"`
						} `json:"networking"`
						Routing struct {
							DefaultTimeout string `json:"defaultTimeout"`
							RetryAttempts  int    `json:"retryAttempts"`
						} `json:"routing"`
					} `json:"spec"`
					Status struct {
						Phase            string `json:"phase"`
						TotalServices    int    `json:"totalServices"`
						EnrolledPods     int    `json:"enrolledPods"`
						ConnectedProxies int    `json:"connectedProxies"`
					} `json:"status"`
				} `json:"items"`
			}

			if err := json.Unmarshal(data, &listResp); err != nil {
				return fmt.Errorf("decode response: %w", err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tPHASE\tMTLS MODE\tPASSTHROUGH\tSERVICES\tENROLLED PODS\tPROXIES")
			for _, m := range listResp.Items {
				mtlsMode := m.Spec.MTLS.Mode
				if mtlsMode == "" {
					mtlsMode = "Permissive"
				}
				passMode := m.Spec.Networking.PassthroughMode
				if passMode == "" {
					passMode = "Passthrough"
				}
				phase := m.Status.Phase
				if phase == "" {
					phase = "Active"
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\t%d\n",
					m.Metadata.Name,
					phase,
					mtlsMode,
					passMode,
					m.Status.TotalServices,
					m.Status.EnrolledPods,
					m.Status.ConnectedProxies,
				)
			}
			return w.Flush()
		},
	}
}

// ─── mesh create ─────────────────────────────────────────────────────────────

func newMeshCreateCmd() *cobra.Command {
	var (
		mtlsMode    string
		passthrough string
	)

	cmd := &cobra.Command{
		Use:   "create <mesh-name>",
		Short: "Create a new isolated multi-tenant Service Mesh",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			c, err := newClient()
			if err != nil {
				return err
			}

			payload := map[string]interface{}{
				"apiVersion": "mesh.tarak.io/v1",
				"kind":       "Mesh",
				"metadata": map[string]interface{}{
					"name": name,
				},
				"spec": map[string]interface{}{
					"mtls": map[string]interface{}{
						"enabled": true,
						"mode":    mtlsMode,
					},
					"networking": map[string]interface{}{
						"passthroughMode": passthrough,
					},
				},
			}

			body, _ := json.Marshal(payload)
			reqURL := fmt.Sprintf("%s/apis/mesh.tarak.io/v1/meshes", c.serverURL)
			_, err = c.post(reqURL, body)
			if err != nil {
				return fmt.Errorf("create mesh failed: %w", err)
			}

			fmt.Printf("✓ Service Mesh '%s' created successfully (mTLS: %s, Passthrough: %s)\n", name, mtlsMode, passthrough)
			fmt.Printf("  Virtual DNS Suffix: .%s.mesh\n", name)
			fmt.Printf("  Workloads can be auto-enrolled with label: mesh.tarak.io/name: %s\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&mtlsMode, "mtls", "Strict", "mTLS Zero-Trust mode: Strict | Permissive")
	cmd.Flags().StringVar(&passthrough, "passthrough", "Passthrough", "Egress passthrough mode: Passthrough | DenyAll")
	return cmd
}

// ─── mesh services ───────────────────────────────────────────────────────────

func newMeshServicesCmd() *cobra.Command {
	var meshName string

	cmd := &cobra.Command{
		Use:   "services [mesh-name]",
		Short: "List all auto-discovered mesh services and .mesh DNS records",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				meshName = args[0]
			}
			if meshName == "" {
				meshName = "default"
			}

			c, err := newClient()
			if err != nil {
				return err
			}

			reqURL := fmt.Sprintf("%s/apis/mesh.tarak.io/v1/meshes/%s/services", c.serverURL, meshName)
			data, err := c.get(reqURL)
			if err != nil {
				return fmt.Errorf("connect to mesh services API: %w", err)
			}

			var listResp struct {
				Items []struct {
					Metadata struct {
						Name      string `json:"name"`
						Namespace string `json:"namespace"`
					} `json:"metadata"`
					Spec struct {
						VirtualHost string `json:"virtualHost"`
						VIP         string `json:"vip"`
						SPIFFEID    string `json:"spiffeId"`
						Ports       []struct {
							Name     string `json:"name"`
							Port     int    `json:"port"`
							Protocol string `json:"protocol"`
						} `json:"ports"`
					} `json:"spec"`
					Status struct {
						EndpointsCount int    `json:"endpointsCount"`
						HealthState    string `json:"healthState"`
					} `json:"status"`
				} `json:"items"`
			}

			if err := json.Unmarshal(data, &listResp); err != nil {
				return fmt.Errorf("decode services response: %w", err)
			}

			if len(listResp.Items) == 0 {
				fmt.Printf("No services enrolled in mesh '%s'.\n", meshName)
				fmt.Println("To enroll a service, annotate its Pod or Namespace with:")
				fmt.Printf("  tarak.io/mesh: \"%s\"\n", meshName)
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
			fmt.Fprintln(w, "SERVICE\tNAMESPACE\tVIRTUAL HOST (.mesh)\tVIP\tPORT\tSPIFFE ID\tHEALTH")
			for _, s := range listResp.Items {
				portStr := "80/TCP"
				if len(s.Spec.Ports) > 0 {
					portStr = fmt.Sprintf("%d/%s", s.Spec.Ports[0].Port, s.Spec.Ports[0].Protocol)
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					s.Metadata.Name,
					s.Metadata.Namespace,
					s.Spec.VirtualHost,
					s.Spec.VIP,
					portStr,
					s.Spec.SPIFFEID,
					s.Status.HealthState,
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVarP(&meshName, "mesh", "m", "default", "Target Service Mesh name")
	return cmd
}

// ─── mesh external ───────────────────────────────────────────────────────────

func newMeshExternalCmd() *cobra.Command {
	var meshName string

	cmd := &cobra.Command{
		Use:   "external [mesh-name]",
		Short: "List external services and TLS origination endpoints for a mesh",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				meshName = args[0]
			}
			if meshName == "" {
				meshName = "default"
			}

			c, err := newClient()
			if err != nil {
				return err
			}

			reqURL := fmt.Sprintf("%s/apis/mesh.tarak.io/v1/meshes/%s/external-services", c.serverURL, meshName)
			data, err := c.get(reqURL)
			if err != nil {
				return err
			}

			var listResp struct {
				Items []struct {
					Metadata struct {
						Name string `json:"name"`
					} `json:"metadata"`
					Spec struct {
						Host           string `json:"host"`
						Port           int    `json:"port"`
						Protocol       string `json:"protocol"`
						TLSOrigination struct {
							Enabled    bool   `json:"enabled"`
							ServerName string `json:"serverName"`
						} `json:"tlsOrigination"`
					} `json:"spec"`
				} `json:"items"`
			}

			if err := json.Unmarshal(data, &listResp); err != nil {
				return err
			}

			if len(listResp.Items) == 0 {
				fmt.Printf("No external services defined in mesh '%s'.\n", meshName)
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tHOST\tPORT\tPROTOCOL\tTLS ORIGINATION\tSERVER NAME")
			for _, ext := range listResp.Items {
				tlsStatus := "Disabled"
				if ext.Spec.TLSOrigination.Enabled {
					tlsStatus = "Enabled"
				}
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
					ext.Metadata.Name,
					ext.Spec.Host,
					ext.Spec.Port,
					ext.Spec.Protocol,
					tlsStatus,
					ext.Spec.TLSOrigination.ServerName,
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVarP(&meshName, "mesh", "m", "default", "Target Service Mesh name")
	return cmd
}

// ─── mesh permissions ────────────────────────────────────────────────────────

func newMeshPermissionsCmd() *cobra.Command {
	var meshName string

	cmd := &cobra.Command{
		Use:     "permissions [mesh-name]",
		Aliases: []string{"perms"},
		Short:   "List Zero-Trust mTLS Traffic Permissions in a mesh",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				meshName = args[0]
			}
			if meshName == "" {
				meshName = "default"
			}

			c, err := newClient()
			if err != nil {
				return err
			}

			reqURL := fmt.Sprintf("%s/apis/mesh.tarak.io/v1/meshes/%s/traffic-permissions", c.serverURL, meshName)
			data, err := c.get(reqURL)
			if err != nil {
				return err
			}

			var listResp struct {
				Items []struct {
					Metadata struct {
						Name string `json:"name"`
					} `json:"metadata"`
					Spec struct {
						Sources []struct {
							Match map[string]string `json:"match"`
						} `json:"sources"`
						Destinations []struct {
							Match map[string]string `json:"match"`
						} `json:"destinations"`
						Action string `json:"action"`
					} `json:"spec"`
				} `json:"items"`
			}

			if err := json.Unmarshal(data, &listResp); err != nil {
				return err
			}

			if len(listResp.Items) == 0 {
				fmt.Printf("No explicit traffic permissions found in mesh '%s' (Default: mTLS Strict Zero-Trust Deny-All).\n", meshName)
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
			fmt.Fprintln(w, "PERMISSION\tSOURCE SERVICE\tDESTINATION SERVICE\tACTION")
			for _, p := range listResp.Items {
				src := "*"
				if len(p.Spec.Sources) > 0 && p.Spec.Sources[0].Match["service"] != "" {
					src = p.Spec.Sources[0].Match["service"]
				}
				dst := "*"
				if len(p.Spec.Destinations) > 0 && p.Spec.Destinations[0].Match["service"] != "" {
					dst = p.Spec.Destinations[0].Match["service"]
				}
				action := p.Spec.Action
				if action == "" {
					action = "Allow"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Metadata.Name, src, dst, action)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVarP(&meshName, "mesh", "m", "default", "Target Service Mesh name")
	return cmd
}

// ─── mesh passthrough ────────────────────────────────────────────────────────

func newMeshPassthroughCmd() *cobra.Command {
	var meshName string

	cmd := &cobra.Command{
		Use:   "passthrough [mesh-name]",
		Short: "List Egress Passthrough policies in a mesh",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				meshName = args[0]
			}
			if meshName == "" {
				meshName = "default"
			}

			c, err := newClient()
			if err != nil {
				return err
			}

			reqURL := fmt.Sprintf("%s/apis/mesh.tarak.io/v1/meshes/%s/passthrough-policies", c.serverURL, meshName)
			data, err := c.get(reqURL)
			if err != nil {
				return err
			}

			var listResp struct {
				Items []struct {
					Metadata struct {
						Name string `json:"name"`
					} `json:"metadata"`
					Spec struct {
						Mode     string   `json:"mode"`
						CIDRs    []string `json:"cidrs"`
						Domains  []string `json:"domains"`
						LogStats bool     `json:"logStats"`
					} `json:"spec"`
				} `json:"items"`
			}

			if err := json.Unmarshal(data, &listResp); err != nil {
				return err
			}

			if len(listResp.Items) == 0 {
				fmt.Printf("No passthrough policies configured in mesh '%s'.\n", meshName)
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
			fmt.Fprintln(w, "POLICY\tMODE\tDOMAINS\tCIDRS\tLOGGING")
			for _, p := range listResp.Items {
				fmt.Fprintf(w, "%s\t%s\t%v\t%v\t%v\n",
					p.Metadata.Name,
					p.Spec.Mode,
					p.Spec.Domains,
					p.Spec.CIDRs,
					p.Spec.LogStats,
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVarP(&meshName, "mesh", "m", "default", "Target Service Mesh name")
	return cmd
}

// ─── mesh patches ────────────────────────────────────────────────────────────

func newMeshPatchesCmd() *cobra.Command {
	var meshName string

	cmd := &cobra.Command{
		Use:   "patches [mesh-name]",
		Short: "List proxy performance and concurrency patches in a mesh",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				meshName = args[0]
			}
			if meshName == "" {
				meshName = "default"
			}

			c, err := newClient()
			if err != nil {
				return err
			}

			reqURL := fmt.Sprintf("%s/apis/mesh.tarak.io/v1/meshes/%s/proxy-patches", c.serverURL, meshName)
			data, err := c.get(reqURL)
			if err != nil {
				return err
			}

			var listResp struct {
				Items []struct {
					Metadata struct {
						Name string `json:"name"`
					} `json:"metadata"`
					Spec struct {
						TargetService  string `json:"targetService"`
						Concurrency    int    `json:"concurrency"`
						BufferBytes    int    `json:"bufferBytes"`
						RequestTimeout string `json:"requestTimeout"`
					} `json:"spec"`
				} `json:"items"`
			}

			if err := json.Unmarshal(data, &listResp); err != nil {
				return err
			}

			if len(listResp.Items) == 0 {
				fmt.Printf("No custom proxy patches in mesh '%s'.\n", meshName)
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
			fmt.Fprintln(w, "PATCH\tTARGET SERVICE\tCONCURRENCY\tBUFFER (KB)\tTIMEOUT")
			for _, patch := range listResp.Items {
				fmt.Fprintf(w, "%s\t%s\t%d\t%d KB\t%s\n",
					patch.Metadata.Name,
					patch.Spec.TargetService,
					patch.Spec.Concurrency,
					patch.Spec.BufferBytes/1024,
					patch.Spec.RequestTimeout,
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVarP(&meshName, "mesh", "m", "default", "Target Service Mesh name")
	return cmd
}

// ─── mesh routes ────────────────────────────────────────────────────────────

func newMeshRoutesCmd() *cobra.Command {
	var meshName string

	cmd := &cobra.Command{
		Use:     "routes [mesh-name]",
		Aliases: []string{"route"},
		Short:   "List canary traffic routes and weighted split policies",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				meshName = args[0]
			}
			if meshName == "" {
				meshName = "default"
			}

			c, err := newClient()
			if err != nil {
				return err
			}

			reqURL := fmt.Sprintf("%s/apis/mesh.tarak.io/v1/meshes/%s/routes", c.serverURL, meshName)
			data, err := c.get(reqURL)
			if err != nil {
				return err
			}

			var listResp struct {
				Items []struct {
					Name           string                 `json:"name"`
					Host           string                 `json:"host"`
					Weights        map[string]int         `json:"weights"`
					Timeout        string                 `json:"timeout"`
					CircuitBreaker map[string]interface{} `json:"circuitBreaker"`
				} `json:"items"`
			}

			if err := json.Unmarshal(data, &listResp); err != nil {
				return err
			}

			if len(listResp.Items) == 0 {
				fmt.Printf("No canary traffic routes defined in mesh '%s'.\n", meshName)
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
			fmt.Fprintln(w, "ROUTE NAME\tHOST\tCANARY WEIGHTS\tTIMEOUT")
			for _, r := range listResp.Items {
				weightsStr := fmt.Sprintf("%v", r.Weights)
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Name, r.Host, weightsStr, r.Timeout)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVarP(&meshName, "mesh", "m", "default", "Target Service Mesh name")
	return cmd
}
