package registry

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"gorm.io/gorm"

	"github.com/Haevnen/audit-logging-api/internal/auth"
	"github.com/Haevnen/audit-logging-api/internal/interactor"
	"github.com/Haevnen/audit-logging-api/internal/repository"
	"github.com/Haevnen/audit-logging-api/internal/service"
	"github.com/Haevnen/audit-logging-api/internal/usecase/log"
	"github.com/Haevnen/audit-logging-api/internal/usecase/tenant"
)

type Registry struct {
	db              *gorm.DB
	sqsClient       *sqs.Client
	s3Client        *s3.Client
	key             string
	archiveQueueURL string
	cleanUpQueueURL string
	s3BucketName    string
}

func NewRegistry(db *gorm.DB, key string, sqsClient *sqs.Client, s3Client *s3.Client, archiveQueueURL string, cleanUpQueueURL string, s3BucketName string) *Registry {
	return &Registry{
		db:              db,
		key:             key,
		sqsClient:       sqsClient,
		archiveQueueURL: archiveQueueURL,
		cleanUpQueueURL: cleanUpQueueURL,
		s3Client:        s3Client,
		s3BucketName:    s3BucketName,
	}
}

func (r *Registry) TenantRepository() repository.TenantRepository {
	return repository.NewTenantRepository(r.db)
}

func (r *Registry) AsyncTaskRepository() repository.AsyncTaskRepository {
	return repository.NewAsyncTaskRepository(r.db)
}

func (r *Registry) LogRepository() repository.LogRepository {
	return repository.NewLogRepository(r.db)
}

func (r *Registry) CreateTenantUseCase() *tenant.CreateTenantUseCase {
	return tenant.NewCreateTenantUseCase(r.TenantRepository())

}

func (r *Registry) ListTenantsUseCase() *tenant.ListTenantsUseCase {
	return tenant.NewListTenantsUseCase(r.TenantRepository())
}

func (r *Registry) CreateLogUseCase() *log.CreateLogUseCase {
	return log.NewCreateLogUseCase(r.LogRepository(), r.TxManager())
}

func (r *Registry) GetLogUseCase() *log.GetLogUseCase {
	return log.NewGetLogUseCase(r.LogRepository())
}

func (r *Registry) DeleteLogUseCase() *log.DeleteLogUseCase {
	return log.NewDeleteLogUseCase(r.AsyncTaskRepository(), r.QueuePublisher(), r.TxManager())
}

func (r *Registry) QueuePublisher() service.SQSPublisher {
	return service.NewSQSPublisherImpl(r.sqsClient, r.archiveQueueURL, r.cleanUpQueueURL)
}

func (r *Registry) S3Publisher() service.S3Publisher {
	return service.NewS3PublisherImpl(r.s3Client, r.s3BucketName)
}

func (r *Registry) Manager() *auth.Manager {
	return auth.NewManager(r.key)
}

func (r *Registry) TxManager() *interactor.TxManager {
	return interactor.NewTxManager(r.db)
}
