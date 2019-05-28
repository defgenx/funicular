package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"log"
	"os"
)

type AWSManager struct {
	config       *aws.Config
	disconnected chan bool
	closed       bool
	S3Manager    *S3Manager
	log          *log.Logger
}

func NewAWSManager(maxRetries uint8) *AWSManager {
	config := &aws.Config {
		MaxRetries: aws.Int(int(maxRetries)),
	}
	return &AWSManager{
		config: config,
		S3Manager: NewS3Manager(config),
		log: log.New(os.Stdout, "AWSManager", log.Ldate|log.Ltime|log.Lshortfile),
	}
}


type S3Manager struct {
	session    *session.Session
	Client     *s3.S3
	S3Conns     []*S3Wrapper
}

func NewS3Manager(config *aws.Config) *S3Manager {
	sess := session.Must(session.NewSession(config))
	s3Client := s3.New(sess)
	return &S3Manager{
		session: sess,
		Client: s3Client,
		S3Conns: make([]*S3Wrapper, 0),
	}
}

func (s3m *S3Manager) AddS3BucketManager() *S3Wrapper {
	s3Wrapper := NewS3Wrapper(s3m.session)
	s3m.S3Conns = append(s3m.S3Conns, s3Wrapper)
	return s3Wrapper
}

// S3 Adapter
type S3Wrapper struct {
	Uploader   *s3manager.Uploader
	Downloader *s3manager.Downloader
}

func NewS3Wrapper(s3Session *session.Session) *S3Wrapper {
	uploader := s3manager.NewUploader(s3Session)
	downloader := s3manager.NewDownloader(s3Session)
	return &S3Wrapper{
		Uploader: uploader,
		Downloader: downloader,
	}
}