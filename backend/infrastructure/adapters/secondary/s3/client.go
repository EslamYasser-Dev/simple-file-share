package s3

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

const (
	emptyPayloadHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	awsService       = "s3"
	awsAlgorithm     = "AWS4-HMAC-SHA256"
	maxListKeys      = 1000
)

// Client is a minimal SigV4-signed S3 client (AWS, MinIO, R2, GCS-compat).
type Client struct {
	endpoint  string
	region    string
	bucket    string
	accessKey string
	secretKey string
	prefix    string
	pathStyle bool
	hc        *http.Client
	now       func() time.Time
}

func NewClient(settings ports.S3Settings) *Client {
	endpoint := strings.TrimRight(strings.TrimSpace(settings.Endpoint), "/")
	if endpoint == "" {
		endpoint = "https://s3." + settings.Region + ".amazonaws.com"
	}
	region := settings.Region
	if region == "" {
		region = "us-east-1"
	}
	pathStyle := settings.PathStyle
	if settings.Endpoint != "" && !strings.Contains(endpoint, "amazonaws.com") {
		pathStyle = true
	}
	return &Client{
		endpoint:  endpoint,
		region:    region,
		bucket:    settings.Bucket,
		accessKey: settings.AccessKey,
		secretKey: settings.SecretKey,
		prefix:    strings.Trim(settings.Prefix, "/"),
		pathStyle: pathStyle,
		hc:        &http.Client{},
		now:       time.Now,
	}
}

func (c *Client) objectKey(path string) string {
	path = strings.TrimPrefix(path, "/")
	if c.prefix == "" {
		return path
	}
	if path == "" {
		return c.prefix + "/"
	}
	return c.prefix + "/" + path
}

func (c *Client) fromObjectKey(key string) string {
	if c.prefix != "" {
		key = strings.TrimPrefix(key, c.prefix+"/")
	}
	return key
}

type listEntry struct {
	Key          string    `xml:"Key"`
	LastModified time.Time `xml:"LastModified"`
	Size         int64     `xml:"Size"`
}

type listBucketResult struct {
	IsTruncated    bool   `xml:"IsTruncated"`
	NextToken      string `xml:"NextContinuationToken"`
	Contents       []listEntry
	CommonPrefixes []struct {
		Prefix string `xml:"Prefix"`
	}
}

type errorResponse struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *errorResponse) Error() string {
	return fmt.Sprintf("s3: %s (%s)", e.Message, e.Code)
}

func isNotFound(err error) bool {
	er, ok := err.(*errorResponse)
	if !ok {
		return false
	}
	return er.StatusCode == http.StatusNotFound ||
		er.Code == "NoSuchKey" || er.Code == "NotFound" || er.Code == "NoSuchBucket"
}

func (c *Client) baseURL() string {
	if c.pathStyle {
		return c.endpoint
	}
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return c.endpoint + "/" + c.bucket
	}
	return u.Scheme + "://" + c.bucket + "." + u.Host
}

type callOptions struct {
	method        string
	key           string
	query         url.Values
	body          []byte
	bodyStream    io.Reader
	contentLength int64
	headers       map[string]string
	// payloadHash overrides the signed payload hash when bodyStream is used
	// without a precomputed digest (empty hash = unsigned/empty body).
	payloadHash string
}

