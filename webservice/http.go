package webservice

import (
	"crypto/tls"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/flexiant/concerto/config"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const contentDispositionRegex = "filename=\\\"([^\\\"]*){1}\\\""

type Webservice struct {
	config *config.Config
	client *http.Client
}

func NewWebService() (*Webservice, error) {
	config, err := config.ConcertoServerConfiguration()
	if err != nil {
		return nil, err
	}

	client, err := httpClient(config)
	if err != nil {
		return nil, err
	}

	return &Webservice{config, client}, nil
}

func httpClient(config *config.Config) (*http.Client, error) {

	// Loads Clients Certificates and creates and 509KeyPair
	cert, err := tls.LoadX509KeyPair(config.Certificate.Cert, config.Certificate.Key)
	if err != nil {
		return nil, err
	}

	// Creates a client with specific transport configurations
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}

	return client, nil
}

func (w *Webservice) Post(endpoint string, json []byte) error {
	log.Debugf("Connecting: %s%s", w.config.ApiEndpoint, endpoint)
	output := strings.NewReader(string(json))
	response, err := w.client.Post(w.config.ApiEndpoint+endpoint, "application/json", output)

	log.Debugf("Posting: %s", output)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	log.Debugf("Status code: %s", response.Status)
	return nil
}

func (w *Webservice) Get(endpoint string) ([]byte, error) {

	log.Debugf("Connecting: %s%s", w.config.ApiEndpoint, endpoint)
	response, err := w.client.Get(w.config.ApiEndpoint + endpoint)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	log.Debugf("Status code: %s", response.Status)
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (w *Webservice) GetFile(endpoint string, directoryPath string) (string, error) {

	log.Debugf("Connecting: %s%s", w.config.ApiEndpoint, endpoint)
	response, err := w.client.Get(w.config.ApiEndpoint + endpoint)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	log.Debugf("Status code: %s", response.Status)

	r, err := regexp.Compile(contentDispositionRegex)
	if err != nil {
		return "", err
	}

	fileName := r.FindStringSubmatch(response.Header.Get("Content-Disposition"))[1]
	if err != nil {
		return "", err
	}
	realFileName := fmt.Sprintf("%s/%s", directoryPath, fileName)

	output, err := os.Create(realFileName)
	if err != nil {
		return "", err
	}
	defer output.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		return "", err
	}

	log.Debugf("%#v bytes downloaded", n)
	return realFileName, nil
}