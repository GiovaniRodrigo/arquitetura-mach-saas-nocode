//go:build integration

// Requer um MinIO acessível. Executar:
//
//	S3_ENDPOINT=localhost:9010 go test -tags integration ./services/export/internal/storage/...
package storage

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func endpoint() string {
	if v := os.Getenv("S3_ENDPOINT"); v != "" {
		return v
	}
	return "localhost:9010"
}

func TestUpload_Presign_Download(t *testing.T) {
	ctx := context.Background()
	s, err := New(ctx, Config{
		Endpoint:   endpoint(),
		AccessKey:  "mach",
		SecretKey:  "machsecret",
		Bucket:     "exports-test",
		PresignTTL: 5 * time.Minute,
	})
	if err != nil {
		t.Skipf("MinIO indisponível (%v)", err)
	}

	conteudo := "{\"origem\":\"design\"}\n{\"origem\":\"regras\"}\n"
	objeto := "itg/export.ndjson"
	if err := s.Upload(ctx, objeto, strings.NewReader(conteudo)); err != nil {
		t.Fatalf("upload: %v", err)
	}

	url, expira, err := s.PresignedGet(ctx, objeto)
	if err != nil {
		t.Fatalf("presign: %v", err)
	}
	if !expira.After(time.Now()) {
		t.Fatalf("expiração deveria ser futura; got %v", expira)
	}

	resp, err := http.Get(url) //nolint:gosec // URL assinado de teste
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("download status %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != conteudo {
		t.Fatalf("conteúdo divergente: %q", body)
	}
}
