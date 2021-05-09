package logging

import (
	"github.com/sirupsen/logrus"
)

// Client provides methods for package level logger logrus
type Client struct {
}

func (c *Client) SetLevel(level logrus.Level) {
	c.Infof("Setting level to ")
	logrus.SetLevel(level)
}

// Infof ...
func (c *Client) Infof(log string) {
	logrus.Info(log)
}
