package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type metadataResponse struct {
	Result []string `json:"result"`
}

type functionResponse struct {
	Result interface{} `json:"result"`
}

func Handler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri := strings.Split(r.RequestURI, "/")
		corsHandler(w, r)
		if strings.HasPrefix(r.RequestURI, "/1.1/functions/") {
			if strings.Compare(r.RequestURI, "/1.1/functions/_ops/metadatas") == 0 {
				metadataHandler(w, r)
			} else {
				if functions[uri[3]] != nil {
					functionHandler(w, r, uri[3])
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}
		} else if strings.HasPrefix(r.RequestURI, "/1.1/call/") {
			if functions[uri[3]] != nil {
				// TODO: RPC Calling
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}
	})
}

func corsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("origin"))

	if r.Method == "OPTIONS" {
		w.Header().Add("Access-Control-Max-Age", "86400")
		w.Header().Add("Access-Control-Allow-Methods", "HEAD, GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Add("Access-Control-Allow-Headers", `Content-Type,X-AVOSCloud-Application-Id,X-AVOSCloud-Application-Key,X-AVOSCloud-Application-Production,X-AVOSCloud-Client-Version,X-AVOSCloud-Request-Sign,X-AVOSCloud-Session-Token,X-AVOSCloud-Super-Key,X-LC-Hook-Key,X-LC-Id,X-LC-Key,X-LC-Prod,X-LC-Session,X-LC-Sign,X-LC-UA,X-Requested-With,X-Uluru-Application-Id,X-Uluru-Application-Key,X-Uluru-Application-Production,X-Uluru-Client-Version,X-Uluru-Session-Token`)
	}
}

func metadataHandler(w http.ResponseWriter, r *http.Request) {
	meta, err := generateMetadata()
	if err != nil {
		errorResponse(w, r, err)
		return
	}
	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	fmt.Fprintln(w, string(meta))
}

func functionHandler(w http.ResponseWriter, r *http.Request, name string) {
	request, err := constructRequest(r, name)
	if err != nil {
		errorResponse(w, r, err)
		return
	}

	resp, err := functions[name].call(request)
	if err != nil {
		errorResponse(w, r, err)
		return
	}

	funcResp := functionResponse{
		Result: resp,
	}
	respJSON, err := json.Marshal(funcResp)
	if err != nil {
		errorResponse(w, r, err)
		return
	}

	w.Header().Add("Contetn-Type", "application/json; charset=UTF-8")
	fmt.Fprintln(w, string(respJSON))
}

func unmarshalBody(r *http.Request) (interface{}, error) {
	body := make(map[string]interface{})
	err := json.NewDecoder(r.Body).Decode(&body)

	if err == io.EOF {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	return body, nil
}

func constructRequest(r *http.Request, name string) (*Request, error) {
	request := new(Request)
	request.Meta = map[string]string{
		"remoteAddr": r.RemoteAddr,
	}
	sessionToken := r.Header.Get("X-LC-Session")
	if !functions[name].NotFetchUser && sessionToken != "" {
		user, err := client.Users.Become(sessionToken)
		if err != nil {
			return nil, err
		}
		request.CurrentUser = user
		request.SessionToken = sessionToken
	}

	if r.Body == nil {
		request.Params = nil
		return request, nil
	}
	params, err := unmarshalBody(r)
	if err != nil {
		return nil, err
	}
	request.Params = params

	return request, nil
}

func errorResponse(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Add("Contetn-Type", "application/json; charset=UTF-8")
	switch err.(type) {
	case *functionError:
		fmt.Fprintln(w, err.Error())
		w.WriteHeader(http.StatusBadRequest)
	default:
		fmt.Fprintln(w, Error(err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func generateMetadata() ([]byte, error) {
	meta := metadataResponse{
		Result: []string{},
	}

	for k := range functions {
		meta.Result = append(meta.Result, k)
	}
	return json.Marshal(meta)
}