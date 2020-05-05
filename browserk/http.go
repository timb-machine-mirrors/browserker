package browserk

import "github.com/wirepair/gcd/gcdapi"

type HTTPRequest struct {
	gcdapi.NetworkRequest
	requestID string
	browserID string
}

type HTTPResponse struct {
	gcdapi.NetworkResponse
	request HTTPRequest
}
