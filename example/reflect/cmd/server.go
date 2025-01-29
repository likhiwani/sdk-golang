package cmd

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"ztna-core/sdk-golang/ziti"
)

func Server(zitiCfg *ziti.Config, serviceName string) {
	ctx, err := ziti.NewContext(zitiCfg)

	if err != nil {
		panic(err)
	}

	listener, err := ctx.Listen(serviceName)
	if err != nil {
		log.Panic(err)
	}
	serve(listener)

	sig := make(chan os.Signal)
	s := <-sig
	log.Infof("received %s: shutting down...", s)
}

func serve(listener net.Listener) {
	log.Infof("ready to accept connections")
	for {
		conn, _ := listener.Accept()
		log.Infof("new connection accepted")
		go accept(conn)
	}
}

func accept(conn net.Conn) {
	if conn == nil {
		panic("connection is nil!")
	}
	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)
	rw := bufio.NewReadWriter(reader, writer)
	//line delimited
	for {
		line, err := rw.ReadString('\n')
		if err != nil {
			log.Error(err)
			break
		}
		log.Info("about to read a string :")
		log.Infof("                  read : %s", strings.TrimSpace(line))
		resp := fmt.Sprintf("you sent me: %s", line)
		_, _ = rw.WriteString(resp)
		_ = rw.Flush()
		log.Infof("       responding with : %s", strings.TrimSpace(resp))
	}
}
