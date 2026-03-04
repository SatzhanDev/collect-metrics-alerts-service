package agent

import (
	"fmt"
	"net/http"
	"strconv"
)

type Sender interface {
	SendGauge(name string, value float64) error
	SendCounter(name string, value int64) error
}

type HTTPSender struct {
	serverAddr string
}

func NewHTTPSender(serverAddr string) *HTTPSender {
	return &HTTPSender{
		serverAddr: serverAddr,
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
