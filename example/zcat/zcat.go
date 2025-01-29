/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
	"ztna-core/sdk-golang/ziti"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/info"
	"github.com/openziti/transport/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	pfxlog.GlobalInit(logrus.InfoLevel, pfxlog.DefaultOptions().SetTrimPrefix("github.com/openziti/"))
}

var verbose bool
var logFormatter string
var retry bool
var identityFile string
var ctrlProxy string
var routerProxy string

func init() {
	root.PersistentFlags().StringVarP(&identityFile, "identity", "i", "", "Identity file path")
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	root.PersistentFlags().BoolVarP(&retry, "retry", "r", false, "Retry after i/o error")
	root.PersistentFlags().StringVar(&logFormatter, "log-formatter", "", "Specify log formatter [json|pfxlog|text]")
	root.PersistentFlags().StringVar(&ctrlProxy, "ctrl-proxy", "", "Specify a proxy to use for controller connections")
	root.PersistentFlags().StringVar(&routerProxy, "router-proxy", "", "Specify a proxy to use for router connections")
}

var root = &cobra.Command{
	Use:   "zcat <service>",
	Short: "Ziti Netcat",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}

		switch logFormatter {
		case "pfxlog":
			logrus.SetFormatter(pfxlog.NewFormatter(pfxlog.DefaultOptions().StartingToday()))
		case "json":
			logrus.SetFormatter(&logrus.JSONFormatter{})
		case "text":
			logrus.SetFormatter(&logrus.TextFormatter{})
		default:
			// let logrus do its own thing
		}
	},
	Args: cobra.RangeArgs(1, 2),
	Run:  runFunc,
}

func main() {
	if err := root.Execute(); err != nil {
		fmt.Printf("error: %s", err)
	}
}

func runFunc(_ *cobra.Command, args []string) {
	log := pfxlog.Logger()
	service := args[0]

	// Get identity config
	cfg, err := ziti.NewConfigFromFile(identityFile)
	if err != nil {
		panic(err)
	}

	if ctrlProxy != "" {
		fmt.Printf("using controller proxy: %s\n", ctrlProxy)
		cfg.CtrlProxy = func(request *http.Request) (*url.URL, error) {
			return url.Parse(ctrlProxy)
		}
	}

	if routerProxy != "" {
		fmt.Printf("using router proxy: %s\n", routerProxy)
		cfg.RouterProxy = func(addr string) *transport.ProxyConfiguration {
			return &transport.ProxyConfiguration{
				Type:    transport.ProxyTypeHttpConnect,
				Address: routerProxy,
			}
		}
	}

	context, err := ziti.NewContext(cfg)

	if err != nil {
		panic(err)
	}

	for {
		opts := &ziti.DialOptions{
			ConnectTimeout: 5 * time.Second,
		}
		if len(args) >= 2 {
			opts.Identity = args[1]
		}
		conn, err := context.DialWithOptions(service, opts)
		if err != nil {
			if retry {
				log.WithError(err).Errorf("unable to dial service: '%v'", service)
				log.Info("retrying in 5 seconds")
				time.Sleep(5 * time.Second)
			} else {
				log.WithError(err).Fatalf("unable to dial service: '%v'", service)
			}
		} else {
			pfxlog.Logger().Info("connected")
			go Copy(conn, os.Stdin)
			Copy(os.Stdout, conn)
			_ = conn.Close()
			if !retry {
				return
			}
		}
	}
}

func Copy(writer io.Writer, reader io.Reader) {
	buf := make([]byte, info.MaxUdpPacketSize)
	bytesCopied, err := io.CopyBuffer(writer, reader, buf)
	pfxlog.Logger().Infof("Copied %v bytes", bytesCopied)
	if err != nil {
		pfxlog.Logger().Errorf("error while copying bytes (%v)", err)
	}
}
