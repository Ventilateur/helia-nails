
AWS_ACCOUNT_ID := 851725466447
VERSION := 5
IMAGE_TAG := $(AWS_ACCOUNT_ID).dkr.ecr.eu-west-3.amazonaws.com/helia-nails/calendar-sync:$(VERSION)

build:
	docker build --platform=linux/amd64 . -t $(IMAGE_TAG)

push:
	docker push $(IMAGE_TAG)

plan:
	cd tf && terraform plan -var 'image_version=$(VERSION)'

apply:
	cd tf && terraform apply -var 'image_version=$(VERSION)'
