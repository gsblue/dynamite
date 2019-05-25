#!/usr/bin/env bash
set -e

function createTable() {
    echo "create sample table"
    aws dynamodb create-table --cli-input-json file://test/sample-table-def.json --region ap-southeast-2 --endpoint-url http://localhost:4569

    echo "populate sample table"
    aws dynamodb batch-write-item --request-items file://test/sample-table-data.json --region ap-southeast-2 --endpoint-url http://localhost:4569
}


function createBucket() {
    echo "create test backup s3 bucket"
    aws s3 mb s3://test --endpoint-url http://localhost:4572
}

function verifyBucket() {
   export bucketName="s3://test/$(date +%F)/"
   echo "verifying backup bucket ${bucketName} is created"
   aws s3 ls "${bucketName}" --endpoint-url http://localhost:4572

}

function testArchive() {
    echo "test archive"
    ./dynamotools archive -t MusicCollection2 -b /test -tf ./test/sample.so -local
}

function run() {
    createTable
    createBucket
    testArchive
    verifyBucket
}

run


