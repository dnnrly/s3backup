#!/usr/bin/env bash

echo "Waiting for S3 availability"

sleep 60

export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_REGION=eu-west-1

export bucket=dnnrly-sync-01
echo "Creating bucket ${bucket}"
aws --endpoint-url http://localstack:4572 s3api create-bucket --bucket ${bucket} --region eu-west-1

eval $@