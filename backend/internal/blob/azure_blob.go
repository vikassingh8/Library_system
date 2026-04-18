package blob

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client uploads blobs to Azure Blob Storage using the REST API directly
// (no SDK required – works without external dependencies).
type Client struct {
	accountName   string
	accountKey    []byte // decoded base64 key
	containerName string
	baseURL       string // https://<account>.blob.core.windows.net
}

// NewClientFromConnString parses an Azure Storage connection string and returns a Client.
// Expected format: DefaultEndpointsProtocol=https;AccountName=...;AccountKey=...;EndpointSuffix=core.windows.net
func NewClientFromConnString(connectionString, containerName string) (*Client, error) {
	params := parseConnString(connectionString)

	accountName, ok := params["AccountName"]
	if !ok || accountName == "" {
		return nil, fmt.Errorf("azure blob: AccountName missing in connection string")
	}

	accountKeyB64, ok := params["AccountKey"]
	if !ok || accountKeyB64 == "" {
		return nil, fmt.Errorf("azure blob: AccountKey missing in connection string")
	}

	keyBytes, err := base64.StdEncoding.DecodeString(accountKeyB64)
	if err != nil {
		return nil, fmt.Errorf("azure blob: failed to decode AccountKey: %w", err)
	}

	endpointSuffix := params["EndpointSuffix"]
	if endpointSuffix == "" {
		endpointSuffix = "core.windows.net"
	}

	protocol := params["DefaultEndpointsProtocol"]
	if protocol == "" {
		protocol = "https"
	}

	baseURL := fmt.Sprintf("%s://%s.blob.%s", protocol, accountName, endpointSuffix)

	return &Client{
		accountName:   accountName,
		accountKey:    keyBytes,
		containerName: containerName,
		baseURL:       baseURL,
	}, nil
}

// Upload uploads data to Azure Blob Storage and returns the public URL.
// blobName should be a unique filename (e.g. "covers/uuid.jpg").
func (c *Client) Upload(blobName string, data []byte, contentType string) (string, error) {
	// Ensure container exists (best-effort, ignore if already exists)
	_ = c.ensureContainer()

	blobURL := fmt.Sprintf("%s/%s/%s", c.baseURL, c.containerName, blobName)

	req, err := http.NewRequest(http.MethodPut, blobURL, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("azure blob: build request: %w", err)
	}

	now := time.Now().UTC().Format(http.TimeFormat)
	req.Header.Set("x-ms-date", now)
	req.Header.Set("x-ms-version", "2020-04-08")
	req.Header.Set("x-ms-blob-type", "BlockBlob")
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = int64(len(data))

	sig, err := c.signRequest(req, int64(len(data)), contentType, blobName)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("SharedKey %s:%s", c.accountName, sig))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("azure blob: upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("azure blob: upload failed (%d): %s", resp.StatusCode, string(body))
	}

	return blobURL, nil
}

// ensureContainer creates the container with public blob access if it doesn't exist.
func (c *Client) ensureContainer() error {
	containerURL := fmt.Sprintf("%s/%s?restype=container", c.baseURL, c.containerName)

	req, err := http.NewRequest(http.MethodPut, containerURL, nil)
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(http.TimeFormat)
	req.Header.Set("x-ms-date", now)
	req.Header.Set("x-ms-version", "2020-04-08")
	req.Header.Set("x-ms-blob-public-access", "blob") // public read for blobs
	req.ContentLength = 0

	sig, err := c.signContainer(req)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("SharedKey %s:%s", c.accountName, sig))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// 201 = created, 409 = already exists – both are fine
	return nil
}

// ─── Shared Key signing ──────────────────────────────────────────────────────

func (c *Client) signRequest(req *http.Request, contentLength int64, contentType, blobName string) (string, error) {
	date := req.Header.Get("x-ms-date")
	canonicalHeaders := fmt.Sprintf("x-ms-blob-type:%s\nx-ms-date:%s\nx-ms-version:%s\n",
		req.Header.Get("x-ms-blob-type"),
		date,
		req.Header.Get("x-ms-version"),
	)

	canonicalResource := fmt.Sprintf("/%s/%s/%s", c.accountName, c.containerName, blobName)

	// canonicalHeaders already ends with \n; concatenate directly so there is no
	// extra newline between the headers block and the canonical resource.
	stringToSign := strings.Join([]string{
		"PUT",                             // HTTP Verb
		"",                                // Content-Encoding
		"",                                // Content-Language
		fmt.Sprintf("%d", contentLength), // Content-Length
		"",                                // Content-MD5
		contentType,                       // Content-Type
		"",                                // Date
		"",                                // If-Modified-Since
		"",                                // If-Match
		"",                                // If-None-Match
		"",                                // If-Unmodified-Since
		"",                                // Range
	}, "\n") + "\n" + canonicalHeaders + canonicalResource

	return c.hmacSign(stringToSign)
}

func (c *Client) signContainer(req *http.Request) (string, error) {
	date := req.Header.Get("x-ms-date")
	canonicalHeaders := fmt.Sprintf("x-ms-blob-public-access:%s\nx-ms-date:%s\nx-ms-version:%s\n",
		req.Header.Get("x-ms-blob-public-access"),
		date,
		req.Header.Get("x-ms-version"),
	)
	canonicalResource := fmt.Sprintf("/%s/%s\nrestype:container", c.accountName, c.containerName)

	// Content-Length must be empty string (not "0") when the request body is empty.
	// canonicalHeaders already ends with \n; concatenate directly.
	stringToSign := strings.Join([]string{
		"PUT", "", "", "", "", "", "", "", "", "", "", "",
	}, "\n") + "\n" + canonicalHeaders + canonicalResource

	return c.hmacSign(stringToSign)
}

func (c *Client) hmacSign(stringToSign string) (string, error) {
	mac := hmac.New(sha256.New, c.accountKey)
	_, err := mac.Write([]byte(stringToSign))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func parseConnString(s string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(s, ";")
	for _, part := range parts {
		idx := strings.Index(part, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(part[:idx])
		val := strings.TrimSpace(part[idx+1:])
		result[key] = val
	}
	return result
}

// ContentTypeFromFilename guesses a MIME type from a filename extension.
func ContentTypeFromFilename(filename string) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".gif"):
		return "image/gif"
	case strings.HasSuffix(lower, ".webp"):
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

// SanitizeBlobName makes a filename safe to use as a blob name.
func SanitizeBlobName(filename string) string {
	return url.PathEscape(filename)
}
