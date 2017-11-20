/**
 *
 * @author  chosen0ne(louzhenlin86@126.com)
 * @date    2017-11-13 19:07:20
 */

package redisproxy

type cliCloseMsg struct {
	cli *redisClient
}

type statsMsg struct {
	inputBytes  int
	outputBytes int
	cmdCount    int
}

func (s *statsMsg) reset() {
	s.cmdCount = 0
	s.inputBytes = 0
	s.outputBytes = 0
}

func (s *statsMsg) add(b *statsMsg) {
	s.cmdCount += b.cmdCount
	s.inputBytes += b.inputBytes
	s.outputBytes += b.outputBytes
}
