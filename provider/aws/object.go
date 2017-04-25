package aws

import (
	"bytes"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/convox/praxis/types"
)

func (p *Provider) ObjectExists(app, key string) (bool, error) {
	bucket, err := p.appResource(app, "Bucket")
	if err != nil {
		return false, err
	}

	_, err = p.S3().HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err, ok := err.(awserr.Error); ok && err.Code() == "NotFound" {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (p *Provider) ObjectFetch(app, key string) (io.ReadCloser, error) {
	bucket, err := p.appResource(app, "Bucket")
	if err != nil {
		return nil, err
	}

	res, err := p.S3().GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if awsError(err) == "NoSuchKey" {
		return nil, fmt.Errorf("no such key: %s", key)
	}
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func (p *Provider) ObjectStore(app, key string, r io.Reader, opts types.ObjectStoreOptions) (*types.Object, error) {
	if key == "" {
		return nil, fmt.Errorf("key must not be blank")
	}

	bucket, err := p.appResource(app, "Bucket")
	if err != nil {
		return nil, err
	}

	mreq := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	if opts.Public {
		mreq.ACL = aws.String("public-read")
	}

	mres, err := p.S3().CreateMultipartUpload(mreq)
	if err != nil {
		return nil, err
	}

	// buf := make([]byte, 5*1024*1024)
	buf := make([]byte, 10*1024*1024)
	i := 1
	parts := []*s3.CompletedPart{}

	for {
		n, err := io.ReadFull(r, buf)
		if err == io.EOF {
			break
		}
		if err != nil && err != io.ErrUnexpectedEOF {
			return nil, err
		}

		res, err := p.S3().UploadPart(&s3.UploadPartInput{
			Body:          bytes.NewReader(buf[0:n]),
			Bucket:        aws.String(bucket),
			ContentLength: aws.Int64(int64(n)),
			Key:           aws.String(key),
			PartNumber:    aws.Int64(int64(i)),
			UploadId:      mres.UploadId,
		})
		if err != nil {
			return nil, err
		}

		parts = append(parts, &s3.CompletedPart{
			ETag:       res.ETag,
			PartNumber: aws.Int64(int64(i)),
		})

		i++
	}

	_, err = p.S3().CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: parts,
		},
		UploadId: mres.UploadId,
	})
	if err != nil {
		return nil, err
	}

	return &types.Object{Key: key}, nil
}
