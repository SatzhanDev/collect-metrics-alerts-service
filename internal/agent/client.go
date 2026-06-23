package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/hashutil"
	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/models"
)

type Sender interface {
	SendGauge(name string, value float64) error
	SendCounter(name string, value int64) error
	SendGaugeJSON(name string, value float64) error
	SendCounterJSON(name string, value int64) error
	SendBatch(ctx context.Context, metrics []models.Metrics) error
}

type HTTPSender struct {
	serverAddr string
	key        string
}

func NewHTTPSender(serverAddr string, key string) *HTTPSender {
	return &HTTPSender{
		serverAddr: serverAddr,
		key:        key,
	}
}

func (s *HTTPSender) SendGauge(name string, value float64) error {
	valueStr := strconv.FormatFloat(value, 'f', -1, 64)

	url := fmt.Sprintf("%s/update/gauge/%s/%s", s.serverAddr, name, valueStr)

	response, err := http.Post(url, "text/plain", nil)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", response.StatusCode)
	}

	return nil
}
func (s *HTTPSender) SendGaugeJSON(name string, value float64) error {

	url := s.serverAddr + "/update"

	var buffer bytes.Buffer
	req := models.Metrics{
		ID:    name,
		MType: models.Gauge,
		Value: &value,
	}

	if err := json.NewEncoder(&buffer).Encode(req); err != nil {
		return err
	}

	response, err := http.Post(url, "application/json", &buffer)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", response.StatusCode)
	}

	return nil
}

func (s *HTTPSender) SendCounter(name string, value int64) error {
	valueStr := strconv.FormatInt(value, 10)

	url := fmt.Sprintf("%s/update/counter/%s/%s",
		s.serverAddr,
		name,
		valueStr,
	)

	resp, err := http.Post(url, "text/plain", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}
func (s *HTTPSender) SendCounterJSON(name string, value int64) error {
	url := s.serverAddr + "/update"

	var buffer bytes.Buffer
	req := models.Metrics{
		ID:    name,
		MType: models.Counter,
		Delta: &value,
	}

	if err := json.NewEncoder(&buffer).Encode(req); err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", &buffer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

func (s *HTTPSender) SendBatch(ctx context.Context, metrics []models.Metrics) error {
	delays := []time.Duration{
		1 * time.Second,
		3 * time.Second,
		5 * time.Second,
	}

	var err error

	for i := 0; i <= len(delays); i++ {
		err = s.sendOnce(ctx, metrics)
		if err == nil {
			return nil
		}

		if i == len(delays) {
			break
		}

		time.Sleep(delays[i])
	}

	return err
}
func (s *HTTPSender) sendOnce(ctx context.Context, metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}
	body, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err = gz.Write(body); err != nil {
		return err
	}
	if err = gz.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		s.serverAddr+"/updates/",
		&buf,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	if s.key != "" {
		hash := hashutil.ComputeHash(buf.Bytes(), s.key)
		req.Header.Set("HashSHA256", hash)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	return nil
}
