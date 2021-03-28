# twupload

Unofficial, and incomplete, implementation of uploading to twitter.

See https://developer.twitter.com/en/docs/twitter-api/v1/media/upload-media/api-reference/post-media-upload

## Usage

```go
import "github.com/nabowler/twupload"


uploader, err := twupload.UploaderConfig{
    TwitterAPIKey: os.Getenv("TWITTER_API_KEY"),
    TwitterAPISecret: os.Getenv("TWITTER_API_SECRET"),
    TwitterAccessToken: os.Getenv("TWITTER_ACCESS_TOKEN"),
    TwitterAccessSecret: os.Getenv("TWITTER_ACCESS_SECRET"),
    
    ChunkSize: 1024 * 1024, // Max 5MB
}.NewUploader()

if err != nil {
    // handle error
}

data := // retrieve data

// set your timeout based on your needs
ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
defer cancel()

response, err := uploader.Upload(ctx, data)

if err != nil {
    // handle err
}

// use the response data in order to create a tweet with the media
```