dev:
	@echo $$AWS_ACCESS_KEY_ID
	jq '.profile |= "uneet-dev" |.stages.production |= (.domain = "prince.dev.unee-t.com" | .zone = "dev.unee-t.com")| .actions[0].emails |= ["kai.hendry+princedev@unee-t.com"] | .lambda.policy[1].Resource[0] = "arn:aws:s3:::dev-media-unee-t/*"' up.json.in > up.json
	up deploy production

demo:
	@echo $$AWS_ACCESS_KEY_ID
	jq '.profile |= "uneet-demo" |.stages.production |= (.domain = "prince.demo.unee-t.com" | .zone = "demo.unee-t.com") | .actions[0].emails |= ["kai.hendry+princedemo@unee-t.com"] | .lambda.policy[1].Resource[0] = "arn:aws:s3:::demo-media-unee-t/*"' up.json.in > up.json
	up deploy production

prod:
	@echo $$AWS_ACCESS_KEY_ID
	jq '.profile |= "uneet-prod" |.stages.production |= (.domain = "prince.unee-t.com" | .zone = "unee-t.com")| .actions[0].emails |= ["kai.hendry+princeprod@unee-t.com"] | .lambda.policy[1].Resource[0] = "arn:aws:s3:::prod-media-unee-t/*"' up.json.in > up.json
	up deploy production

localtest:
	curl -X POST -d '{ "document_url": "https://media.dev.unee-t.com/2018-10-11/12345678-5c33ae52.html" }' localhost:3000

remotetest:
	curl -X POST -d '{ "document_url": "https://s3-ap-southeast-1.amazonaws.com/dev-media-unee-t/2018-08-17/tee.html" }' https://prince.dev.unee-t.com

logs:
	up logs production
