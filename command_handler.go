/**
 *
 * @author  chosen0ne(louzhenlin86@126.com)
 * @date    2017-11-14 18:15:09
 */

package redisproxy

const (
	ARITY_MAX  = 1000    // maximum number of the command arugement
	ARITY_EVEN = 1000000 // arguments number of the commend is even
	ARITY_ODD  = 1000001 // arguments number of the commend is odd
	ARITY_DFT  = -1      // no limit for argument number of the command
)

type CommandHandler interface {
	Name() string
	Arity() int
	Handle(cmd *Command) (Response, error)
}

type DefaultCommandHandler struct {
	CmdName  string
	CmdArity int
}

func (h *DefaultCommandHandler) Name() string {
	return h.CmdName
}

func (h *DefaultCommandHandler) Arity() int {
	return h.CmdArity
}

type pingCommandHandler struct {
	*DefaultCommandHandler
}

func (h *pingCommandHandler) Handle(cmd *Command) (Response, error) {
	return pongResponse, nil
}
