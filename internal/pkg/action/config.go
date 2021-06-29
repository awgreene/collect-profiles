package action

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Configuration struct {
	RESTConfig *rest.Config
	Client     client.Client
	Scheme     *runtime.Scheme
}

func (c *Configuration) Load() error {
	// creates the in-cluster config
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	sch := scheme.Scheme
	cl, err := client.New(restConfig, client.Options{
		Scheme: sch,
	})
	if err != nil {
		return err
	}

	c.Scheme = scheme.Scheme
	c.Client = cl
	c.RESTConfig = restConfig

	return nil
}
