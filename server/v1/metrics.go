/*
Copyright 2024 BlackRock, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	pb "github.com/tinymultiverse/tinyapp/pkg/server/api/v1/proto"
	"github.com/tinymultiverse/tinyapp/util"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
)

// PrometheusSecret is a struct that holds the Prometheus secrets.
type prometheusSecret struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type PrometheusResults struct {
	Data DataResults `json:"data"`
}

type DataResults struct {
	Result []map[string]interface{} `json:"result"`
}

type UserNameData struct {
	Username string `json:"username"`
}

func (s *Server) GetTinyAppLogs(ctx context.Context, in *pb.GetTinyAppLogsRequest) (*pb.GetTinyAppLogsResponse, error) {
	logger := zap.S().With("appId", in.AppId)
	logger.Info("Received request to get app logs")

	appId := in.AppId

	podsList, err := s.k8sClient.CoreV1().Pods(s.env.TinyAppNamespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", util.K8sNameLabel, appId),
	})
	if err != nil {
		logger.Errorf("failed to get pods list: %s", err)
		return nil, err
	}

	if len(podsList.Items) == 0 {
		logger.Errorf("no pods found for app id: %s", appId)
		return nil, errors.New("no pods found for app id")
	}

	// Assume there is only one pod for the app
	podName := podsList.Items[0].Name
	logsRequest := s.k8sClient.CoreV1().Pods(s.env.TinyAppNamespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: "app",
	})

	logs, err := readLogs(logsRequest)
	if err != nil {
		logger.Errorf("failed to read logs: %s", err)
		return nil, err
	}

	logger.Info("Successfully retrieved app logs")

	return &pb.GetTinyAppLogsResponse{
		Logs: logs,
	}, nil
}

// readLogs reads logs as string from logsRequest.
func readLogs(logsRequest *restclient.Request) (string, error) {
	if logsRequest == nil {
		return "", nil
	}

	reader, err := logsRequest.Stream(context.Background())
	if err != nil {
		return "", errors.WithMessage(err, "failed to stream logs")
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(reader); err != nil {
		return "", errors.WithMessage(err, "failed to read logs from reader")
	}

	return buf.String(), nil
}

func (s *Server) GetTinyAppAccessMetrics(ctx context.Context, in *pb.GetTinyAppAccessMetricsRequest) (*pb.GetTinyAppAccessMetricsResponse, error) {
	logger := zap.S()
	logger.Info("Received request to get app access metrics")

	appId := in.AppId
	timeRange := in.TimePeriod
	podNameRegex := appId + "-.*"

	userNameQuery := fmt.Sprintf("sum by(username) (increase(username_counter{kubernetes_pod_name=~\"%s\"}[%s]))", podNameRegex, timeRange)
	accessMap, err := getAppAccessCount(s.promSecret, s.env.PrometheusUrl, userNameQuery)
	if err != nil {
		logger.Errorf("unable to decode query response from prometheus: %s", err)
		return nil, err
	}

	getTinyAppMetricsResponse := &pb.GetTinyAppAccessMetricsResponse{
		// Once integrated with OAuth, we will have more granular "per user" access metric.
		// For now, there will always be one "user".
		NumberOfAccess: accessMap[util.AnyUserName],
	}

	logger.Info("Successfully retrieved app access metrics")

	return getTinyAppMetricsResponse, nil
}

func (s *Server) GetTinyAppUsageMetrics(ctx context.Context, in *pb.GetTinyAppUsageMetricsRequest) (*pb.GetTinyAppUsageMetricsResponse, error) {
	logger := zap.S()
	logger.Info("Received request to get app usage metrics")

	timeRange := in.TimePeriod
	podNameRegex := in.AppId + "-.*"

	// queries to get cpu and memory usage and limits
	cpuUsageQuery := fmt.Sprintf("sum(rate(container_cpu_usage_seconds_total{container=\"app\", pod=~\"%s\"}[%s]))", podNameRegex, timeRange)
	cpuLimitsQuery := fmt.Sprintf("sum(kube_pod_container_resource_limits{container=\"app\", resource=\"cpu\", pod=~\"%s\"})", podNameRegex)
	memoryUsageQuery := fmt.Sprintf("sum(container_memory_working_set_bytes{container=\"app\", pod=~\"%s\"})", podNameRegex)
	memoryLimitsQuery := fmt.Sprintf("sum(kube_pod_container_resource_limits{container=\"app\", resource=\"memory\", pod=~\"%s\"})", podNameRegex)

	queryList := [4]string{cpuUsageQuery, cpuLimitsQuery, memoryUsageQuery, memoryLimitsQuery}
	// array to store the results from each query
	var usageMetricsResult [4]float64

	for i, query := range queryList {
		// call the decode method, get the result
		// add result to resulting object
		usageMetric, err := getUsageMetric(s.promSecret, s.env.PrometheusUrl, query)
		if err != nil {
			logger.Errorf("unable to decode query response from prometheus: %s", err)
			usageMetricsResult[i] = 0
		} else {
			usageMetricsResult[i] = *usageMetric
		}
	}

	getTinyAppMetricsResponse := &pb.GetTinyAppUsageMetricsResponse{
		CpuUsage:          usageMetricsResult[0],
		CpuLimit:          usageMetricsResult[1],
		MemoryUsage:       usageMetricsResult[2] / math.Pow(10, 9),
		MemoryLimit:       usageMetricsResult[3] / math.Pow(10, 9),
		PercentCpuUsed:    usageMetricsResult[0] / usageMetricsResult[1],
		PercentMemoryUsed: usageMetricsResult[2] / usageMetricsResult[3],
	}

	logger.Info("Successfully retrieved app usage metrics")

	return getTinyAppMetricsResponse, nil
}

// getAppAccessCount queries prometheus for app access count per user.
// Returns a list of usernames and the number of times each user accessed the app.
func getAppAccessCount(promSecret prometheusSecret, prometheusUrl, query string) (map[string]int32, error) {
	var resultsArray, err = queryAndDecode(promSecret, prometheusUrl, query)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to make query to prometheus")
	}

	if len(resultsArray) == 0 {
		return nil, errors.New("no results found for the given query")
	}

	userNameCountMap := make(map[string]int32)
	for _, result := range resultsArray {
		metricJson, err := json.Marshal(result["metric"])
		if err != nil {
			return nil, errors.WithMessage(err, "failed to marshal metric data")
		}

		var userNameData UserNameData
		if err := json.Unmarshal(metricJson, &userNameData); err != nil {
			return nil, errors.WithMessage(err, "failed to unmarshal username data")
		}

		count, err := extractFloatValue(result)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to parse username count to float")
		}

		userNameCountMap[userNameData.Username] = int32(count)
	}

	zap.S().Debug("userNameCountMap: ", userNameCountMap)
	return userNameCountMap, nil
}

// getUsageMetric queries prometheus for usage metrics (cpu or memory).
// Returns result in float64.
func getUsageMetric(promSecret prometheusSecret, prometheusUrl, query string) (*float64, error) {
	var resultsArray, err = queryAndDecode(promSecret, prometheusUrl, query)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to make query to prometheus")
	}

	if len(resultsArray) == 0 {
		return nil, errors.New("no results found for the given query")
	}

	metricFloat, err := extractFloatValue(resultsArray[0])
	if err != nil {
		return nil, errors.WithMessage(err, "failed to extract metric value")
	}

	return &metricFloat, nil
}

// queryAndDecode makes given query and returns the decoded response.
func queryAndDecode(promSecret prometheusSecret, prometheusUrl, query string) ([]map[string]interface{}, error) {
	var prometheusResults PrometheusResults
	var resultsArray []map[string]interface{}

	if body, err := getQueryResults(promSecret, prometheusUrl, query); err != nil {
		return nil, fmt.Errorf("error reading body: %s", err)
	} else {
		decoder := json.NewDecoder(strings.NewReader(string(body)))
		for decoder.More() {
			if err := decoder.Decode(&prometheusResults); err != nil {
				return nil, fmt.Errorf("error decoding results: %s", err)
			}
			resultsArray = prometheusResults.Data.Result
		}
	}

	return resultsArray, nil
}

func extractIntValue(queryResult map[string]interface{}) (int32, error) {
	// "value" array contains timestamp at index 0 and count at index 1.
	metric := queryResult["value"].([]any)[1].(string)
	metricInt, err := strconv.ParseInt(metric, 10, 32)
	if err != nil {
		return 0, errors.WithMessage(err, "failed to convert string to int")
	}

	return int32(metricInt), nil
}

func extractFloatValue(queryResult map[string]interface{}) (float64, error) {
	metric := queryResult["value"].([]any)[1].(string)
	metricFloat, err := strconv.ParseFloat(metric, 64)
	if err != nil {
		return 0, errors.WithMessage(err, "failed to convert string to float")
	}

	return metricFloat, nil
}

// getQueryResults builds and sends the query to the Prometheus API and returns the response body.
func getQueryResults(promSecret prometheusSecret, prometheusUrl, query string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, prometheusUrl, nil)
	q := req.URL.Query()

	// Adding and encoding query to URL
	q.Add("query", query)
	req.URL.RawQuery = q.Encode()

	if err != nil {
		return nil, fmt.Errorf("failed to create request to Prometheus API, err:  %s", err.Error())
	}

	req.SetBasicAuth(promSecret.Username, promSecret.Password)

	response, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to complete request to Prometheus API, err:  %s", err.Error())
	}

	zap.S().Debug("Finished querying results")
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}
