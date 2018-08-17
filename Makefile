deploy:
	up deploy production

localtest:
	curl -X POST -d '{ "document_url": "https://s3-ap-southeast-1.amazonaws.com/dev-media-unee-t/2018-08-17/tee.html" }' localhost:3000

logs:
	up logs production
