FROM --platform=linux/amd64 golang:1.21-alpine as build

ARG AWS_REGION
ARG AWS_ACCESS_KEY_ID
ARG AWS_SECRET_ACCESS_KEY
ARG LAYER_ARN

ENV AWS_REGION=${AWS_REGION}
ENV AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
ENV AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}

WORKDIR /app

# Download aws lambda layer
RUN apk add aws-cli curl unzip
RUN curl $(aws lambda get-layer-version-by-arn --arn ${LAYER_ARN} --query 'Content.Location' --output text) --output layer.zip
RUN unzip layer.zip -d /opt
RUN rm layer.zip

# Copy dependencies list
COPY go.mod go.sum ./

# Build with optional lambda.norpc tag
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o main

# Copy artifacts to a clean image
FROM --platform=linux/amd64 public.ecr.aws/lambda/provided:al2023

COPY --from=build /app/main ./main
COPY --from=build /opt /opt

ENTRYPOINT [ "./main" ]
