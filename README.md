# relay-webhook
This service accepts webhooks from Rancher, converts them to simple text
and then posts them to MS Teams

The service can relay webhooks to any number of targets.  The config file contains a list of target webhooks.

The URL that rancher posts to includes the target name and is in the following format:

http://myhost.example.com:8080/relay/{target}