package cos

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/tencentyun/cos-go-sdk-v5"

	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/store"
)

type Client struct {
	Client *cos.Client
}

// NewClient create a new COS client
//
// Refer to https://cloud.tencent.com/document/product/436/31215
func NewClient(config *store.StorageSetting) *Client {
	// https://cloud.tencent.com/document/product/436/31215
	bucketURLStr := fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com", config.Bucket, config.AccessKey, config.Region)
	serviceURLStr := fmt.Sprintf("https://cos.%s.myqcloud.com", config.Region)
	bucketURL, err := url.Parse(bucketURLStr)
	if err != nil {
		panic(err)
	}
	serviceURL, err := url.Parse(serviceURLStr)
	if err != nil {
		panic(err)
	}
	b := &cos.BaseURL{
		BucketURL:  bucketURL,
		ServiceURL: serviceURL,
	}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.AccessKey,
			SecretKey: config.SecretKey,
		},
	})

	return &Client{
		Client: client,
	}
}

// UploadObject upload object to COS, return the object URL
func (c *Client) UploadObject(ctx context.Context, key string, file io.Reader) (string, error) {
	_, err := c.Client.Object.Put(ctx, key, file, nil)
	if err != nil {
		log.Errorf("Failed to upload object to COS: %s", err)
		return "", err
	}

	return fmt.Sprintf("%s/%s", c.Client.BaseURL.BucketURL.String(), key), nil
}

func (c *Client) GetObjectURL(key string) string {
	u, _ := url.Parse(c.Client.BaseURL.BucketURL.String() + "/" + key)
	return u.String()
}

func (c *Client) DeleteObject(ctx context.Context, key string) error {
	_, err := c.Client.Object.Delete(ctx, key)
	if err != nil {
		log.Errorf("Failed to delete object from COS: %s", err)
		return err
	}
	return nil
}

// MoveObject copy object from srcKey to destKey and delete srcKey
func (c *Client) MoveObject(ctx context.Context, srcKey, destKey string) error {
	_, _, err := c.Client.Object.Copy(ctx, destKey, c.GetObjectURL(srcKey), nil)
	if err != nil {
		log.Errorf("Failed to move object from COS: %s", err)
		return err
	}
	err = c.DeleteObject(ctx, srcKey)
	if err != nil {
		log.Errorf("Failed to delete object from COS: %s", err)
		return err
	}
	return nil
}

func (c *Client) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	res, err := c.Client.Object.Get(ctx, key, nil)
	if err != nil {
		log.Errorf("Failed to get object from COS: %s", err)
		return nil, err
	}
	return res.Body, nil
}
