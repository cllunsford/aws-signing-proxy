aws-signing-proxy
=================

aws-signing-proxy is a proxy service, written in go, for automatically signing requests made to AWS endpoints.  It leverages the aws-sdk-go library to sign requests to arbitrary URLs in AWS.  I wrote it to connect a kibana instance to an AWS Elasticsearch cluster using an IAM role instead of hard-coding IPs in the access policy.  Other uses may exist.

## Usage

```
export AWS_ACCESS_KEY_ID=<xxx>
export AWS_SECRET_ACCESS_KEY=<xxx>
./aws-signing-proxy -target https://search-my-cluster.us-west-2.es.amazonaws.com
```

## License

MIT 2016 (c) Chris Lunsford
