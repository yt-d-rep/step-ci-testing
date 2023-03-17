// Package oapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.12.4 DO NOT EDIT.
package oapi

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

// Id defines model for Id.
type Id = int

// Morph defines model for Morph.
type Morph struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// Morphs defines model for Morphs.
type Morphs = []Morph

// NewMorph defines model for NewMorph.
type NewMorph struct {
	Name string `json:"name"`
}

// BadRequest defines model for BadRequest.
type BadRequest struct {
	Message string `json:"message"`
}

// NotFound defines model for NotFound.
type NotFound struct {
	Message string `json:"message"`
}

// PostMorphJSONRequestBody defines body for PostMorph for application/json ContentType.
type PostMorphJSONRequestBody = NewMorph

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// GetMorphs request
	GetMorphs(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// PostMorph request with any body
	PostMorphWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	PostMorph(ctx context.Context, body PostMorphJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetMorphById request
	GetMorphById(ctx context.Context, id Id, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) GetMorphs(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetMorphsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostMorphWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostMorphRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostMorph(ctx context.Context, body PostMorphJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostMorphRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetMorphById(ctx context.Context, id Id, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetMorphByIdRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewGetMorphsRequest generates requests for GetMorphs
func NewGetMorphsRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/morphs")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewPostMorphRequest calls the generic PostMorph builder with application/json body
func NewPostMorphRequest(server string, body PostMorphJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewPostMorphRequestWithBody(server, "application/json", bodyReader)
}

// NewPostMorphRequestWithBody generates requests for PostMorph with any type of body
func NewPostMorphRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/morphs")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewGetMorphByIdRequest generates requests for GetMorphById
func NewGetMorphByIdRequest(server string, id Id) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/morphs/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// GetMorphs request
	GetMorphsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetMorphsResponse, error)

	// PostMorph request with any body
	PostMorphWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostMorphResponse, error)

	PostMorphWithResponse(ctx context.Context, body PostMorphJSONRequestBody, reqEditors ...RequestEditorFn) (*PostMorphResponse, error)

	// GetMorphById request
	GetMorphByIdWithResponse(ctx context.Context, id Id, reqEditors ...RequestEditorFn) (*GetMorphByIdResponse, error)
}

type GetMorphsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Morphs
}

// Status returns HTTPResponse.Status
func (r GetMorphsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetMorphsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type PostMorphResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Morph
	JSON400      *struct {
		Message string `json:"message"`
	}
}

// Status returns HTTPResponse.Status
func (r PostMorphResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostMorphResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetMorphByIdResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Morph
	JSON404      *struct {
		Message string `json:"message"`
	}
}

// Status returns HTTPResponse.Status
func (r GetMorphByIdResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetMorphByIdResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// GetMorphsWithResponse request returning *GetMorphsResponse
func (c *ClientWithResponses) GetMorphsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetMorphsResponse, error) {
	rsp, err := c.GetMorphs(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetMorphsResponse(rsp)
}

// PostMorphWithBodyWithResponse request with arbitrary body returning *PostMorphResponse
func (c *ClientWithResponses) PostMorphWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostMorphResponse, error) {
	rsp, err := c.PostMorphWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostMorphResponse(rsp)
}

func (c *ClientWithResponses) PostMorphWithResponse(ctx context.Context, body PostMorphJSONRequestBody, reqEditors ...RequestEditorFn) (*PostMorphResponse, error) {
	rsp, err := c.PostMorph(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostMorphResponse(rsp)
}

// GetMorphByIdWithResponse request returning *GetMorphByIdResponse
func (c *ClientWithResponses) GetMorphByIdWithResponse(ctx context.Context, id Id, reqEditors ...RequestEditorFn) (*GetMorphByIdResponse, error) {
	rsp, err := c.GetMorphById(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetMorphByIdResponse(rsp)
}

// ParseGetMorphsResponse parses an HTTP response from a GetMorphsWithResponse call
func ParseGetMorphsResponse(rsp *http.Response) (*GetMorphsResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetMorphsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Morphs
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParsePostMorphResponse parses an HTTP response from a PostMorphWithResponse call
func ParsePostMorphResponse(rsp *http.Response) (*PostMorphResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostMorphResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Morph
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest

	}

	return response, nil
}

// ParseGetMorphByIdResponse parses an HTTP response from a GetMorphByIdWithResponse call
func ParseGetMorphByIdResponse(rsp *http.Response) (*GetMorphByIdResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetMorphByIdResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Morph
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON404 = &dest

	}

	return response, nil
}

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (GET /morphs)
	GetMorphs(ctx echo.Context) error

	// (POST /morphs)
	PostMorph(ctx echo.Context) error

	// (GET /morphs/{id})
	GetMorphById(ctx echo.Context, id Id) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// GetMorphs converts echo context to params.
func (w *ServerInterfaceWrapper) GetMorphs(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetMorphs(ctx)
	return err
}

// PostMorph converts echo context to params.
func (w *ServerInterfaceWrapper) PostMorph(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PostMorph(ctx)
	return err
}

// GetMorphById converts echo context to params.
func (w *ServerInterfaceWrapper) GetMorphById(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "id" -------------
	var id Id

	err = runtime.BindStyledParameterWithLocation("simple", false, "id", runtime.ParamLocationPath, ctx.Param("id"), &id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetMorphById(ctx, id)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/morphs", wrapper.GetMorphs)
	router.POST(baseURL+"/morphs", wrapper.PostMorph)
	router.GET(baseURL+"/morphs/:id", wrapper.GetMorphById)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/8RUwY7TMBD9ldXA0dt02T1UOfYAqhBdxHXVg0mmrVeJ7R1PClWVf0djJ9m2BApSETfH",
	"sd+8ee95DlC42juLlgPkByAM3tmA8WOuyy/40mBg+SqcZbRxqb2vTKHZOJs9B2dlLxRbrLWsPDmPxCaB",
	"1BiC3qAs8buufYWQg7E7XZnyZqerBkEB771sByZjN9C2CghfGkNYQv40QKyGg+7rMxYMrZwsMRRkvJCB",
	"XDjfUEe6VbB0/N41trx2A9bxzToCX5P8ckCVf4lQ5LAoT8rfDVjGMm6QpNVPjvw2NldVj2vIn84bMRdB",
	"zqibcoS1OsBbwjXk8CZ7zU7Wkc2W+C0RaVc9p1ScsY6L313ubg41NZHeRxt71J/csbo+s2bpqNbVRV/i",
	"xXFTjF07wWTDVUJAf1uYW8bAAqVghxSSZXeT6WQqFJ1Hq72BHO4n08k9KPCaU+9ZPciwwZg/6SCmT4yF",
	"D8idUOr0Ab6bTv8quBe1DWOpe/wYxWG9CTGw6aDY510YofvZhcQXkqIYeO7K/dWYvkaoTab9S0V+JYiC",
	"h1RqDGGglB2NyFENW9W7nx1M2V6MwHy/KGN0SNfISCG+4+M3a4ShRAtUF355p8fZZmpQ/aEGCxk1q/8p",
	"8sNlkYchPi6xjEqkXS/WaZnKFXEYNFRBDltmn2dZ3Ny6wPlsNpuBCNDB9gOlh29X7Y8AAAD//3V/NcMj",
	"BwAA",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
