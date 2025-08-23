#!/bin/bash

set -e

echo "Creating localstack resources"

# Create SQS queue for log cleanup async processing
awslocal sqs create-queue \
  --queue-name log-cleanup-queue \
  --attributes VisibilityTimeout=300,MessageRetentionPeriod=86400,DelaySeconds=0,ReceiveMessageWaitTimeSeconds=20
echo "SQS queue 'log-cleanup-queue' created!"

awslocal sqs create-queue \
  --queue-name log-archival-queue \
  --attributes VisibilityTimeout=300,MessageRetentionPeriod=86400,DelaySeconds=0,ReceiveMessageWaitTimeSeconds=20
echo "SQS queue 'log-archival-queue' created!"

awslocal sqs create-queue \
  --queue-name index-queue \
  --attributes VisibilityTimeout=300,MessageRetentionPeriod=86400,DelaySeconds=0,ReceiveMessageWaitTimeSeconds=20
echo "SQS queue 'index-queue' created!"

# Create S3 Bucket for log archiving before deleting
awslocal s3 mb s3://log-archive
echo "S3 bucket 'log-archive' created!"
