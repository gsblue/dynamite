package cmd

import (
	"errors"
	"fmt"
	"plugin"

	"github.com/gsblue/dynamotools/archive"
	"github.com/urfave/cli"
)

// BuildArchive builds the cli command for archive funationality
func BuildArchive() cli.Command {
	return cli.Command{
		Name: "archive",
		Usage: `region [aws region name] table [dynamo table name] tableindex [index to use for scanning] 
						partitions [scan partitions for parallel scanning] limit [limit for scanning no of records] 
						bucket [s3 bucket name] chunksize [chunk sizes (in MB) to be uploaded to the bucket] 
						concurrency [concurrency for uploads to the bucket]`,
		Description: "archive scans the [table] using the specified [tableindex] and saves it the s3 [bucket]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "region, r",
				Value: "ap-southeast-2",
				Usage: "aws region name where your dynamodb table and s3 bucket is",
			},
			cli.StringFlag{
				Name:  "table, t",
				Usage: "dynamodb table name",
			},
			cli.StringFlag{
				Name:  "tableindex, i",
				Usage: "index for scanning the dynamo table",
			},
			cli.IntFlag{
				Name:  "partitions, p",
				Value: 1,
				Usage: "partitions for parallel scanning",
			},
			cli.IntFlag{
				Name:  "limit, l",
				Value: 100,
				Usage: "limit for scanning records",
			},
			cli.StringFlag{
				Name:  "filtername, fn",
				Usage: "name of the scan filter attribute",
			},
			cli.StringFlag{
				Name:  "filtertype, ft",
				Usage: "type of the scan filter attribute (string|number)",
			},
			cli.StringFlag{
				Name:  "filteroperator, fo",
				Usage: "operator for the scan filter ( < | = | > )",
			},
			cli.StringFlag{
				Name:  "filtervalue, fv",
				Usage: "value for the scan filter",
			},
			cli.StringFlag{
				Name:  "bucket, b",
				Usage: "name of the bucket to store the archived data",
			},
			cli.Int64Flag{
				Name:  "chunksize, cs",
				Value: 16,
				Usage: "chunk sizes (in MB) to be uploaded to the bucket",
			},
			cli.Int64Flag{
				Name:  "concurrency, uc",
				Value: 10,
				Usage: "concurrency for uploads to the bucket",
			},
			cli.StringFlag{
				Name:  "prefix, pf",
				Usage: "folder where archived data will be stored (optional)",
			},
			cli.StringFlag{
				Name:  "transform, tf",
				Usage: "`.so` plugin file path for archive data transformation",
			},
			cli.BoolFlag{
				Name:  "local",
				Usage: "tool runs against https://github.com/localstack/localstack",
			},
		},
		SkipFlagParsing: false,
		Before: func(c *cli.Context) error {
			if c.String("table") == "" && c.String("t") == "" {
				return cli.NewExitError("missing value for [table]", 86)
			} else if c.String("bucket") == "" && c.String("b") == "" {
				return cli.NewExitError("missing value for [bucket]", 86)
			}
			return nil
		},
		Action: func(c *cli.Context) error {
			var transformer archive.Transformer
			if c.String("transform") != "" {
				var err error
				transformer, err = parseTransformPlugin(c.String("transform"))
				if err != nil {
					return cli.NewExitError(fmt.Sprintf("failed loading tranformation plugin: %s", err), 86)
				}
			}
			return archive.ToS3(&archive.S3ArchiveConfig{
				Region:             c.String("region"),
				TableName:          c.String("table"),
				TableIndex:         c.String("tableindex"),
				ScanPartitions:     c.Int("partitions"),
				ScanLimit:          c.Int("limit"),
				ScanFilterName:     c.String("filtername"),
				ScanFilterType:     c.String("filtertype"),
				ScanFilterOperator: c.String("filteroperator"),
				ScanFilterValue:    c.String("filtervalue"),
				UploadBucket:       c.String("bucket"),
				UploadChunkSize:    c.Int64("chunksize"),
				UploadConcurrency:  c.Int("concurrency"),
				BackupPrefix:       c.String("prefix"),
				Transformer:        transformer,
				RunOnLocalStack:    c.Bool("local"),
			})

		},
	}
}

func parseTransformPlugin(path string) (archive.Transformer, error) {
	tfPlugin, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}
	symTransformer, err := tfPlugin.Lookup("Transformer")
	if err != nil {
		return nil, err
	}

	transformer, ok := symTransformer.(archive.Transformer)
	if !ok {
		return nil, errors.New("plugin does not implement the interface `Transform(input []map[string]interface{}) []map[string]interface{}`")
	}
	return transformer, nil
}
