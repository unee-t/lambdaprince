deploy:
	up deploy production

localtest:
	curl -X POST -d '{ "document_url": "https://s3-ap-southeast-1.amazonaws.com/dev-media-unee-t/2018-08-17/tee.html" }' localhost:3000

remotetest:
	curl -X POST -d '{ "document_url": "https://s3-ap-southeast-1.amazonaws.com/dev-media-unee-t/2018-08-17/tee.html" }' https://prince.dev.unee-t.com

logs:
	up logs production
