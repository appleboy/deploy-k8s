package config

type (
	// Config for the kube server.
	K8S struct {
		Server    string
		SkipTLS   bool
		CaCert    string
		Token     string
		Namespace string
		ProxyURL  string
		Templates []string
		Output    string

		// kube config file
		ClusterName  string
		AuthInfoName string
		ContextName  string
	}
)