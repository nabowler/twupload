package twupload

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
)

type (
	uploader struct {
		httpClient *http.Client
		chunkSize  int
	}
)

const (
	UnknownMediaType sentinelError = "Unknown Media Type"
)

var (
	_ Uploader = (*uploader)(nil)
)

func (uploader uploader) Upload(ctx context.Context, data []byte) (MediaUploadResponse, error) {
	var upResponse MediaUploadResponse

	var category MediaCategory
	mimeType := mimetype.Detect(data).String()
	if strings.HasPrefix(mimeType, "video/") {
		category = MediaCategoryTweetVideo
	} else if mimeType == "image/gif" {
		category = MediaCategoryGif
	} else if strings.HasPrefix(mimeType, "image/") {
		category = MediaCategoryImage
	} else {
		return upResponse, UnknownMediaType
	}

	// uploading follows the flow
	//   POST INIT
	//   1..n POST APPEND
	//   POST FINALIZE
	//   while processing
	//      GET STATUS

	initResponse, err := uploader.initUpload(ctx, len(data), category, mimeType)
	if err != nil {
		return initResponse, err
	}

	for i := 0; i*uploader.chunkSize < len(data); i++ {
		low := i * uploader.chunkSize
		high := (i + 1) * uploader.chunkSize
		if high > len(data) {
			high = len(data)
		}

		err = uploader.uploadChunk(ctx, initResponse.MediaIDString, i, data[low:high])
		if err != nil {
			return initResponse, fmt.Errorf("Unable to upload chunk %d %w", i, err)
		}
	}

	upResponse, err = uploader.finalize(ctx, initResponse.MediaIDString)
	if err != nil {
		return upResponse, fmt.Errorf("Unable to finalize upload %w", err)
	}

	upResponse, err = uploader.handleProcessing(ctx, upResponse)
	if err != nil {
		return upResponse, err
	}

	return upResponse, nil
}

func (uploader uploader) initUpload(ctx context.Context, dataLen int, category MediaCategory, mimeType string) (MediaUploadResponse, error) {
	var upResponse MediaUploadResponse

	u, err := url.Parse(baseURL)
	if err != nil {
		return upResponse, err
	}
	q := u.Query()
	q.Set("media_category", string(category))
	q.Set("command", "INIT")
	q.Set("total_bytes", strconv.Itoa(dataLen))
	q.Set("media_type", mimeType)
	u.RawQuery = q.Encode()

	resp, err := post(ctx, uploader.httpClient, u.String(), url.Values{})
	if err != nil {
		return upResponse, err
	}

	err = handleHttpResponse(resp, &upResponse)
	if err != nil {
		return upResponse, err
	}

	return upResponse, nil
}

func (uploader uploader) uploadChunk(ctx context.Context, mediaIdString string, segment int, data []byte) error {
	encoded := base64.StdEncoding.EncodeToString(data)

	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}
	q := u.Query()
	q.Set("media_id", mediaIdString)
	q.Set("command", "APPEND")
	q.Set("segment_index", strconv.Itoa(segment))
	u.RawQuery = q.Encode()

	resp, err := post(ctx, uploader.httpClient, u.String(), url.Values{
		"media_data": []string{encoded},
	})

	if err != nil {
		return err
	}

	return handleHttpResponse(resp, nil)
}

func (uploader uploader) finalize(ctx context.Context, mediaIdString string) (MediaUploadResponse, error) {
	var upResponse MediaUploadResponse

	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}
	q := u.Query()
	q.Set("media_id", mediaIdString)
	q.Set("command", "FINALIZE")
	u.RawQuery = q.Encode()

	resp, err := post(ctx, uploader.httpClient, u.String(), url.Values{})
	if err != nil {
		return upResponse, err
	}

	return upResponse, handleHttpResponse(resp, &upResponse)
}

func (uploader uploader) handleProcessing(ctx context.Context, origResponse MediaUploadResponse) (MediaUploadResponse, error) {
	if origResponse.ProcessingInfo == nil {
		return origResponse, nil
	}
	if origResponse.ProcessingInfo.State == ProcessingStateSucceeded {
		return origResponse, nil
	}
	if origResponse.ProcessingInfo.ProcessingError != nil {
		return origResponse, origResponse.ProcessingInfo.ProcessingError
	}

	time.Sleep(time.Duration(origResponse.ProcessingInfo.CheckAfterSecs) * time.Second)

	u, err := url.Parse(baseURL)
	if err != nil {
		return origResponse, err
	}
	q := u.Query()
	q.Set("media_id", origResponse.MediaIDString)
	q.Set("command", "STATUS")
	u.RawQuery = q.Encode()

	resp, err := get(ctx, uploader.httpClient, u.String())
	if err != nil {
		return origResponse, err
	}

	var upResponse MediaUploadResponse
	err = handleHttpResponse(resp, &upResponse)
	if err != nil {
		return upResponse, err
	}

	return uploader.handleProcessing(ctx, upResponse)
}
