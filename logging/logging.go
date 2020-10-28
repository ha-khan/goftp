package logging

import (
	"github.com/sirupsen/logrus"
)

// Client provides methods for package level logger logrus
type Client struct {
}

// Infof ...
func (c *Client) Infof(log string) {
	logrus.Info(log)
}
