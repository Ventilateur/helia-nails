FROM --platform=linux/amd64 golang:1.21 as build

WORKDIR /app

# Copy dependencies list
COPY go.mod go.sum ./

# Build with optional lambda.norpc tag
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o main

# Copy artifacts to a clean image
FROM --platform=linux/amd64 public.ecr.aws/lambda/provided:al2023

COPY --from=build /app/main ./main

ENTRYPOINT [ "./main" ]
