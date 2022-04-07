package client

import (
	"fmt"
)

func (c *HTTPClient) ChainProxyURL(chainUUID string) string {
	return GetProxyURL(c.config.URL, chainUUID)
}

func (c *HTTPClient) ChainTesseraProxyURL(chainUUID string) string {
	return GetProxyTesseraURL(c.config.URL, chainUUID)
}

func GetProxyURL(proxyURL, chainUUID string) string {
	return fmt.Sprintf("%s/proxy/chains/%s", proxyURL, chainUUID)
}

func GetProxyTesseraURL(proxyURL, chainUUID string) string {
	return fmt.Sprintf("%s/proxy/chains/tessera/%s", proxyURL, chainUUID)
}
