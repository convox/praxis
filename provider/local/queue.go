package local

import (
	"fmt"
	"io/ioutil"
	"path"
	"time"

	"github.com/convox/praxis/types"
)

func (p *Provider) QueueFetch(app, queue string, opts types.QueueFetchOptions) (map[string]string, error) {
	tsChan := make(chan string)
	errChan := make(chan error)
	done := make(chan bool)

	timeout := 10
	if opts.Timeout > 0 {
		timeout = opts.Timeout
	}

	go func() {
		for {
			select {
			case <-time.Tick(100 * time.Millisecond):
				dirs, err := ioutil.ReadDir(path.Join(p.Root, fmt.Sprintf("apps/%s/queues/%s/", app, queue)))
				if err != nil {
					errChan <- err
					return
				}

				if len(dirs) > 0 {
					tsChan <- dirs[0].Name()
					return
				}

			case <-done:
				return
			}
		}
	}()

	var ts string
	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		done <- true
		return nil, nil
	case err := <-errChan:
		return nil, err
	case ts = <-tsChan:
		//setting value
	}

	var attrs map[string]string
	if err := p.Load(fmt.Sprintf("apps/%s/queues/%s/%s/attrs.json", app, queue, ts), &attrs); err != nil {
		return nil, err
	}

	if err := p.DeleteAll(fmt.Sprintf("apps/%s/queues/%s/%s", app, queue, ts)); err != nil {
		return nil, err
	}

	return attrs, nil
}

func (p *Provider) QueueStore(app, queue string, attrs map[string]string) error {
	if err := p.Store(fmt.Sprintf("apps/%s/queues/%s/%d/attrs.json", app, queue, time.Now().UnixNano()), attrs); err != nil {
		return err
	}

	return nil
}
