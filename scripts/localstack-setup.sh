#!/usr/bin/env bash

echo "Waiting for S3 availability"
echo ""

for i in {1..60} ; do
    echo -e "\e[1A\r${i}/60"
    sleep 1
done

echo "Done waiting"

export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_REGION=eu-west-1

export bucket=test-bucket
echo "Creating bucket ${bucket}"
aws --endpoint-url http://localstack:4572 s3api create-bucket --bucket ${bucket} --region eu-west-1

eval $@