package main

type (
	// Config for the kube server.
	Config struct {
		Server    string
		SkipTLS   bool
		CaCert    string
		Token     string
		Namespace string
	}

	// Plugin values.
	Plugin struct {
		Config *Config
	}
)

func (p *Plugin) Exec() error {
	return nil
}
