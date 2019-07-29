package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	key, secret string
	BaseURL     *url.URL
	httpClient  *http.Client
}

var netClient *Client

func NewClient(key, secret string) *Client {
	url, err := url.Parse("https://api.github.com")
	if err != nil {
		panic(err)
	}
	cli := Client{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		BaseURL: url,
	}

	return &cli
}

func handleError(w http.ResponseWriter, err error) {
	resp := make(map[string]interface{})
	resp["error"] = ""
	status := 500

	switch errt := err.(type) {
	case ApiError:
		resp["error"] = errt.Error()
		status = errt.HTTPStatus
	default:
		resp["error"] = errt.Error()
	}
	marshalAndWrite(w, resp, status)
}

func marshalAndWrite(w http.ResponseWriter, resp map[string]interface{}, status int) {
	if enc, err := json.Marshal(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "InternalServerError")
		return
	} else {
		w.WriteHeader(status)
		w.Write(enc)
		return
	}
}

type ApiError struct {
	HTTPStatus int
	Err        error
}

func (ae ApiError) Error() string {
	return ae.Err.Error()
}

func init() {
	netClient = NewClient("123", "456")
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/repos":
		doGetRepos(w, r)
	case "/commits":
		doGetCommits(w, r)
	default:
		err := ApiError{
			404,
			fmt.Errorf("unknown method"),
		}
		handleError(w, err)
		//	 h.wrapperDoSomeJob(w, r)
	}

}

func doGetCommits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	user := r.FormValue("user")
	path := fmt.Sprintf("/users/%v/repos", user)
	err, body := netClient.runGet(path, map[string]string{})
	if err != nil {
		err = ApiError{500, err}
		handleError(w, err)
		return
	}
	var repos []Repo
	if err := json.Unmarshal(body, &repos); err != nil {
		err = ApiError{500, err}
		handleError(w, err)
		return
	}
	var commits []*Commit
	for _, repo := range repos {
		c := Commit{}
		path = fmt.Sprintf("repos/%v/%v/commits", user, repo)
		params := map[string]string{
			"author": user,
		}
		err, body := netClient.runGet(path, params)
		if err != nil {
			err = ApiError{500, err}
			handleError(w, err)
			return
		}
		if err := json.Unmarshal(body, &c); err != nil {
			err = ApiError{500, err}
			handleError(w, err)
			return
		}
		commits = append(commits, &c)
	}
	fmt.Printf("%#+v\n", commits)
	w.Write(body)

}

func doGetRepos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	user := r.FormValue("user")
	path := fmt.Sprintf("/users/%v/repos", user)
	err, body := netClient.runGet(path, map[string]string{})
	if err != nil {
		err = ApiError{
			500,
			err,
		}
		handleError(w, err)
		return
	}
	var repos []Repo
	json.Unmarshal(body, &repos)
	fmt.Printf("%#+v\n", repos)
	w.Write(body)

}

func (client *Client) runGet(path string, params map[string]string) (error, []byte) {
	url := client.BaseURL
	rel, err := url.Parse(path)
	if err != nil {
		return err, nil
	}

	q := rel.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	rel.RawQuery = q.Encode()
	req, err := http.NewRequest(http.MethodGet, rel.String(), nil)
	if err != nil {
		fmt.Println("error happend", err)
		return err, nil
	}
	fmt.Printf("[INFO] Do %v request", rel.String())
	resp, err := client.httpClient.Do(req)
	if err != nil {
		fmt.Println("error happend", err)
		return err, nil
	}

	defer resp.Body.Close() // важный пункт!
	respBody, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("[INFO] Got %v bytes", len(respBody))
	return nil, respBody
}
