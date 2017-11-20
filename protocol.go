/**
 *
 * @author  chosen0ne(louzhenlin86@126.com)
 * @date    2017-03-15 15:37:42
 */

package redisproxy

import (
	"bytes"
	"chosen0ne.com/goutils"
	"io"
	"strconv"
)

const (
	STAR   = '*'
	PLUS   = '+'
	DOLLAR = '$'
	CR     = '\r'
	LF     = '\n'
)

func (cli *redisClient) parseCommand() (*Command, error) {
	b, err := cli.in.ReadByte()
	if err == io.EOF {
		// connection has been closed
		return nil, err
	} else if err != nil {
		// other errors
		logger.Error("failed to read mode byte")
		return nil, err
	}
	cli.stats.inputBytes++

	if b == STAR {
		// multi bulk command
		cli.stats.cmdCount++
		return cli.parseMutilBulkCmd()
	} else if b == PLUS {
		// one line command
		cli.stats.cmdCount++
		return cli.parseLineCmd()
	} else {
		logger.Error("unexpected mode byte, need '*' or '+', read: %c", b)
		return nil, goutils.NewErr("unexpected mode byte, need '*' or '+', read: %c", b)
	}
}

func (cli *redisClient) parseMutilBulkCmd() (cmd *Command, err error) {
	cmd = newCommand(cli)

	// 1. read bulk count
	var line []byte
	if line, err = cli.readLine(); err != nil {
		logger.Error("failed to read line for bulk count")
		return nil, goutils.NewErr("faield to read line for bulk count")
	}

	// e.g. 1\r\n，'*' is already skipped.
	var count int
	if count, err = strconv.Atoi(string(line)); err != nil {
		logger.Error("failed to convert to bulk count, line: %s", string(line))
		return nil, goutils.WrapErr(err)
	}
	cmd.setFieldCount(count)
	logger.Debug("field count: %d", cmd.fieldCount)

	// 2. read each bulk
	for i := 0; i < cmd.fieldCount; i++ {
		// 2.1 read string length
		if line, err = cli.readLine(); err != nil {
			logger.Error("failed to read line for bulk")
			return nil, goutils.NewErr("failed to read line for bulk")
		}
		if line[0] != DOLLAR {
			logger.Error("unexpected tag, expected: '$', read: '%c'", line[0])
			return nil, goutils.NewErr("unexpected line tag, expected: '$', read: '%c'", line[0])
		}
		if count, err = strconv.Atoi(string(line[1:])); err != nil {
			logger.Error("failed to convert to string length, line: %s", string(line))
			return nil, goutils.WrapErr(err)
		}

		// 2.2 read string
		if line, err = cli.readLine(); err != nil {
			logger.Error("failed to read line for string")
			return nil, goutils.WrapErr(err)
		}

		if len(line) != count {
			logger.Error("string length doesn't match, expected: %d, read: %s",
				count, string(line))
			return nil, goutils.NewErr("bulk string length doesn't match")
		}
		cmd.addField(line)
	}

	return cmd, nil
}

func (cli *redisClient) parseLineCmd() (*Command, error) {
	cmd := newCommand(cli)
	cmd.setFieldCount(1)

	if line, err := cli.readLine(); err != nil {
		logger.Error("failed to read line for line cmd")
		return nil, goutils.NewErr("failed to read line for line cmd")
	} else {
		cmd.addField(line)
	}

	return cmd, nil
}

func (cli *redisClient) readLine() ([]byte, error) {
	var buf bytes.Buffer

	// read buffer until \r\n
	for {
		data, err := cli.in.ReadBytes(LF)
		logger.Debug("read data, len: %d, d: %s", len(data), string(data))
		if err != nil {
			logger.Error("failed to read bulk")
			return nil, goutils.WrapErr(err)
		}

		cli.stats.inputBytes += len(data)

		if n, err := buf.Write(data); err != nil {
			logger.Error("failed to copy to buffer")
			return nil, goutils.WrapErr(err)
		} else if n != len(data) {
			logger.Error("cann't copy all the data to buffer")
			return nil, goutils.NewErr("cann't copy all the data to buffer")
		}

		if data[len(data)-2] == CR {
			break
		}
	}

	data := buf.Bytes()
	return data[:len(data)-2], nil
}

// size of the buffer of input stream of the socket
// used to close the client which send a large request
func (cli *redisClient) inputBufSize() int {
	return cli.in.Buffered()
}
