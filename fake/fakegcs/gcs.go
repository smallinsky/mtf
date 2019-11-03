package fakegcs

import (
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"

	"github.com/gorilla/mux"
)

type GCStorage struct {
	OnObjectInsert func(BucketObject, io.Reader) error
	OnObjectGet    func(BucketObject, io.Writer) error
}

func (f *GCStorage) handleInsert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	switch vars["uploadType"] {
	case "multipart":
		if err := f.handleMultipart(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "not supported content", http.StatusBadRequest)
	}
}

func (f *GCStorage) handleMultipart(rw http.ResponseWriter, r *http.Request) error {
	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return err
	}
	defer r.Body.Close()
	reader := multipart.NewReader(r.Body, params["boundary"])
	poartMeta, err := reader.NextPart()
	if err != nil {
		return err
	}
	defer poartMeta.Close()
	var obj meta
	if err := json.NewDecoder(poartMeta).Decode(&obj); err != nil {
		return err
	}
	partContent, err := reader.NextPart()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	defer partContent.Close()
	if f.OnObjectInsert != nil {
		f.OnObjectInsert(BucketObject{
			Object: obj.Object,
			Bucket: obj.Bucket,
		}, partContent)
	}
	return nil
}

type meta struct {
	Bucket string `json:"bucket"`
	Object string `json:"name"`
}

func (f *GCStorage) handleGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if f.OnObjectGet != nil {
		f.OnObjectGet(BucketObject{
			Bucket: vars["bucket"],
			Object: vars["object"],
		}, w)
	}
}

type tokenJSON struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func (f GCStorage) handleToken(w http.ResponseWriter, r *http.Request) {
	tj := tokenJSON{
		AccessToken:  "access-token",
		TokenType:    "Bearer",
		RefreshToken: "refresh-token",
		ExpiresIn:    3600,
	}

	if err := json.NewEncoder(w).Encode(&tj); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (f GCStorage) AddMuxRoute(r *mux.Router) *mux.Router {
	r.Path("/b/{bucket:.+}/o").Queries("uploadType", "{uploadType}").Methods(http.MethodPost).HandlerFunc(f.handleInsert)
	r.Path("/{bucket:.+}/{object}").Methods(http.MethodGet).HandlerFunc(f.handleGet)
	r.Path("/token").Methods(http.MethodPost).HandlerFunc(f.handleToken)

	r.NotFoundHandler = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		panic("not found")
	})
	return r
}

type BucketObject struct {
	Bucket string
	Object string
}
