aws-signing-proxy
=================

aws-signing-proxy is a proxy service, written in go, for automatically signing requests made to AWS endpoints.  It leverages the aws-sdk-go library to sign requests to arbitrary URLs in AWS.  I wrote it to connect a kibana instance to an AWS Elasticsearch cluster using an IAM role instead of hard-coding IPs in the access policy.  Other uses may exist.

## Usage

```
./aws-signing-proxy -target https://my-cluster.es.us-west-2.aws.com
```

## License

MIT 2016 (c) Chris Lunsford
