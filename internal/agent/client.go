package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/cryptoutil"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/hashutil"
	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/models"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/pool"
)

// outboundIP определяет IP-адрес хоста агента, который будет отправляться
// серверу в заголовке X-Real-IP, чтобы тот мог проверить принадлежность
// агента доверенной подсети. Используется трюк с "подключением" по UDP:
// реальные пакеты при этом не отправляются, ядро только выбирает локальный
// адрес интерфейса, через который прошёл бы трафик к указанному хосту.
// Если определить адрес не удалось (например, нет сети), возвращается
// пустая строка — заголовок в этом случае просто не будет добавлен.
func outboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return ""
	}

	return localAddr.IP.String()
}

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
	pubKey     *rsa.PublicKey
	bufPool    *pool.Pool[*bytes.Buffer]
	realIP     string
}

func NewHTTPSender(serverAddr string, key string) *HTTPSender {
	return &HTTPSender{
		serverAddr: serverAddr,
		key:        key,
		bufPool: pool.New(func() *bytes.Buffer {
			return &bytes.Buffer{}
		}),
		realIP: outboundIP(),
	}
}

// SetPublicKey задаёт публичный ключ RSA, которым будут шифроваться тела
// запросов, отправляемых на сервер. Передача nil отключает шифрование.
func (s *HTTPSender) SetPublicKey(pub *rsa.PublicKey) {
	s.pubKey = pub
}

// bodyReader возвращает io.Reader для тела запроса: если публичный ключ
// не задан, возвращается fallback (несжатые/неизменённые данные), иначе
// payload шифруется публичным ключом и оборачивается в новый io.Reader.
func (s *HTTPSender) bodyReader(payload []byte, fallback io.Reader) (io.Reader, error) {
	if s.pubKey == nil {
		return fallback, nil
	}

	encrypted, err := cryptoutil.Encrypt(s.pubKey, payload)
	if err != nil {
		return nil, fmt.Errorf("encrypt payload: %w", err)
	}

	return bytes.NewReader(encrypted), nil
}

// setRealIP выставляет заголовок X-Real-IP с адресом хоста агента, если
// его удалось определить. Сервер использует этот заголовок, чтобы
// проверить принадлежность агента доверенной подсети.
func (s *HTTPSender) setRealIP(req *http.Request) {
	if s.realIP != "" {
		req.Header.Set("X-Real-IP", s.realIP)
	}
}

func (s *HTTPSender) SendGauge(name string, value float64) error {
	valueStr := strconv.FormatFloat(value, 'f', -1, 64)

	url := fmt.Sprintf("%s/update/gauge/%s/%s", s.serverAddr, name, valueStr)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	s.setRealIP(req)

	response, err := http.DefaultClient.Do(req)
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

	// Берём buffer из пула вместо создания нового
	buffer := s.bufPool.Get()
	defer s.bufPool.Put(buffer) // Reset() вызовется автоматически при возврате

	metric := models.Metrics{
		ID:    name,
		MType: models.Gauge,
		Value: &value,
	}

	if err := json.NewEncoder(buffer).Encode(metric); err != nil {
		return err
	}

	body, err := s.bodyReader(buffer.Bytes(), buffer)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	s.setRealIP(req)

	response, err := http.DefaultClient.Do(req)
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

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	s.setRealIP(req)

	resp, err := http.DefaultClient.Do(req)
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

	// Берём buffer из пула вместо создания нового
	buffer := s.bufPool.Get()
	defer s.bufPool.Put(buffer) // Reset() вызовется автоматически при возврате

	metric := models.Metrics{
		ID:    name,
		MType: models.Counter,
		Delta: &value,
	}

	if err := json.NewEncoder(buffer).Encode(metric); err != nil {
		return err
	}

	body, err := s.bodyReader(buffer.Bytes(), buffer)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	s.setRealIP(req)

	resp, err := http.DefaultClient.Do(req)
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
	// Берём buffer из пула вместо создания нового
	buf := s.bufPool.Get()
	defer s.bufPool.Put(buf) // Reset() вызовется автоматически при возврате

	gz := gzip.NewWriter(buf)
	if _, err = gz.Write(body); err != nil {
		return err
	}
	if err = gz.Close(); err != nil {
		return err
	}

	var hash string
	if s.key != "" {
		hash = hashutil.ComputeHash(buf.Bytes(), s.key)
	}

	reqBody, err := s.bodyReader(buf.Bytes(), buf)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		s.serverAddr+"/updates/",
		reqBody,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	s.setRealIP(req)

	if hash != "" {
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
