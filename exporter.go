package docker_hub_exporter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Namespace of the prometheus metrics
const Namespace = "docker_hub_image"

var (
	dockerHubImageLastUpdated = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "last_updated"),
		"docker_hub_exporter: Docker Image Last Updated",
		[]string{"image", "user"}, nil,
	)
	dockerHubImagePullsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "pulls_total"),
		"docker_hub_exporter: Docker Image Pulls Total.",
		[]string{"image", "user"}, nil,
	)
	dockerHubImageStars = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "stars"),
		"docker_hub_exporter: Docker Image Stars.",
		[]string{"image", "user"}, nil,
	)
	dockerHubImageIsAutomated = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "is_automated"),
		"docker_hub_exporter: Docker Image Is Automated.",
		[]string{"image", "user"}, nil,
	)
)

// Exporter is used to store Metrics data
type Exporter struct {
	timeout           time.Duration
	baseURL           string
	organisations     []string
	images            []string
	logger            *log.Logger
	connectionRetries int
}

type OrganisationResult struct {
	Count    int           `json:"count"`
	Next     string        `json:"next"`
	Previous string        `json:"previous"`
	Results  []ImageResult `json:"results"`
}

type ImageResult struct {
	Name        string    `json:"name"`
	User        string    `json:"user"`
	StarCount   float64   `json:"star_count"`
	IsAutomated bool      `json:"is_automated"`
	PullCount   float64   `json:"pull_count"`
	LastUpdated time.Time `json:"last_updated"`
}

// New creates a new Exporter and returns it
func New(organisations, images []string, connectionRetries int, opts ...Option) *Exporter {
	e := &Exporter{
		timeout:       	   time.Second * 5,
		baseURL:           "https://hub.docker.com/v2/repositories/",
		organisations:     organisations,
		images:            images,
		logger:            log.New(ioutil.Discard, "docker_hub_exporter: ", log.LstdFlags),
		connectionRetries: connectionRetries,
	}

	for _, opt := range opts {
		opt(e)
	}

	e.logger.Printf("Organisations to monitor: %v", e.organisations)
	e.logger.Printf("Images to monitor: %v", e.images)

	return e
}

type Option func(*Exporter)

func WithLogger(logger *log.Logger) Option {
	return func(e *Exporter) { e.logger = logger }
}

func WithBaseURL(baseURL string) Option {
	return func(e *Exporter) { e.baseURL = baseURL }
}

func WithTimeout(timeout time.Duration) Option {
	return func(e *Exporter) { e.timeout = timeout }
}

// Describe implements the prometheus.Collector interface.
func (e Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- dockerHubImageLastUpdated
	ch <- dockerHubImagePullsTotal
	ch <- dockerHubImageStars
	ch <- dockerHubImageIsAutomated
}

// Collect implements the prometheus.Collector interface.
func (e Exporter) Collect(ch chan<- prometheus.Metric) {
	e.logger.Println("Collecting metrics")

	e.collectMetrics(ch)
}

func (e Exporter) collectMetrics(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(e.organisations) + len(e.images))

	for _, url := range e.organisations {
		go func(url string) {
			if url != "" {
				response, err := e.getOrgMetrics(fmt.Sprintf("%s%s", e.baseURL, url))

				if err != nil {
					e.logger.Println("error ", err)
					wg.Done()
					return
				}

				for _, orgResp := range response {
					for _, result := range orgResp.Results {
						e.processImageResult(result, ch)
					}
				}
			}

			wg.Done()
		}(strings.TrimSpace(url))
	}

	for _, url := range e.images {
		go func(url string) {
			if url != "" {
				response, err := e.getImageMetrics(fmt.Sprintf("%s%s", e.baseURL, url))

				if err != nil {
					e.logger.Println("error ", err)
					wg.Done()
					return
				}

				e.processImageResult(response, ch)
			}

			wg.Done()
		}(strings.TrimSpace(url))
	}

	wg.Wait()
}

func (e Exporter) processImageResult(result ImageResult, ch chan<- prometheus.Metric) {
	if result.Name != "" && result.User != "" {
		var isAutomated float64
		if result.IsAutomated {
			isAutomated = float64(1)
		} else {
			isAutomated = float64(0)
		}

		lastUpdated := float64(result.LastUpdated.UnixNano()) / 1e9

		ch <- prometheus.MustNewConstMetric(dockerHubImageStars, prometheus.GaugeValue, result.StarCount, result.Name, result.User)
		ch <- prometheus.MustNewConstMetric(dockerHubImageIsAutomated, prometheus.GaugeValue, isAutomated, result.Name, result.User)
		ch <- prometheus.MustNewConstMetric(dockerHubImagePullsTotal, prometheus.CounterValue, result.PullCount, result.Name, result.User)
		ch <- prometheus.MustNewConstMetric(dockerHubImageLastUpdated, prometheus.GaugeValue, lastUpdated, result.Name, result.User)
	}
}

func (e Exporter) getImageMetrics(url string) (ImageResult, error) {
	imageResult := ImageResult{}

	body, err := e.getResponse(url)
	if err != nil {
		return ImageResult{}, err
	}

	err = json.Unmarshal(body, &imageResult)
	if err != nil {
		return ImageResult{}, fmt.Errorf("Error unmarshalling response: %v", err)
	}

	return imageResult, nil
}

func (e Exporter) getOrgMetrics(url string) ([]OrganisationResult, error) {
	orgResult := OrganisationResult{}

	body, err := e.getResponse(url)
	if err != nil {
		return []OrganisationResult{}, err
	}

	err = json.Unmarshal(body, &orgResult)
	if err != nil {
		return []OrganisationResult{}, fmt.Errorf("Error unmarshalling response: %v", err)
	}

	if orgResult.Count == 0 || len(orgResult.Results) == 0 {
		return []OrganisationResult{}, fmt.Errorf("No images found for url: %s", url)
	}

	if orgResult.Next != "" {
		orgResult1, err := e.getOrgMetrics(orgResult.Next)
		if err != nil {
			return []OrganisationResult{}, err
		}

		return append([]OrganisationResult{orgResult}, orgResult1...), nil
	}

	return []OrganisationResult{orgResult}, nil
}

// getResponse collects an individual http.response and returns a *Response
func (e Exporter) getResponse(url string) ([]byte, error) {

	e.logger.Printf("Fetching %s \n", url)

	resp, err := e.getHTTPResponse(url) // do this earlier

	if err != nil {
		return nil, fmt.Errorf("Error converting body to byte array: %v", err)
	}

	// Read the body to a byte array so it can be used elsewhere
	body, err := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("Error converting body to byte array: %v", err)
	}

	return body, nil
}

// getHTTPResponse handles the http client creation, token setting and returns the *http.response
func (e Exporter) getHTTPResponse(url string) (*http.Response, error) {
	
	client := &http.Client{
		Timeout: e.timeout,
	}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, fmt.Errorf("Failed to create http request: %v", err)
	}

	var retries = e.connectionRetries
	for retries > 0 {
		resp, err := client.Do(req)
		if err != nil {
			retries -= 1

			if retries == 0 {
				return nil, err
			} else {
				e.logger.Printf("Retrying HTTP request %s", url)
			}
		} else {
			return resp, nil
		}
	}
	return nil, nil
}
