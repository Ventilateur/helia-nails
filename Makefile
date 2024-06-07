AWS_DEFAULT_PROFILE := helia
AWS_REGION := eu-west-3
AWS_ACCOUNT_ID := 851725466447
LAYER_ARN := "arn:aws:lambda:eu-west-3:780235371811:layer:AWS-Parameters-and-Secrets-Lambda-Extension:11"

VERSION := 15
IMAGE_TAG := $(AWS_ACCOUNT_ID).dkr.ecr.eu-west-3.amazonaws.com/helia-nails/calendar-sync:$(VERSION)

build:
	docker build --platform=linux/amd64 . -t $(IMAGE_TAG) \
		--build-arg AWS_REGION=$(AWS_REGION) \
		--build-arg AWS_ACCESS_KEY_ID=$$(aws configure get aws_access_key_id) \
		--build-arg AWS_SECRET_ACCESS_KEY=$$(aws configure get aws_secret_access_key) \
		--build-arg LAYER_ARN=$(LAYER_ARN)

push:
	docker push $(IMAGE_TAG)

plan:
	cd tf && terraform plan -var 'image_version=$(VERSION)'

apply:
	cd tf && terraform apply -auto-approve -var 'image_version=$(VERSION)'
