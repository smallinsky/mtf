package fakegcs

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

func TestStorageInsert(t *testing.T) {
	sync := make(chan struct{})
	fakeStorage := &GCStorage{
		OnObjectInsert: func(o BucketObject, r io.Reader) error {
			close(sync)
			return nil
		},
	}
	r := mux.NewRouter()
	StorageHost = "{[0-9]:.+}"
	r = fakeStorage.AddMuxRoute(r)
	cst := httptest.NewServer(r)
	defer cst.Close()

	readHostEnv := strings.Replace(cst.URL, "http://", "", -1)
	if err := os.Setenv("STORAGE_EMULATOR_HOST", readHostEnv); err != nil {
		t.Fatalf("failed to set readHost env: %v", err)
	}

	hc := &http.Client{
		Transport: &oauth2.Transport{
			Source: new(tokenSupplier),
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := []option.ClientOption{option.WithHTTPClient(hc), option.WithEndpoint(cst.URL)}

	sc, err := storage.NewClient(ctx, opts...)
	if err != nil {
		t.Fatalf("Failed to create storage client: %v", err)
	}
	defer sc.Close()

	obj := sc.Bucket("bucket/a/b").Object("somefile.txt")
	w := obj.NewWriter(ctx)

	_, err = w.Write([]byte(fmt.Sprintf("%s", strings.Repeat("Z", 64))))
	if err != nil {
		t.Fatalf("got write error: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("got close error: %v", err)
	}
	<-sync
}

func TestHandleGet(t *testing.T) {
	fakeStorage := &GCStorage{
		OnObjectGet: func(bo BucketObject, w io.Writer) error {
			maxFileSize := 1 << 20 // 1MB
			chunkSize := maxFileSize / 4
			for i := 0; i < maxFileSize; {
				now := time.Now().Format(time.RFC3339Nano)
				n, _ := fmt.Fprintf(w, "%s%s", now, strings.Repeat("z", chunkSize))
				i += n
			}
			return nil
		},
	}
	muxRouter := mux.NewRouter()
	StorageHost = "{[0-9]:.+}"
	muxRouter = fakeStorage.AddMuxRoute(muxRouter)
	cst := httptest.NewServer(muxRouter)
	defer cst.Close()

	readHostEnv := strings.Replace(cst.URL, "http://", "", -1)
	if err := os.Setenv("STORAGE_EMULATOR_HOST", readHostEnv); err != nil {
		t.Fatalf("failed to set readHost env: %v", err)
	}

	hc := &http.Client{
		Transport: &oauth2.Transport{
			Source: new(tokenSupplier),
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := []option.ClientOption{option.WithHTTPClient(hc), option.WithEndpoint(cst.URL)}

	sc, err := storage.NewClient(ctx, opts...)
	if err != nil {
		t.Fatalf("Failed to create storage client: %v", err)
	}
	defer sc.Close()

	obj := sc.Bucket("bucket_name").Object("object_name")
	r, err := obj.NewReader(ctx)
	if err != nil {
		t.Fatalf("got err during new reader call %v", err)
	}

	buff, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	t.Logf("\nReceived %v bytes\n", len(buff))

	if err := r.Close(); err != nil {
		t.Fatalf("got reader close error: %v", err)
	}
}

type tokenSupplier int

func (ts *tokenSupplier) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken:  "access-token",
		TokenType:    "Bearer",
		RefreshToken: "refresh-token",
		Expiry:       time.Now().Add(time.Hour),
	}, nil
}
