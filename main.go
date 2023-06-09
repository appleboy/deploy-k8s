package main

import (
	"os"
	"strconv"
	"time"

	"github.com/appleboy/deploy-k8s/config"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

	isTerm := isatty.IsTerminal(os.Stdout.Fd())
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:     os.Stderr,
			NoColor: !isTerm,
		},
	)
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}

	app := cli.NewApp()
	app.Name = "Deploy Kubernetes plugin"
	app.Usage = "Generate a Kubeconfig or creating & updating K8s Deployments."
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
		&cli.StringSliceFlag{
			Name:    "deployment",
			Usage:   "Name of the Kubernetes deployment to update",
			EnvVars: []string{"PLUGIN_DEPLOYMENT", "INPUT_DEPLOYMENT"},
		},
		&cli.StringSliceFlag{
			Name:    "container",
			Usage:   "Name of the container within the deployment to update",
			EnvVars: []string{"PLUGIN_CONTAINER", "INPUT_CONTAINER"},
		},
		&cli.StringFlag{
			Name:    "image",
			Usage:   "New image and tag for the container",
			EnvVars: []string{"PLUGIN_IMAGE", "INPUT_IMAGE"},
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
		&cli.StringFlag{
			Name:    "output",
			Usage:   "Generate Kubernetes config file",
			EnvVars: []string{"PLUGIN_OUTPUT", "INPUT_OUTPUT"},
		},
		&cli.StringFlag{
			Name:    "cluster-name",
			Usage:   "",
			EnvVars: []string{"PLUGIN_CLUSTER_NAME", "INPUT_CLUSTER_NAME"},
			Value:   "default",
		},
		&cli.StringFlag{
			Name:    "authinfo-name",
			Usage:   "",
			EnvVars: []string{"PLUGIN_AUTHINFO_NAME", "INPUT_AUTHINFO_NAME"},
			Value:   "default",
		},
		&cli.StringFlag{
			Name:    "context-name",
			Usage:   "",
			EnvVars: []string{"PLUGIN_CONTEXT_NAME", "INPUT_CONTEXT_NAME"},
			Value:   "default",
		},
		&cli.BoolFlag{
			Name:    "debug",
			Usage:   "enable debug mode",
			EnvVars: []string{"PLUGIN_DEBUG", "INPUT_DEBUG"},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("can't run app")
	}
}

func run(c *cli.Context) error {
	if c.Bool("debug") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.With().Caller().Logger()
	}

	plugin := &Plugin{
		Config: &config.K8S{
			Server:       c.String("server"),
			SkipTLS:      c.Bool("skip-tls"),
			CaCert:       c.String("ca-cert"),
			Namespace:    c.String("namespace"),
			Deployment:   c.StringSlice("deployment"),
			Container:    c.StringSlice("container"),
			Image:        c.String("image"),
			ProxyURL:     c.String("proxy-url"),
			Templates:    c.StringSlice("templates"),
			Output:       c.String("output"),
			ClusterName:  c.String("cluster-name"),
			AuthInfoName: c.String("authinfo-name"),
			ContextName:  c.String("context-name"),
			Debug:        c.Bool("debug"),
		},
		AuthInfo: &config.AuthInfo{
			Token: c.String("token"),
		},
	}

	if plugin.Config.Debug {
		spew.Dump(plugin)
	}

	return plugin.Exec()
}
