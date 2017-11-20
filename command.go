/**
 *
 * @author  chosen0ne(louzhenlin86@126.com)
 * @date    2017-03-15 16:11:06
 */

package redisproxy

import (
	"bytes"
	"fmt"
)

// Command is the request from clients
type Command struct {
	// fields = Name + Arguments...
	fields     [][]byte
	attrs      map[string]interface{}
	fieldCount int
	cli        *redisClient
}

func newCommand(cli *redisClient) *Command {
	cmd := &Command{cli: cli}
	cmd.attrs = make(map[string]interface{})
	cmd.fields = make([][]byte, 0)
	return cmd
}

func (cmd *Command) Name() string {
	return string(bytes.ToUpper(cmd.fields[0]))
}

func (cmd *Command) Arg(idx int) string {
	return string(cmd.fields[idx+1])
}

func (cmd *Command) Attr(name string) interface{} {
	attr, ok := cmd.attrs[name]
	if ok {
		return attr
	}

	return nil
}

func (cmd *Command) SetAttr(name string, attr interface{}) {
	cmd.attrs[name] = attr
}

func (cmd *Command) setFieldCount(count int) {
	cmd.fieldCount = count
}

func (cmd *Command) ArgCount() int {
	return cmd.fieldCount - 1
}

func (cmd *Command) addField(arg []byte) {
	cmd.fields = append(cmd.fields, arg)
}

func (cmd *Command) String() string {
	buf := &bytes.Buffer{}

	buf.WriteString("Command@{fields:\"")
	isFirst := true
	for _, f := range cmd.fields {
		if isFirst {
			isFirst = false
		} else {
			buf.WriteByte(' ')
		}
		buf.Write(f)
	}
	buf.WriteString("\"}@attrs=")
	fmt.Fprintf(buf, "%v", cmd.attrs)

	return string(buf.Bytes())
}

func (cmd *Command) Client() *redisClient {
	return cmd.cli
}
