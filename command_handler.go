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
	ARITY_DFT  = 0       // no limit for argument number of the command
)

type CommandHandler interface {
	Name() string
	// Arity() returns the rule of the arguments for the command.
	// The value for Arity() may be as follows:
	//	ARITY_DFT: no limitation
	//	ARITY_EVEN: arguments count of the command must be even
	//	ARITY_ODD: arguments count of the command must be odd
	//  a positive number: arguments count of the command must equal the value
	//	a negative number: arguments count of the command must not exceed the value
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
