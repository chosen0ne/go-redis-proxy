/**
 *
 * @author  chosen0ne(louzhenlin86@126.com)
 * @date    2017-11-15 16:33:08
 */

package redisproxy

import (
	"bytes"
	"fmt"
	util "github.com/chosen0ne/goutils"
)

var (
	// wheather enable log the debug info or not
	// for some costful operation, such as command.String(),
	// so we support to disable log debug info
	enableDebug = false // wheather enable debug info or not
)

func debugInfo(format string, val ...interface{}) {
	if !enableDebug {
		return
	}

	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, format, val...)

	callerInfo := util.CallerInfo(2)
	logger.Info("|DEBUG INFO|%s|%s", callerInfo, string(buf.Bytes()))
}
