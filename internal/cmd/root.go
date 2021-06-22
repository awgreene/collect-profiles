package cmd

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"

	"github.com/awgreene/collect-profiles/internal/pkg/action"
	"github.com/awgreene/collect-profiles/internal/pkg/log"
	"github.com/awgreene/collect-profiles/internal/version"
)

const (
	profileConfigMapLabelKey = "olm.openshift.io/pprof"
)

var (
	rootCmd = newCmd()

	// Used for flags
	namespace   string
	configPath  string
	tlsCertPath string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "The Kubernetes namespace")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config-path", "c", "", "The Kubernetes namespace")
	rootCmd.MarkFlagRequired("config-path")
	rootCmd.PersistentFlags().StringVarP(&tlsCertPath, "tls-cert-path", "", "", "The client certificate provided when making requests against the pprof endpoint(s)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func newCmd() *cobra.Command {
	var cfg action.Configuration
	return &cobra.Command{
		Use:     "collect-profiles endpoint:argument",
		Short:   "Retrieve the pprof data from an endpoint",
		Long:    `Retrieve the pprof data from an endpoint`,
		Version: version.Version,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return cfg.Load()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("must specify endpoint")
			}

			// get configmap
			file, err := os.Open(configPath)
			if err != nil {
				return err
			}
			defer file.Close()

			configMap := &corev1.ConfigMap{}
			if err := yaml.NewYAMLOrJSONDecoder(file, 50).Decode(configMap); err != nil {
				return err
			}

			klog.Infof("Config = %v", configMap)
			if configMap.Data["suspend"] == "true" {
				klog.Infof("CronJob disabled, exiting")
				return nil
			}

			cert, err := tls.LoadX509KeyPair(tlsCertPath+corev1.TLSCertKey, tlsCertPath+corev1.TLSPrivateKeyKey)
			if err != nil {
				return err
			}

			validatedArguments := []*argument{}
			for _, arg := range args {
				a, err := newArgument(arg)
				if err != nil {
					return err
				}
				validatedArguments = append(validatedArguments, a)
			}

			httpClient := http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
						Certificates:       []tls.Certificate{cert},
					},
				},
			}

			for _, a := range validatedArguments {
				response, err := httpClient.Do(&http.Request{
					Method: http.MethodGet,
					URL:    a.endpoint,
				})
				if err != nil {
					klog.Errorf("error reading from pprof endpoint %s: %v", a.endpoint.String(), err)
					continue
				}

				var b bytes.Buffer
				if _, err := io.Copy(&b, response.Body); err != nil {
					return fmt.Errorf("error reading response body: %v", err)
				}
				trueBool := true
				configMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: a.generateName,
						Namespace:    namespace,
						Labels: map[string]string{
							profileConfigMapLabelKey: "",
						},
					},
					Immutable: &trueBool,
					BinaryData: map[string][]byte{
						"profile.pb.gz": b.Bytes(),
					},
				}
				// Group by generate name (map where key == genNAme value == list of strings, remove newest name, delete list)
				// exit if failure

				// TODO: Add Retry?
				if err := cfg.Client.Create(context.TODO(), configMap); err != nil {
					klog.Infof("error creating ConfigMap: %v", err)
				}

				// Group by generate name (map where key == genName value == list of strings, remove newest name, delete list)
				// Log errors, attempt to delete each / dont exit early
			}
			return nil
		},
	}
}

type argument struct {
	generateName string
	endpoint     *url.URL
}

func newArgument(s string) (*argument, error) {
	splitStrings := strings.SplitN(s, ":", 2)
	if len(splitStrings) != 2 {
		return nil, fmt.Errorf("Error")
	}

	endpoint, err := url.Parse(splitStrings[1])
	if err != nil {
		return nil, err
	}

	if strings.ToLower(endpoint.Scheme) != "https" {
		return nil, fmt.Errorf("Endpoint.Scheme must be HTTPS")
	}

	arg := &argument{
		generateName: splitStrings[0],
		endpoint:     endpoint,
	}

	return arg, nil
}
