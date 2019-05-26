package restore

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"golang.org/x/sync/errgroup"
)

const (
	localDynamoDBEndpoint = "http://localhost:4569"
	localS3Endpoint       = "http://localhost:4572"
)

// DynamoResotreConfig provides the configuration for archiving dynamo table to s3
type DynamoResotreConfig struct {
	Region          string
	TableName       string
	Workers         int
	Bucket          string
	RestoreFile     string
	RunOnLocalStack bool
}

// ToDyanmo restores the data from the file in the s3 bucket to the specified dynamo table
func ToDyanmo(c *DynamoResotreConfig) error {
	var dynamoEndpoint, s3Endpoint string
	if c.RunOnLocalStack {
		dynamoEndpoint = localDynamoDBEndpoint
		s3Endpoint = localS3Endpoint
	}

	dynamoSession, err := getNewAwsSession(c.Region, dynamoEndpoint)
	if err != nil {
		return err
	}
	s3Session, err := getNewAwsSession(c.Region, s3Endpoint)
	if err != nil {
		return err
	}
	dl := s3manager.NewDownloader(s3Session)

	localFile := fmt.Sprintf("restore-file-%s", time.Now().Format("2006-01-02"))
	file, err := os.Create(localFile)
	if err != nil {
		return err
	}

	defer file.Close()
	defer os.Remove(file.Name())

	log.Println("downloading restore file ....")
	_, err = dl.Download(file, &s3.GetObjectInput{
		Bucket: &c.Bucket,
		Key:    &c.RestoreFile,
	})

	if err != nil {
		return err
	}

	_, _ = file.Seek(0, 0)
	dec := json.NewDecoder(file)
	itemsChan := make(chan map[string]interface{})

	log.Println("starting dynmo writer")
	//db := dynamodb.New(s)
	grp, ctx := errgroup.WithContext(context.Background())

	log.Println("workers ", c.Workers)
	for index := 0; index < c.Workers; index++ {
		grp.Go(func() error {
			return NewDynamoBatchWriter(dynamodb.New(dynamoSession), c.TableName).Write(itemsChan)
		})
	}
	stop := false

	for {
		if stop {
			break
		}

		var items []map[string]interface{}
		err := dec.Decode(&items)
		if err == io.EOF {
			break
		}

		if err != nil {
			close(itemsChan)
			return err
		}
		for _, item := range items {
			select {
			case itemsChan <- item:
			case <-ctx.Done():
				stop = true
			}
		}
	}
	close(itemsChan)
	if err := grp.Wait(); err != nil {
		return err
	}
	log.Printf("completed restoring to %s", c.TableName)
	return nil
}

func getNewAwsSession(region, endpoint string) (*session.Session, error) {
	awsconfig := defaults.Config().WithRegion(region).WithEndpoint(endpoint) //.WithLogLevel(aws.LogDebugWithRequestErrors)
	awsconfig.Credentials = defaults.CredChain(awsconfig, defaults.Handlers())
	return session.NewSession(awsconfig)
}
