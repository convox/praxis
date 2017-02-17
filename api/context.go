package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/convox/logger"
	"github.com/gorilla/mux"
)

type Context struct {
	logger   *logger.Logger
	request  *http.Request
	response http.ResponseWriter
}

func (c *Context) Form(name string) string {
	return c.request.FormValue(name)
}

func (c *Context) Header(name string) string {
	return c.request.Header.Get(name)
}

func (c *Context) LogError(err error) {
	log := c.logger.At("end")

	switch t := err.(type) {
	case Error:
		switch t.Code / 100 {
		case 4:
			log.Logf("state=error type=user code=%d error=%q", t.Code, t.Error())
		case 5:
			log.Logf("state=error type=server code=%d error=%q", t.Code, t.Error())
		default:
			log.Logf("state=error type=unknown code=%d error=%q", t.Code, t.Error())
		}
	case error:
		log.Logf("state=error code=500 error=%q", t.Error())
	case nil:
	default:
		log.Logf("state=error code=500 error=%q", "unknown error type")
	}
}

func (c *Context) LogParams(names ...string) {
	params := make([]string, len(names))

	for i, name := range names {
		params[i] = fmt.Sprintf("%s=%q", name, c.request.FormValue(name))
	}

	c.logger.At("params").Logf(strings.Join(params, " "))
}

func (c *Context) LogSuccess() {
	c.logger.At("end").Success()
}

func (c *Context) Logf(format string, args ...interface{}) {
	c.logger.Logf(format, args...)
}

func (c *Context) RenderJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if _, err := c.response.Write(data); err != nil {
		return err
	}

	return nil
}

func (c *Context) Start(format string, args ...interface{}) {
	c.logger = c.logger.Start()
	c.logger.At("start").Logf(format, args...)
}

func (c *Context) Tag(format string, args ...interface{}) {
	c.logger = c.logger.Namespace(format, args...)
}

func (c *Context) Var(name string) string {
	return mux.Vars(c.request)[name]
}
