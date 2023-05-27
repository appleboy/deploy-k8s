package main

import (
	"log"
	"os"
	"strconv"
	"time"

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
			Usage:   "kubernetes server",
			EnvVars: []string{"PLUGIN_SERVER", "INPUT_SERVER"},
			Value:   "us-east-1",
		},
		&cli.BoolFlag{
			Name:    "skip-tls",
			Usage:   "Skip TLS verify",
			EnvVars: []string{"PLUGIN_SKIP_TLS_VERIFY", "INPUT_SKIP_TLS_VERIFY"},
		},
		&cli.StringFlag{
			Name:    "ca-cert",
			Usage:   "ca cert raw content",
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
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	plugin := &Plugin{
		Config: &Config{
			Server:    c.String("server"),
			SkipTLS:   c.Bool("skip-tls"),
			CaCert:    c.String("ca-cert"),
			Token:     c.String("token"),
			Namespace: c.String("namespace"),
		},
	}

	return plugin.Exec()
}
