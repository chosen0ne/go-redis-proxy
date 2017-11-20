/**
 *
 * @author  chosen0ne(louzhenlin86@126.com)
 * @date    2017-03-15 11:09:41
 */

package redisproxy

import (
	"fmt"
	"github.com/chosen0ne/gologging"
	util "github.com/chosen0ne/goutils"
	"net"
	"runtime"
	"time"
)

var (
	logger *gologging.Logger
)

type RedisProxy interface {
	Start(confPath string) error
	Stop() error
	AddCommandHandler(handler CommandHandler)
}

type redisStats struct {
	totalInputCommands uint64
	totalInputBytes    uint64
	totalOutputBytes   uint64
	ops                float32
	lastTotalCommands  uint64
	clientMsgProcessed uint64
}

// redisServer implements RedisProxy interface
type redisServer struct {
	laddr          string
	maxClients     int
	logPath        string
	curClients     int
	cliMaxIdleTime int
	stats          redisStats
	clients        map[string]*redisClient
	// channel: client process goroutine -> main goroutine
	//		use to sync all concurrent updates to redisServer, such as client close.
	fromCliChan chan interface{}
	engine      *commandEngine // used to hanle each command request inbound
}

func NewRedisProxy() RedisProxy {
	return &redisServer{engine: newCommandEngine()}
}

func (srv *redisServer) Start(confPath string) error {
	// NOTICE: no need to recover a panic, as here is the start point of the main goroutine.
	// The main goroutine will exit if panic, and the process will also exit.

	// load config
	config, err := srv.loadConfig(confPath)
	if err != nil {
		gologging.Exception(err, "failed to load config, config path: %s", confPath)
		return util.WrapErr(err)
	}

	// init server
	srv.init(config)

	go util.GoroutineLoop(srv.processCliMsg, logger)
	go util.GoroutineScheduled(time.Second, srv.serverCron, logger)

	listener, err := net.Listen("tcp", srv.laddr)
	if err != nil {
		logger.Error("failed to config listen socket, listen addr: %s", srv.laddr)
		return util.WrapErr(err)
	}

	logger.Info("listen on %s", srv.laddr)
	for {
		srv.acceptLoop(listener)
	}
}

func (srv *redisServer) Stop() error {
	// TODO
	return nil
}

func (srv *redisServer) AddCommandHandler(handler CommandHandler) {
	srv.engine.addCommandHandler(handler)
}

func (srv *redisServer) init(config *redisConfig) error {
	srv.clients = make(map[string]*redisClient)
	srv.fromCliChan = make(chan interface{}, config.FromCliChanSize)
	srv.maxClients = config.MaxClients
	srv.laddr = fmt.Sprintf(":%d", config.Port)

	enableDebug = config.EnableDebug

	lv := gologging.NewLevelString(config.LogLevel)
	if !lv.IsValid() {
		return util.NewErr("unknown log level")
	}

	// init logging
	gologging.TimeRotate().LogPath(config.LogPath).FileName(config.LogFile).Level(lv).Config("redis")
	logger = gologging.GetLogger("redis")

	logger.Info("server init by config: %s", config)

	return nil
}

func (srv *redisServer) acceptLoop(listener net.Listener) {
	defer func() {
		if err := recover(); err != nil {
			switch err.(type) {
			case error:
				logger.Exception(err.(error), "panic recovered in accept loop")
			default:
				logger.Error("unknown error recovered in accept loop, info: %v", err)
			}
		}
	}()

	conn, err := listener.Accept()
	if err != nil {
		logger.Error("error occurs when accept, listen addr: %s", srv.laddr)
		// failed to init becaues of the failure of socket listening.
		// main goroutine exit
		runtime.Goexit()
	}

	if srv.curClients >= srv.maxClients {
		// TODO: send failed message to client and close connection
		handleMaxCliError(conn)
		return
	}

	logger.Info("accept connection from %s:%s", conn.RemoteAddr().Network(),
		conn.RemoteAddr().String())

	cli := newClient(conn, srv.fromCliChan, srv.engine)

	srv.clients[conn.RemoteAddr().String()] = cli

	go cli.processRequests()
	srv.curClients++
}

func handleMaxCliError(conn net.Conn) {
	resp := NewError("exceed max client count, close.")
	responseTo(conn, resp)
	if err := conn.Close(); err != nil {
		logger.Exception(err, "failed to close connection, client: %s", conn.RemoteAddr())
	}
}

func (srv *redisServer) processCliMsg() {
	msg := <-srv.fromCliChan
	switch msg.(type) {
	case *cliCloseMsg:
		srv.processCliClose(msg.(*cliCloseMsg))
	case *statsMsg:
		srv.processStats(msg.(*statsMsg))
	default:
		logger.Error("unknown message from client, msg: %v", msg)
	}

	srv.stats.clientMsgProcessed++
}

func (srv *redisServer) processCliClose(msg *cliCloseMsg) {
	srv.curClients--
	delete(srv.clients, msg.cli.conn.RemoteAddr().String())

	msg.cli.close()
}

func (srv *redisServer) processStats(msg *statsMsg) {
	srv.stats.totalInputBytes += uint64(msg.inputBytes)
	srv.stats.totalOutputBytes += uint64(msg.outputBytes)
	srv.stats.totalInputCommands += uint64(msg.cmdCount)
}

func (srv *redisServer) serverCron(t time.Time) {
	srv.stats.ops = float32(srv.stats.totalInputCommands - srv.stats.lastTotalCommands)
	srv.stats.lastTotalCommands = srv.stats.totalInputCommands

	logger.Debug("serverCron %v", t)
	//for ip, cli := range srv.clients {
	//    logger.Info("cli %s => %d", ip, cli.inputBufSize())
	//}
}
