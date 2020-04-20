package browserker

import "github.com/wirepair/gcd/gcdapi"

type HttpRequest struct {
	gcdapi.NetworkRequest
	requestID string
	browserID string
}

type HttpResponse struct {
	gcdapi.NetworkResponse
	requestID string
	browserID string
}
