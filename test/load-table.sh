#!/usr/bin/env bash
awslocal dynamodb create-table --cli-input-json file://sample-table-def.json --region ap-southeast-2

awslocal dynamodb batch-write-item --request-items file://sample-table-data.json --region ap-southeast-2