func (c *Client) do(opts callOptions) (*http.Response, error) {
	base := c.baseURL()
	decodedPath, encodedPath := c.fullPath(opts.key)

	rawQuery := ""
	if opts.query != nil {
		rawQuery = opts.query.Encode()
	}

	reqURL := base + encodedPath
	if rawQuery != "" {
		reqURL += "?" + rawQuery
	}

	var bodyReader io.Reader
	switch {
	case opts.bodyStream != nil:
		// The HTTP client always closes req.Body; shield the caller's reader.
		bodyReader = nopCloser{opts.bodyStream}
	case len(opts.body) > 0:
		bodyReader = bytes.NewReader(opts.body)
	}

	req, err := http.NewRequest(opts.method, reqURL, bodyReader)
	if err != nil {
		return nil, err
	}
	req.URL.Path = decodedPath
	req.URL.RawPath = encodedPath
	if rawQuery != "" {
		req.URL.RawQuery = rawQuery
	}

	payloadHash := opts.payloadHash
	if payloadHash == "" {
		if len(opts.body) > 0 {
			sum := sha256.Sum256(opts.body)
			payloadHash = hex.EncodeToString(sum[:])
		} else if opts.bodyStream == nil {
			payloadHash = emptyPayloadHash
		} else {
			payloadHash = emptyPayloadHash
		}
	}
	if opts.contentLength > 0 {
		req.ContentLength = opts.contentLength
	} else if len(opts.body) > 0 {
		req.ContentLength = int64(len(opts.body))
	}
	for k, v := range opts.headers {
		req.Header.Set(k, v)
	}
	if opts.method == http.MethodPut && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/octet-stream")
	}
	c.sign(req, payloadHash)

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		return nil, parseError(resp)
	}
	return resp, nil
}

func encodePath(key string) string {
	parts := strings.Split(key, "/")
	for i, p := range parts {
		parts[i] = uriEncode(p, false)
	}
	return strings.Join(parts, "/")
}

// fullPath returns the decoded and SigV4-encoded request paths (including the
// bucket segment when addressing path-style).
func (c *Client) fullPath(key string) (decoded, encoded string) {
	if c.pathStyle {
		decoded = "/" + c.bucket
		encoded = "/" + uriEncode(c.bucket, true)
		if key != "" {
			decoded += "/" + key
			encoded += "/" + encodePath(key)
		}
		return decoded, encoded
	}
	if key == "" {
		return "/", "/"
	}
	return "/" + key, "/" + encodePath(key)
}

// nopCloser prevents net/http from closing a caller-owned reader.
type nopCloser struct{ io.Reader }

func (nopCloser) Close() error { return nil }

func (c *Client) sign(req *http.Request, payloadHash string) {
	now := c.now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")

	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("x-amz-content-sha256", payloadHash)
	req.Header.Set("host", req.URL.Host)

	signedHeaders := "host;x-amz-content-sha256;x-amz-date"
	canonicalHeaders := "host:" + req.URL.Host + "\n" +
		"x-amz-content-sha256:" + payloadHash + "\n" +
		"x-amz-date:" + amzDate + "\n"

	canonicalURI := req.URL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}
	canonicalQuery := canonicalQueryString(req.URL.RawQuery)

	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI,
		canonicalQuery,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	scope := dateStamp + "/" + c.region + "/" + awsService + "/aws4_request"
	stringToSign := strings.Join([]string{
		awsAlgorithm,
		amzDate,
		scope,
		hexSHA256([]byte(canonicalRequest)),
	}, "\n")

	signature := hex.EncodeToString(hmacSHA256(c.signingKey(dateStamp), stringToSign))
	req.Header.Set("Authorization", fmt.Sprintf(
		"%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		awsAlgorithm, c.accessKey, scope, signedHeaders, signature))
}

func (c *Client) signingKey(dateStamp string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+c.secretKey), dateStamp)
	kRegion := hmacSHA256(kDate, c.region)
	kService := hmacSHA256(kRegion, awsService)
	return hmacSHA256(kService, "aws4_request")
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func hexSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func canonicalQueryString(raw string) string {
	if raw == "" {
		return ""
	}
	values, err := url.ParseQuery(raw)
	if err != nil {
		return raw
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		vs := values[k]
		sort.Strings(vs)
		for _, v := range vs {
			parts = append(parts, uriEncode(k, true)+"="+uriEncode(v, true))
		}
	}
	return strings.Join(parts, "&")
}

