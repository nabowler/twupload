package twupload

import (
	"context"

	"github.com/dghubble/oauth1"
)

type (
	Uploader interface {
		Upload(ctx context.Context, data []byte) (MediaUploadResponse, error)
	}

	MediaUploadResponse struct {
		MediaID          int64           `json:"media_id"`
		MediaIDString    string          `json:"media_id_string"`
		MediaKey         string          `json:"media_key"`
		Size             int64           `json:"size"`
		ExpiresAfterSecs int64           `json:"expires_after_secs"`
		Image            *Image          `json:"image"`
		Video            *Video          `json:"video"`
		ProcessingInfo   *ProcessingInfo `json:"processing_info"`
	}

	Image struct {
		ImageType string `json:"image_type"`
		W         int64  `json:"w"`
		H         int64  `json:"h"`
	}

	Video struct {
		VideoType string `json:"video_type"`
	}

	ProcessingInfo struct {
		State           ProcessingState  `json:"state"`
		CheckAfterSecs  int              `json:"check_after_secs"`
		ProgressPercent int              `json:"progress_percent"`
		ProcessingError *ProcessingError `json:"error"`
	}

	ProcessingError struct {
		Code    int    `json:"code"`
		Name    string `json:"name"`
		Message string `json:"message"`
	}

	MediaCategory string

	ProcessingState string

	UploaderConfig struct {
		TwitterAPIKey       string
		TwitterAPISecret    string
		TwitterAccessToken  string
		TwitterAccessSecret string

		ChunkSize int
	}

	sentinelError string
)

const (
	MediaCategoryImage        MediaCategory = "tweet_image"
	MediaCategoryGif          MediaCategory = "tweet_gif"
	MediaCategoryTweetVideo   MediaCategory = "tweet_video"
	MediaCategoryAmplifyVideo MediaCategory = "amplify_video"

	ProcessingStatePending    ProcessingState = "pending"
	ProcessingStateInProgress ProcessingState = "in_progress"
	ProcessingStateSucceeded  ProcessingState = "succeeded"
	PorcessingStateFailed     ProcessingState = "failed"

	defaultChunkSize = 512 * 1024

	MaxChunkSize = 5 * 1024 * 1024

	baseURL = "https://upload.twitter.com/1.1/media/upload.json"
)

func (uc UploaderConfig) NewUploader() (Uploader, error) {
	config := oauth1.NewConfig(uc.TwitterAPIKey, uc.TwitterAPISecret)
	token := oauth1.NewToken(uc.TwitterAccessToken, uc.TwitterAccessSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	chunkSize := defaultChunkSize
	if uc.ChunkSize > MaxChunkSize {
		chunkSize = MaxChunkSize
	} else if uc.ChunkSize > 0 {
		chunkSize = uc.ChunkSize
	}

	return uploader{
		httpClient: httpClient,
		chunkSize:  chunkSize,
	}, nil
}

func (err sentinelError) Error() string {
	return string(err)
}

func (err ProcessingError) Error() string {
	return err.Message
}
