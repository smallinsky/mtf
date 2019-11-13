package port

import (
	"context"
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

func TestGCStorage(t *testing.T) {
	t.Skip()
	port := NewGCStoragePort()
	r := mux.NewRouter()
	port.registerRuter(r)

	cst := httptest.NewServer(r)

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

	t.Run("ObjectGet", func(t *testing.T) {
		go func() {
			rcv, err := port.receive()
			if err != nil {
				t.Fatalf("failed to receive message")
			}
			msg, ok := rcv.(*StorageGetRequest)
			if !ok {
				t.Fatalf("wrong type")
			}
			t.Logf("Got %T: %+v", msg, msg)

			port.send(&StorageGetResponse{
				Content: []byte("to jest content"),
			})
		}()

		obj := sc.Bucket("bucket_name").Object("object_name")
		reader, err := obj.NewReader(ctx)
		if err != nil {
			t.Fatalf("got err during new reader call %v", err)
		}

		buff, err := ioutil.ReadAll(reader)
		if err != nil {
			t.Fatalf("failed to read: %v", err)
		}
		t.Logf("Bucket content: '%s'", string(buff))
	})

	t.Run("ObjectInsert", func(t *testing.T) {
		go func() {
			rcv, err := port.receive()
			if err != nil {
				t.Fatalf("failed to receive message")
			}
			msg, ok := rcv.(*StorageInsertRequest)
			if !ok {
				t.Fatalf("wrong type")
			}
			t.Logf("Got %T: %+v", msg, string(msg.Content))

			port.send(&StorageInsertResponse{})
		}()

		obj := sc.Bucket("bucket/a/b").Object("somefile.txt")
		w := obj.NewWriter(ctx)

		_, err = w.Write([]byte("write content"))
		if err != nil {
			t.Fatalf("got write error: %v", err)
		}
		if err := w.Close(); err != nil {
			t.Fatalf("got close error: %v", err)
		}
	})

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
