/**
 *
 * @author  chosen0ne(louzhenlin86@126.com)
 * @date    2017-11-14 14:56:32
 */

package redisproxy

import (
	util "github.com/chosen0ne/goutils"
	"time"
)

var (
	cmdTable []CommandHandler = []CommandHandler{
		&pingCommandHandler{&DefaultCommandHandler{"PING", 0}},
	}
)

type commandStats struct {
	calls int
	time  int64
}

type commandEngine struct {
	commandsTable map[string]CommandHandler
	cmdStatsTable map[string]commandStats
}

func newCommandEngine() *commandEngine {
	table := &commandEngine{make(map[string]CommandHandler), make(map[string]commandStats)}

	// as cmdTable is a object array, to avoid copy each object in for-range, so we use for-i here
	for i := 0; i < len(cmdTable); i++ {
		handler := cmdTable[i]
		table.addCommandHandler(handler)
	}

	return table
}

func (e *commandEngine) addCommandHandler(handler CommandHandler) {
	e.commandsTable[handler.Name()] = handler
	e.cmdStatsTable[handler.Name()] = commandStats{}
}

func (e *commandEngine) handle(cmd *Command) (Response, error) {
	cmdHandler, ok := e.commandsTable[cmd.Name()]
	if !ok {
		return nil, util.NewErr("unknown command, cmd: %s", cmd.Name())
	}

	if err := validateCmd(cmdHandler, cmd); err != nil {
		return NewError(err.Error()), nil
	}

	s := time.Now().UnixNano()
	resp, err := cmdHandler.Handle(cmd)
	elapsed := time.Now().UnixNano() - s

	if stats, ok := e.cmdStatsTable[cmd.Name()]; ok {
		stats.calls++
		stats.time += elapsed
	} else {
		panic(util.NewErr("unexpected error, a command handler without stats"))
	}

	if err != nil {
		return nil, util.WrapErrorf(err, "failed to handle command, cmd info: %s", cmd)
	} else {
		return resp, nil
	}
}

func validateCmd(cmdHandler CommandHandler, cmd *Command) error {
	arity := cmdHandler.Arity()
	if arity == ARITY_ODD || arity == ARITY_EVEN {
		if cmd.ArgCount()%2 != arity%2 {
			return util.NewErr("the odevity of the command is error")
		}
		return nil
	}

	if cmd.ArgCount() > ARITY_MAX {
		return util.NewErr("argumentss number exceeds the limit, arg count: %d", cmd.ArgCount())
	}

	if arity > 0 && arity != cmd.ArgCount() {
		return util.NewErr("arguemnts number is not matched with the cmd, cmd: %d, need: %d",
			cmd.ArgCount(), arity)
	}

	if arity < 0 && cmd.ArgCount() > 0-arity {
		return util.NewErr("arguments number is too many, only %d arguments at most", 0-arity)
	}

	return nil
}
