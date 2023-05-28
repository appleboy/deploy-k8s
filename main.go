package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/appleboy/deploy-k8s/config"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/cli/v2"
)

// Version set at compile-time
var (
	Version string
)

func main() {
	// Load env-file if it exists first
	if filename, found := os.LookupEnv("PLUGIN_ENV_FILE"); found {
		_ = godotenv.Load(filename)
	}

	app := cli.NewApp()
	app.Name = "Deploy Kubernetes plugin"
	app.Usage = "Deploy Kubernetes plugin"
	app.Copyright = "Copyright (c) " + strconv.Itoa(time.Now().Year()) + " Bo-Yi Wu"
	app.Authors = []*cli.Author{
		{
			Name:  "Bo-Yi Wu",
			Email: "appleboy.tw@gmail.com",
		},
	}
	app.Action = run
	app.Version = Version
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "server",
			Usage:   "Server is the address of the kubernetes cluster (https://hostname:port).",
			EnvVars: []string{"PLUGIN_SERVER", "INPUT_SERVER"},
		},
		&cli.BoolFlag{
			Name:    "skip-tls",
			Usage:   "InsecureSkipTLSVerify skips the validity check for the server's certificate.",
			EnvVars: []string{"PLUGIN_SKIP_TLS_VERIFY", "INPUT_SKIP_TLS_VERIFY"},
		},
		&cli.StringFlag{
			Name:    "ca-cert",
			Usage:   "CertificateAuthorityData contains PEM-encoded certificate authority certificates.",
			EnvVars: []string{"PLUGIN_CA_CERT", "INPUT_CA_CERT"},
		},
		&cli.StringFlag{
			Name:    "token",
			Usage:   "kubernetes service account token",
			EnvVars: []string{"PLUGIN_TOKEN", "INPUT_TOKEN"},
		},
		&cli.StringFlag{
			Name:    "namespace",
			Usage:   "kubernetes namespace",
			EnvVars: []string{"PLUGIN_NAMESPACE", "INPUT_NAMESPACE"},
		},
		&cli.StringFlag{
			Name:    "proxy-url",
			Usage:   "URLs with http, https, and socks5",
			EnvVars: []string{"PLUGIN_PROXY_URL", "INPUT_PROXY_URL"},
		},
		&cli.StringSliceFlag{
			Name:    "templates",
			Usage:   "template files, support glob pattern",
			EnvVars: []string{"PLUGIN_TEMPLATES", "INPUT_TEMPLATES"},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	plugin := &Plugin{
		Config: &config.K8S{
			Server:    c.String("server"),
			SkipTLS:   c.Bool("skip-tls"),
			CaCert:    c.String("ca-cert"),
			Token:     c.String("token"),
			Namespace: c.String("namespace"),
			ProxyURL:  c.String("proxy-url"),
			Templates: c.StringSlice("templates"),
		},
	}

	return plugin.Exec()
}
