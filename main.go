package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client/metadata"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/kelseyhightower/envconfig"
)

type EnvConfig struct {
	Target      string
	Port        int         `default:"8080"`
	Service     string      `default:"es"`
	EnableCors  bool        `envconfig:"enable_cors" default:"false"`
	CorsOptions CorsOptions `envconfig:"cors_options"`
}

type AppConfig struct {
	Service         string
	FlushInterval   time.Duration
	IdleConnTimeout time.Duration
	DialTimeout     time.Duration
	EnableCors      bool
	CorsOptions     CorsOptions
}

// Based on: https://github.com/rs/cors
type CorsOptions struct {
	// AllowedOrigins is a list of origins a cross-domain request can be executed from.
	// If the special "*" value is present in the list, all origins will be allowed.
	// Default value is ["*"]
	AllowedOrigins []string `envconfig:"allowed_origins" default:"*"`
}

func newDirector(target *url.URL, creds *credentials.Credentials, region string, appConfig AppConfig) func(req *http.Request) {
	return func(req *http.Request) {
		// Rewrite request to desired server host
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host

		// To perform the signing, we leverage aws-sdk-go
		// aws.request performs more functions than we need here
		// we only populate enough of the fields to successfully
		// sign the request
		config := aws.NewConfig().WithCredentials(creds).WithRegion(region)

		clientInfo := metadata.ClientInfo{
			ServiceName: appConfig.Service,
		}

		operation := &request.Operation{
			Name:       "",
			HTTPMethod: req.Method,
			HTTPPath:   req.URL.Path,
		}

		handlers := request.Handlers{}
		handlers.Sign.PushBack(v4.SignSDKRequest)

		// Do we need to use request.New ? Or can we create a raw Request struct and
		//  jus swap out the HTTPRequest with our own existing one?
		awsReq := request.New(*config, clientInfo, handlers, nil, operation, nil, nil)
		// Referenced during the execution of awsReq.Sign():
		//  req.Config.Credentials
		//  req.Config.LogLevel.Value()
		//  req.Config.Logger
		//  req.ClientInfo.SigningRegion (will default to Config.Region)
		//  req.ClientInfo.SigningName (will default to ServiceName)
		//  req.ClientInfo.ServiceName
		//  req.HTTPRequest
		//  req.Time
		//  req.ExpireTime
		//  req.Body

		// Set the body in the awsReq for calculation of body Digest
		// iotuil.ReadAll reads the Body from the stream so it can be copied into awsReq
		// This drains the body from the original (proxied) request.
		// To fix, we replace req.Body with a copy (NopCloser provides io.ReadCloser interface)
		if req.Body != nil {
			buf, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Printf("error reading request body: %v\n", err)
			}
			req.Body = ioutil.NopCloser(bytes.NewBuffer(buf))

			awsReq.SetBufferBody(buf)
		}

		// Use the updated req.URL for creating the signed request
		// We pass the full URL object to include Host, Scheme, and any params
		awsReq.HTTPRequest.URL = req.URL
		// These are now set above via req, but it's imperative that this remains
		//  correctly set before calling .Sign()
		//awsReq.HTTPRequest.URL.Scheme = target.Scheme
		//awsReq.HTTPRequest.URL.Host = target.Host

		// Perform the signing, updating awsReq in place
		if err := awsReq.Sign(); err != nil {
			log.Printf("error signing: %v\n", err)
		}

		// Write the Signed Headers into the Original Request
		for k, v := range awsReq.HTTPRequest.Header {
			req.Header[k] = v
		}
	}
}

// NewSigningProxy proxies requests to AWS services which require URL signing using the provided credentials
func NewSigningProxy(target *url.URL, creds *credentials.Credentials, region string, appConfig AppConfig) *httputil.ReverseProxy {
	director := newDirector(target, creds, region, appConfig)

	// transport is http.DefaultTransport but with the ability to override some
	// timeouts
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   appConfig.DialTimeout,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:        100,
		IdleConnTimeout:     appConfig.IdleConnTimeout,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	proxy := &httputil.ReverseProxy{
		Director:      director,
		FlushInterval: appConfig.FlushInterval,
		Transport:     transport,
	}

	if appConfig.EnableCors {
		proxy.ModifyResponse = func(res *http.Response) error {
			res.Header.Add("Access-Control-Allow-Origin", strings.Join(appConfig.CorsOptions.AllowedOrigins, ","))
			return nil
		}
	}

	return proxy
}

func main() {
	// Adding envconfig to allow setting key vars via ENV
	var e EnvConfig
	if err := envconfig.Process("ASP", &e); err != nil {
		log.Fatal(err.Error())
	}
	if val, ok := os.LookupEnv("AWS_ES_TARGET"); ok {
		e.Target = val
	}

	var targetFlag = flag.String("target", e.Target, "target url to proxy to")
	var portFlag = flag.Int("port", e.Port, "listening port for proxy")
	var serviceFlag = flag.String("service", e.Service, "AWS Service.")
	var regionFlag = flag.String("region", os.Getenv("AWS_REGION"), "AWS region for credentials")
	var flushInterval = flag.Duration("flush-interval", 0, "Flush interval to flush to the client while copying the response body.")
	var idleConnTimeout = flag.Duration("idle-conn-timeout", 90*time.Second, "the maximum amount of time an idle (keep-alive) connection will remain idle before closing itself. Zero means no limit.")
	var dialTimeout = flag.Duration("dial-timeout", 30*time.Second, "The maximum amount of time a dial will wait for a connect to complete.")
	var enableCors = flag.Bool("cors", e.EnableCors, "Whether or not to enable CORS on the responses")

	flag.Parse()

	appC := AppConfig{
		Service:         *serviceFlag,
		FlushInterval:   *flushInterval,
		IdleConnTimeout: *idleConnTimeout,
		DialTimeout:     *dialTimeout,
		EnableCors:      *enableCors,
		CorsOptions:     e.CorsOptions,
	}

	// Validate target URL
	if len(*targetFlag) == 0 {
		log.Fatal("Requires target URL to proxy to. Please use the -target flag")
	}
	targetURL, err := url.Parse(*targetFlag)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Get credentials:
	// Environment variables > local aws config file > remote role provider
	// https://github.com/aws/aws-sdk-go/blob/master/aws/defaults/defaults.go#L88
	creds := defaults.CredChain(defaults.Config(), defaults.Handlers())
	if _, err = creds.Get(); err != nil {
		// We couldn't get any credentials
		fmt.Println(err)
		return
	}

	// Region order of precident:
	// regionFlag > os.Getenv("AWS_REGION") > "us-west-2"
	region := *regionFlag
	if len(region) == 0 {
		region = "us-west-2"
	}

	// Start the proxy server
	proxy := NewSigningProxy(targetURL, creds, region, appC)
	listenString := fmt.Sprintf(":%v", *portFlag)
	fmt.Printf("Listening on %v\n", listenString)
	if err := http.ListenAndServe(listenString, proxy); err != nil {
		panic(err)
	}
}
