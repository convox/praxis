package frontend

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
)

func createProxy(ip, port, target string) (net.Listener, error) {
	log := Log.At("proxy.create").Namespace("ip=%s port=%s target=%s", ip, port, target).Start()

	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		return nil, err
	}

	go handleListener(ln)

	log.Success()

	return ln, nil
}

func handleListener(ln net.Listener) {
	for {
		cn, err := ln.Accept()
		if err != nil {
			fmt.Printf("err = %+v\n", err)
			continue
		}

		go handleConnection(cn)
	}
}

func handleConnection(cn net.Conn) {
	log := Log.At("proxy.connect").Start()

	defer cn.Close()

	ip, port, err := net.SplitHostPort(cn.LocalAddr().String())
	if err != nil {
		cn.Write([]byte(fmt.Sprintf("error: %s\n", err)))
		log.Error(err)
		return
	}

	log = log.Namespace("ip=%s port=%s", ip, port)

	pi, err := strconv.Atoi(port)
	if err != nil {
		cn.Write([]byte(fmt.Sprintf("error: %s\n", err)))
		log.Error(err)
		return
	}

	ep, ok := endpoints[ip][pi]
	if !ok {
		cn.Write([]byte(fmt.Sprintf("no endpoint\n")))
		return
	}

	log = log.Namespace("target=%s host=%q", ep.Target, ep.Host)

	out, err := net.Dial("tcp", ep.Target)
	if err != nil {
		cn.Write([]byte(fmt.Sprintf("error: %s\n", err)))
		return
	}

	defer out.Close()

	var wg sync.WaitGroup

	wg.Add(2)

	go copyAsync(out, cn, &wg)
	go copyAsync(cn, out, &wg)

	wg.Wait()

	log.Success()
}

func copyAsync(w io.Writer, r io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	io.Copy(w, r)
}