func uriEncode(s string, encodeSlash bool) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch >= 'A' && ch <= 'Z', ch >= 'a' && ch <= 'z', ch >= '0' && ch <= '9',
			ch == '-', ch == '.', ch == '_', ch == '~':
			b.WriteByte(ch)
		case ch == '/' && !encodeSlash:
			b.WriteByte(ch)
		default:
			fmt.Fprintf(&b, "%%%02X", ch)
		}
	}
	return b.String()
}

func parseError(resp *http.Response) *errorResponse {
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
	er := &errorResponse{StatusCode: resp.StatusCode}
	var doc struct {
		Code    string `xml:"Code"`
		Message string `xml:"Message"`
	}
	if xml.Unmarshal(data, &doc) == nil && (doc.Code != "" || doc.Message != "") {
		er.Code = doc.Code
		er.Message = doc.Message
		return er
	}
	er.Code = http.StatusText(resp.StatusCode)
	er.Message = strings.TrimSpace(string(data))
	if er.Message == "" {
		er.Message = http.StatusText(resp.StatusCode)
	}
	return er
}

func (c *Client) headObject(key string) (size int64, mod time.Time, err error) {
	resp, err := c.do(callOptions{method: http.MethodHead, key: key})
	if err != nil {
		return 0, time.Time{}, err
	}
	defer resp.Body.Close()
	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		mod, _ = http.ParseTime(lm)
	}
	if resp.ContentLength >= 0 {
		size = resp.ContentLength
	}
	return size, mod, nil
}

func (c *Client) getObject(key string) (io.ReadCloser, int64, time.Time, error) {
	resp, err := c.do(callOptions{method: http.MethodGet, key: key})
	if err != nil {
		return nil, 0, time.Time{}, err
	}
	var mod time.Time
	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		mod, _ = http.ParseTime(lm)
	}
	return resp.Body, resp.ContentLength, mod, nil
}

func (c *Client) putObjectBytes(key string, body []byte) error {
	resp, err := c.do(callOptions{
		method: http.MethodPut,
		key:    key,
		body:   body,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

// putObjectStream signs with a precomputed SHA-256 (hex) of the full body.
func (c *Client) putObjectStream(key string, r io.Reader, size int64, payloadHash string) error {
	resp, err := c.do(callOptions{
		method:        http.MethodPut,
		key:           key,
		bodyStream:    r,
		contentLength: size,
		payloadHash:   payloadHash,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

func (c *Client) deleteObject(key string) error {
	resp, err := c.do(callOptions{method: http.MethodDelete, key: key})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

func (c *Client) copyObject(srcKey, dstKey string) error {
	headers := map[string]string{
		"x-amz-copy-source": "/" + uriEncode(c.bucket, true) + "/" + encodePath(srcKey),
	}
	resp, err := c.do(callOptions{
		method:  http.MethodPut,
		key:     dstKey,
		headers: headers,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

type listOptions struct {
	prefix    string
	delimiter string
	token     string
}

func (c *Client) listObjects(opts listOptions) (*listBucketResult, error) {
	q := url.Values{}
	q.Set("list-type", "2")
	q.Set("max-keys", fmt.Sprintf("%d", maxListKeys))
	if opts.prefix != "" {
		q.Set("prefix", opts.prefix)
	}
	if opts.delimiter != "" {
		q.Set("delimiter", opts.delimiter)
	}
	if opts.token != "" {
		q.Set("continuation-token", opts.token)
	}
	resp, err := c.do(callOptions{method: http.MethodGet, query: q})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result listBucketResult
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("s3: decode list: %w", err)
	}
	return &result, nil
}

func (c *Client) listAll(opts listOptions, fn func(*listBucketResult) error) error {
	for {
		result, err := c.listObjects(opts)
		if err != nil {
			return err
		}
		if err := fn(result); err != nil {
			return err
		}
		if !result.IsTruncated || result.NextToken == "" {
			return nil
		}
		opts.token = result.NextToken
	}
}
