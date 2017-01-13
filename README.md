aws-signing-proxy
=================

aws-signing-proxy is a proxy service, written in go, for automatically signing requests made to AWS endpoints.  It leverages the aws-sdk-go library to sign requests to arbitrary URLs in AWS.  I wrote it to connect a kibana instance to an AWS Elasticsearch cluster using an IAM role instead of hard-coding IPs in the access policy.  Other uses may exist.

## Usage

```
export AWS_ACCESS_KEY_ID=<xxx>
export AWS_SECRET_ACCESS_KEY=<xxx>
export AWS_REGION=<xxx>
export LOGIN_USER=<xxx>
export LOGIN_HASHED_PASSWORD=<xxx>
./aws-signing-proxy -target https://search-my-cluster.us-west-2.es.amazonaws.com

-target can also be passed as env var: export TARGET_URL=<xxx>

```
### Credential chain

AWS credentials are looked up in the following order:
1. Environment variable accessible to the app
2. The `~/.aws/credentials` config file's default profile
3. Any [IAM instance profile role](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use_switch-role-ec2.html) assigned to the instance.

## Notes, Tips

### Signature Expired

If you see:

`{"message":"Signature expired: 20160415T172935Z is now earlier than 20160415T174424Z (20160415T174924Z - 5 min.)"}`

verify that the clock/time is in sync on the proxy host.

### Kibana Forbidden index write

For AWS Elasticsearch, the built-in kibana populates the .kibana index.  If you see:

`ClusterBlockException[blocked by: [FORBIDDEN/8/index write (api)];]`

try changing the kibana index setting to use a different index.  The [marcbachmann/kibana4](https://github.com/marcbachmann/dockerfile-kibana4) docker image allows you to change this easily by setting the ```KIBANA_INDEX``` environment variable.


## License

MIT 2016 (c) Chris Lunsford
