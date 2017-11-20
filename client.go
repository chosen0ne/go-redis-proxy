/**
 *
 * @author  chosen0ne(louzhenlin86@126.com)
 * @date    2017-03-15 12:17:03
 */

package redisproxy

import (
	"bufio"
	"io"
	"net"
	"time"
)

const (
	_CMD_COUNT_FOR_STATS = 10
)

type redisClient struct {
	conn       net.Conn
	lastIOtime int64 // used to close the inactive clients
	in         *bufio.Reader
	out        *bufio.Writer
	msgOutChan chan<- interface{}
	engine     *commandEngine
	stats      statsMsg // used to record the stats of a command
}

func newClient(conn net.Conn, outChan chan<- interface{}, engine *commandEngine) *redisClient {
	r := &redisClient{conn: conn, lastIOtime: 0, engine: engine}
	r.in = bufio.NewReader(conn)
	r.out = bufio.NewWriter(conn)
	r.msgOutChan = outChan

	return r
}

func (cli *redisClient) processRequests() {
	defer cli.recoverFunc()

	stats := &statsMsg{}
	for {
		cli.stats.reset()

		// read command
		cmd, err := cli.parseCommand()
		if err != nil {
			if err == io.EOF {
				// client has closed the connection
				logger.Debug("connection has been closed by %s", cli.conn.RemoteAddr().String())
			} else {
				logger.Exception(err, "faield to parse a command, close. client: %s",
					cli.conn.RemoteAddr())
			}

			// Connection recycle is done by main goroutine
			cli.msgOutChan <- &cliCloseMsg{cli}

			// jump out of the for loop
			break
		}

		cli.lastIOtime = time.Now().UnixNano()

		// successful to parse a command
		debugInfo("recv command, cmd: %s", cmd)

		if response, err := cli.engine.handle(cmd); err != nil {
			logger.Exception(err, "failed to handler cmd")
		} else {
			logger.Track("send response, resp: %s, cli: %s", response, cli)

			if err := cli.response(response); err != nil {
				logger.Exception(err, "failed to response, cli: %s, resp: %s", cli, response)
			}
		}

		stats.add(&cli.stats)
		if stats.cmdCount >= _CMD_COUNT_FOR_STATS {
			cli.msgOutChan <- stats
			stats.reset()
		}
	}
}

func (cli *redisClient) close() {
	if err := cli.conn.Close(); err != nil {
		logger.Exception(err, "failed to close client, client: %s", cli.conn.RemoteAddr().String())
	}
}

func (cli *redisClient) recoverFunc() {
	if err := recover(); err != nil {
		switch err.(type) {
		case error:
			e := err.(error)
			logger.Exception(e, "panic recovered, client will be closed. client: %s", cli.conn.RemoteAddr())
		default:
			logger.Error("unknown error recovered, client will be closed. client: %v", err)
		}

		// recycle
		cli.msgOutChan <- &cliCloseMsg{cli}
	}
}

func (cli *redisClient) String() string {
	return cli.conn.RemoteAddr().String()
}

func (cli *redisClient) response(resp Response) error {
	n, err := responseTo(cli.out, resp)
	if err != nil {
		return err
	}

	cli.stats.outputBytes += n
	if err := cli.out.Flush(); err != nil {
		return err
	}
	return nil
}
