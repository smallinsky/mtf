package port

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/smallinsky/mtf/fake/fakegcs"
)

func NewGCStoragePort() *GCStorage {
	return &GCStorage{
		inEvent:  make(chan interface{}),
		outEvent: make(chan interface{}),
	}
}

func (s GCStorage) onObjectInsert(bo fakegcs.BucketObject, r io.Reader) error {
	buff, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatalf("[ERR] failed to read object content: %v", err)
	}
	req := &StorageInsertRequest{
		Bucket:  bo.Bucket,
		Object:  bo.Object,
		Content: buff,
	}

	select {
	case s.inEvent <- req:
	case <-time.Tick(time.Second * 3):
		log.Fatalf("gcs response not provided 1")
	}

	select {
	case <-s.outEvent:
		return nil
	case <-time.Tick(time.Second * 3):
		log.Fatalf("gcs response not provided 2")
		return nil
	}
}

func (s GCStorage) onObjectGet(bo fakegcs.BucketObject, w io.Writer) error {
	req := &StorageGetRequest{
		Bucket: bo.Bucket,
		Object: bo.Object,
	}

	select {
	case s.inEvent <- req:
	case <-time.Tick(time.Second * 3):
		log.Fatalf("gcs response not provided 3")
		return nil
	}

	select {
	case msg := <-s.outEvent:
		r, ok := msg.(*StorageGetResponse)
		if !ok {
			log.Fatalf("faield to receive event aa %T", msg)
		}

		_, err := io.Copy(w, bytes.NewReader(r.Content))
		if err != nil {
			log.Fatalf("faield to write content")
		}
	case <-time.Tick(time.Second * 3):
		log.Fatalf("gcs response not provided 4")
	}
	return nil
}

func (s GCStorage) registerRuter(r *mux.Router) {
	fgcs := &fakegcs.GCStorage{
		OnObjectInsert: s.onObjectInsert,
		OnObjectGet:    s.onObjectGet,
	}
	fgcs.AddMuxRoute(r)
}

type GCStorage struct {
	inEvent  chan interface{}
	outEvent chan interface{}
}

type StorageInsertRequest struct {
	Bucket  string
	Object  string
	Content []byte
}

type StorageInsertResponse struct {
}

type StorageGetRequest struct {
	Bucket string
	Object string
}

type StorageGetResponse struct {
	Content []byte
}

func (s *GCStorage) receive(opts ...Opt) (interface{}, error) {
	select {
	case <-time.Tick(time.Second * 3):
		return nil, errors.Errorf("failed to receive  message, deadline exeeded")
	case msg := <-s.inEvent:
		return msg, nil
	}
}

func (s *GCStorage) send(msg interface{}, opts ...PortOpt) error {
	select {
	case s.outEvent <- msg:
		return nil
	case <-time.Tick(time.Second * 3):
		return errors.Errorf("failed to receive  message, deadline exeeded")
	}
}

func (s *GCStorage) Send(ctx context.Context, i interface{}) error {
	return s.send(i)
}

func (s *GCStorage) Receive(ctx context.Context) (interface{}, error) {
	return s.receive()
}

func (p *GCStorage) Kind() Kind {
	return KIND_SERVER
}

func (p *GCStorage) Name() string {
	return "http_server"
}
