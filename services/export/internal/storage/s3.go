// Package storage carrega o artefato de exportação no S3/MinIO e emite um
// Presigned URL de expiração curta para download direto pelo utilizador (RF05,
// critério 7) — sem expor credenciais nem passar os bytes pelo Gateway.
package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Armazenamento é o subconjunto usado pelo servidor (facilita testes com fake).
type Armazenamento interface {
	Upload(ctx context.Context, objeto string, r io.Reader) error
	PresignedGet(ctx context.Context, objeto string) (url string, expiraEm time.Time, err error)
}

// S3 implementa Armazenamento sobre um cliente MinIO (compatível com S3).
type S3 struct {
	client     *minio.Client
	bucket     string
	presignTTL time.Duration
}

// Config parametriza a conexão ao S3/MinIO.
type Config struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	Bucket     string
	UseSSL     bool
	PresignTTL time.Duration
}

// New cria o cliente e garante a existência do bucket.
func New(ctx context.Context, cfg Config) (*S3, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("storage: cliente: %w", err)
	}

	existe, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("storage: bucket exists: %w", err)
	}
	if !existe {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("storage: criar bucket: %w", err)
		}
	}

	ttl := cfg.PresignTTL
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &S3{client: client, bucket: cfg.Bucket, presignTTL: ttl}, nil
}

// Upload envia o objeto por streaming (tamanho desconhecido → multipart), mantendo
// a RAM constante mesmo para exportações grandes (RNF01).
func (s *S3) Upload(ctx context.Context, objeto string, r io.Reader) error {
	_, err := s.client.PutObject(ctx, s.bucket, objeto, r, -1, minio.PutObjectOptions{
		ContentType: "application/x-ndjson",
	})
	if err != nil {
		return fmt.Errorf("storage: upload %s: %w", objeto, err)
	}
	return nil
}

// PresignedGet devolve um URL de download com expiração curta e o instante de
// expiração (para gravar no Job).
func (s *S3) PresignedGet(ctx context.Context, objeto string) (string, time.Time, error) {
	u, err := s.client.PresignedGetObject(ctx, s.bucket, objeto, s.presignTTL, nil)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("storage: presign %s: %w", objeto, err)
	}
	return u.String(), time.Now().Add(s.presignTTL), nil
}
