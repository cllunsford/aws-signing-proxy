package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	//"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client/metadata"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/private/signer/v4"
)

var targetFlag = flag.String("target", "", "target url to proxy to")
var portFlag = flag.Int("port", 8080, "listening port for proxy")
var regionFlag = flag.String("region", os.Getenv("AWS_REGION"), "AWS region for credentials")

func NewSigningProxy(target *url.URL, creds *credentials.Credentials, region string) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		// Rewrite request to desired server host
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host

		// To perform the signing, we leverage aws-sdk-go
		// aws.request performs more functions than we need here
		// we only populate enough of the fields to successfully
		// sign the request
		config := aws.NewConfig().WithCredentials(creds)
		if len(strings.TrimSpace(region)) > 0 {
			config = config.WithRegion(region)
		} else {
			config = config.WithRegion("us-west-2")
		}

		clientInfo := metadata.ClientInfo{
			ServiceName: "es",
		}

		operation := &request.Operation{
			Name:       "",
			HTTPMethod: req.Method,
			HTTPPath:   req.URL.Path,
		}

		handlers := request.Handlers{}
		handlers.Sign.PushBack(v4.Sign)

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

		// !!! Need to set the body via req.SetBufferBody([]byte) to fix signing hmac
		//if b, err := ioutil.ReadAll(req.Body); err == nil {
		//	awsReq.SetBufferBody(b)
		//}

		// Setting the URL.Host manually. v4.Sign() will create the Host Header
		//  from this value
		awsReq.HTTPRequest.URL.Scheme = target.Scheme
		awsReq.HTTPRequest.URL.Host = target.Host

		// Perform the signing, updating awsReq in place
		if err := awsReq.Sign(); err != nil {
			fmt.Println(err)
		}

		// Write the Signed Headers into the Original Request
		for k, v := range awsReq.HTTPRequest.Header {
			req.Header[k] = v
		}
	}

	return &httputil.ReverseProxy{
		Director: director,
	}
}

func main() {
	flag.Parse()

	targetUrl, err := url.Parse(*targetFlag)
	if err != nil {
		fmt.Println(err)
	}

	creds := credentials.NewEnvCredentials()
	region := *regionFlag

	proxy := NewSigningProxy(targetUrl, creds, region)
	listenString := fmt.Sprintf(":%v", *portFlag)
	fmt.Printf("Listening on %v\n", listenString)
	http.ListenAndServe(listenString, proxy)
}
