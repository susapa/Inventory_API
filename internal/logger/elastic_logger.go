package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
)

var esClient *elasticsearch.Client

// InitElasticsearch initializes the ES client for logging
func InitElasticsearch() {
	esURL := os.Getenv("ELASTICSEARCH_URL")
	if esURL == "" {
		esURL = "http://localhost:9200"
	}

	cfg := elasticsearch.Config{
		Addresses: []string{esURL},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Printf("[ERROR] Failed to create Elasticsearch client: %v", err)
		return
	}

	// Verify connection
	res, err := client.Info()
	if err != nil {
		log.Printf("[ERROR] Failed to connect to Elasticsearch: %v (Is it running?)", err)
		return
	}
	defer res.Body.Close()

	esClient = client
	log.Println("[INFO] Successfully connected to Elasticsearch for logging")
}

// LogToElastic sends log data to Elasticsearch
func LogToElastic(index string, data interface{}) {
	if esClient == nil {
		return
	}

	payload, err := json.Marshal(data)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal log data: %v", err)
		return
	}

	// Dynamic index name by date: inventory-logs-2024-03-12
	indexName := fmt.Sprintf("%s-%s", index, time.Now().Format("2006-01-02"))

	_, err = esClient.Index(
		indexName,
		bytes.NewReader(payload),
		esClient.Index.WithContext(context.Background()),
	)

	if err != nil {
		log.Printf("[ERROR] Failed to send log to Elasticsearch: %v", err)
	}
}

// StructuredLog represents the log format for Elasticsearch
type StructuredLog struct {
	Timestamp  time.Time `json:"timestamp"`
	Level      string    `json:"level"`
	Message    string    `json:"message"`
	Path       string    `json:"path"`
	Method     string    `json:"method"`
	StatusCode int       `json:"status_code"`
	Latency    int64     `json:"latency_ms"` // Changed to int64 for numerical analysis
	IP         string    `json:"ip"`
}

// ElasticLoggerMiddleware logs every request to Elasticsearch
func ElasticLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next() // Process request
		latency := time.Since(start).Milliseconds()

		logData := StructuredLog{
			Timestamp:  time.Now(),
			Level:      "INFO",
			Message:    fmt.Sprintf("%s %s completed", c.Request.Method, c.Request.URL.Path),
			Path:       c.Request.URL.Path,
			Method:     c.Request.Method,
			StatusCode: c.Writer.Status(),
			Latency:    latency,
			IP:         c.ClientIP(),
		}

		// Send to ES in a goroutine so it doesn't block the request
		go LogToElastic("inventory-api-logs", logData)
	}
}
