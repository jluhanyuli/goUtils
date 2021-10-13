package network
import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/segmentio/ksuid"
	"io"
	"net"
	"strings"
	"errors"
	"time"
)
type TcpProtocol interface {
	Packing(data []byte) []byte //打包数据
	UnPacking(data []byte)      //解析数据
	ReadMsg() <-chan []byte     //读取解析出来的有效数据
}

//异常处理
func recoverPainc(lg *logrus.Logger, f ...func()) {
	if r := recover(); r != nil {
		if lg != nil {
			lg.Fatalf("异常:%s", r)
		}
		if f != nil {
			for _, fu := range f {
				go fu()
			}
		}
	}
}
// tcp服务端配置
type TcpServerConfig struct {
	Port               int64                                                  //监听端口
	ServerAddress      string                                                 //服务监听地址
	Log               *logrus.Logger                                          //日志
	ErrorHandler       func(errType ErrorType, err error, clientID ...string) //错误处理
	NewClientHandler   func(clientID string) TcpProtocol                      //新客户端连接回调,返回该客户端处理协议,可以返回nil
	MessageHandler     func(clientID string, msg []byte)                      //消息处理
	CloseHandler       func()                                                 //服务关闭回调
	ClientCloseHandler func(clientID string)                                  //客户端关闭回调
}

//TCP服务器对象
type TcpServer struct {
	ctx      context.Context
	config   *TcpServerConfig
	conn     net.Listener
	clients  map[string]*SClient
	readData chan []byte //读取到的数据
	doClose  bool        //关闭操作
}

//新建一个服务端
func NewTcpServer(ctx context.Context, config *TcpServerConfig) (*TcpServer, error) {
	if config == nil {
		return nil, errors.New("配置信息不能为空")
	}
	ret := &TcpServer{
		ctx:     ctx,
		config:  config,
		clients: map[string]*SClient{},
	}
	return ret, nil
}

//启动监听该方法会
func (s *TcpServer) Listen() {
	var err error
	s.doClose = false
	s.conn, err = net.Listen("tcp", fmt.Sprintf("%s:%d", s.config.ServerAddress, s.config.Port))
	if err != nil {
		if s.config.Log != nil {
			s.config.Log.Infof("查询Socket监听失败:%s", err.Error())
		}
		if s.config.ErrorHandler != nil {
			s.config.ErrorHandler(ListenErr, err)
		}
		return
	}
	if s.config.Log != nil {
		s.config.Log.Infof("服务器监听开启 => %s:%d", s.config.ServerAddress, s.config.Port)
	}
	go s.handleConn()
	<-s.ctx.Done()
	s.Close()
}

//关闭服务
func (s *TcpServer) Close() {
	s.doClose = true
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	if s.config.CloseHandler != nil {
		s.config.CloseHandler()
	}
	for _, v := range s.clients {
		v.Close()
	}
}

//处理新连接
func (s *TcpServer) handleConn() {
	defer recoverPainc(s.config.Log, s.handleConn)
	for {
		if conn, err := s.conn.Accept(); err != nil {
			if s.doClose { //关闭
				return
			}
			if s.config.Log != nil {
				s.config.Log.Errorf("监听请求连接失败:%s", err.Error())
			}
			if s.config.ErrorHandler != nil {
				s.config.ErrorHandler(ListenErr, err)
			}
			s.Close()
			return
		} else {
			s.newClientAccept(conn)
		}
	}
}

//发送数据
func (s *TcpServer) Write(clientID string, msg []byte) error {
	if s.conn == nil {
		return errors.New("连接已关闭")
	}
	client := s.clients[clientID]
	if client == nil {
		return errors.New("客户端链接不存在")
	}
	return client.write(msg)
}

//读取消息
func (s *TcpServer) messageRecv(clientID string, msg []byte) {
	if s.config.MessageHandler != nil {
		s.config.MessageHandler(clientID, msg)
	}
}

// 关闭指定客户端
func (s *TcpServer) CloseClient(clientID string) {
	if c, ok := s.clients[clientID]; ok {
		if c != nil {
			c.Close()
		} else {
			delete(s.clients, clientID)
		}
	}
}

func (s *TcpServer) newClientAccept(conn net.Conn) {
	sclient := &SClient{
		conn:   conn,
		server: s,
		ID:     clientIDGen(),
	}
	var protocol TcpProtocol
	if s.config.NewClientHandler != nil {
		protocol = s.config.NewClientHandler(sclient.ID)
	}
	sclient.protocol = protocol
	sclient.doClose = false
	go sclient.readData()
	if protocol != nil {
		go sclient.recvProtocolMsg()
	}
	s.clients[sclient.ID] = sclient
}

// 客户端id生成器
func clientIDGen() string {
	id := ksuid.New()
	return strings.ToUpper(id.String())
}

//TCP连接到服务器的客户端
type SClient struct {
	conn     net.Conn
	server   *TcpServer
	protocol TcpProtocol
	ID       string //客户端唯一标示
	doClose  bool   //关闭
}

//读取客户端数据
func (s *SClient) readData() {
	defer s.Close()
	data := make([]byte, 1024)
	for {
		i, err := s.conn.Read(data)
		if s.doClose {
			return
		} else if err != nil {
			if err == io.EOF { //读取结束
				return
			}
			if s.server.config.Log != nil {
				s.server.config.Log.Errorf("%s=>数据读取错误:%s", s.ID, err.Error())
			}
			if s.server.config.ErrorHandler != nil {
				s.server.config.ErrorHandler(ReadErr, err, s.ID)
			}
			return
		}
		if s.protocol != nil {
			s.protocol.UnPacking(data[0:i])
		} else {
			s.server.messageRecv(s.ID, data[0:i])
		}
	}
}

// 获取通讯协议解析出来的消息
func (s *SClient) recvProtocolMsg() {
	defer recoverPainc(s.server.config.Log, s.recvProtocolMsg)
	if s.protocol != nil {
		for {
			msg := <-s.protocol.ReadMsg()
			s.server.messageRecv(s.ID, msg)
		}
	}
}

//发送数据
func (s *SClient) write(msg []byte) error {
	if s.conn == nil {
		return errors.New("连接已关闭")
	}
	if s.protocol != nil {
		msg = s.protocol.Packing(msg)
	}
	_, err := s.conn.Write(msg)
	return err
}

//关闭连接
func (s *SClient) Close() {
	if !s.doClose && s.server.config.ClientCloseHandler != nil {
		s.server.config.ClientCloseHandler(s.ID)
	}
	s.doClose = true
	delete(s.server.clients, s.ID)
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}


const bitlength = 8 //数据长度占用字节数

//Protocol 自定义协议解析
type Protocol struct {
	data             chan []byte   //解析成功的数据
	byteBuffer       *bytes.Buffer //数据存储中心
	dataLength       int64         //当前消息数据长度
	heartBeatData    []byte        //心跳内容
	heartBeatDataHex string        //心跳内容字符串
}

//NewProtocol 初始化一个Protocol
// chanLength为解析成功数据channel缓冲长度
func NewProtocol(heartBeatData []byte, chanLength ...int) *Protocol {
	length := 100
	if chanLength != nil && len(chanLength) > 0 {
		length = chanLength[0]
	}
	return &Protocol{
		data:             make(chan []byte, length),
		byteBuffer:       bytes.NewBufferString(""),
		heartBeatData:    heartBeatData,
		heartBeatDataHex: hex.EncodeToString(heartBeatData),
	}
}

//Packet 封包
func (p *Protocol) Packing(message []byte) []byte {
	return append(intToByte(int64(len(message))), message...)
}

//Read 获取数据读取的channel对象
func (p *Protocol) ReadMsg() <-chan []byte {
	return p.data
}

//解析成功的数据请用Read方法获取
func (p *Protocol) UnPacking(buffer []byte) {
	p.byteBuffer.Write(buffer)
	for { //多条数据循环处理
		length := p.byteBuffer.Len()
		if length < bitlength { //前面8个字节是长度
			return
		}
		p.dataLength = byteToInt(p.byteBuffer.Bytes()[0:bitlength])
		if int64(length) < p.dataLength+bitlength { //数据长度不够,等待下次读取数据
			return
		}
		data := make([]byte, p.dataLength+bitlength)
		p.byteBuffer.Read(data)
		msg := data[bitlength:]
		if p.heartBeatData != nil && len(msg) == len(p.heartBeatData) &&
			p.heartBeatDataHex == hex.EncodeToString(msg) {
			continue
		}
		p.data <- msg
	}
}

//重置
func (p *Protocol) Reset() {
	p.dataLength = 0
	p.byteBuffer.Reset() //清空重新开始
}

func intToByte(num int64) []byte {
	ret := make([]byte, 8)
	binary.BigEndian.PutUint64(ret, uint64(num))
	return ret
}

func byteToInt(data []byte) int64 {
	return int64(binary.BigEndian.Uint64(data))
}


// tcp客户端配置
type TcpClientConfig struct {
	ServerAddress     string                 //服务地址
	AutoReConnect     bool                   //自动重连
	ReConnectWaitTime time.Duration          //重连等待时间
	Protocol          TcpProtocol            //连接处理协议,如果没有则收到什么数据返回什么数据,写入什么数据发送什么数据
	Log               *logrus.Logger           //日志
	ErrorHandler      func(ErrorType, error) //错误处理
	ConnectHandler    func()                 //连接成功调用
	MessageHandler    func([]byte)           //消息处理
	CloseHandler      func()                 //连接关闭处理
	HeartBeatData     []byte                 //心跳内容,如果为空不发送心跳信息
	HeartBeatTime     time.Duration          //心跳间隔
	ConnectTimeOut    time.Duration          //连接超时事件
}

// tcp客户端
type TcpClient struct {
	ctx           context.Context  //上下文context
	conn          net.Conn         //socket连接
	config        *TcpClientConfig //配置
	isConnect     chan bool        //是否连接
	reConnectChan chan bool        //重连出发器
	connectSucc   bool             //连接成功
	doClose       bool             //关闭操作
}

//创建Tcp客户端
func NewTcpClient(ctx context.Context, config *TcpClientConfig) (*TcpClient, error) {
	if config == nil {
		return nil, errors.New("配置信息不能为空")
	}
	ret := &TcpClient{
		config:        config,
		ctx:           ctx,
		isConnect:     make(chan bool, 1),
		reConnectChan: make(chan bool, 1),
	}
	go ret.processMsg() //解析数据
	go ret.stateCheckGoroutine()
	return ret, nil
}

//连接服务器,该方法会阻塞知道连接异常或ctx关闭
func (c *TcpClient) connectServer() {
	defer recoverPainc(c.config.Log, c.connectServer)
	if c.connectSucc {
		return
	}
	c.doClose = false
	var err error
	c.conn, err = net.DialTimeout("tcp",
		c.config.ServerAddress, c.config.ConnectTimeOut)
	if err != nil {
		c.connectSucc = false
		c.isConnect <- false
		if c.config.ErrorHandler != nil {
			c.config.ErrorHandler(ConnectErr, err)
		}
		c.reConnectChan <- true //发起重连接
		return
	}
	c.connectSucc = true
	go c.heartbeat() //心跳...
	go c.readData()  //接受数据
	if c.config.ConnectHandler != nil {
		c.config.ConnectHandler()
	}
	c.isConnect <- true
	if c.config.Log != nil {
		c.config.Log.Info("服务器连接成功")
	}
}

//连接服务器返回连接结果是否成功
func (c *TcpClient) ConnectAsync() <-chan bool {
	go c.connectServer()
	return c.isConnect
}

//连接服务器返回连接结果是否成功
func (c *TcpClient) Connect() bool {
	c.connectServer()
	return <-c.isConnect
}

//读取数据
func (c *TcpClient) readData() {
	data := make([]byte, 1024)
	for {
		if c.conn == nil || c.doClose {
			return
		}
		i, err := c.conn.Read(data)
		if c.doClose {
			return
		}
		if err != nil {
			if err != io.EOF && c.config.ErrorHandler != nil {
				c.config.ErrorHandler(ReadErr, err)
			}
			c.connectClose()
			c.reConnectChan <- true //发起重连接
			return
		}
		if c.config.Protocol != nil {
			c.config.Protocol.UnPacking(data[0:i])
		} else if c.config.MessageHandler != nil {
			c.config.MessageHandler(data[0:i])
		}
	}
}

// 状态检测线程
func (c *TcpClient) stateCheckGoroutine() {
	defer recoverPainc(c.config.Log, c.stateCheckGoroutine)
	for {
		select {
		case <-c.reConnectChan:
			if c.connectSucc || !c.config.AutoReConnect { //连接成功或者不需要重连的直接返回
				continue
			}
			time.Sleep(c.config.ReConnectWaitTime)
			if c.doClose {
				return
			}
			c.Connect()
		case <-c.ctx.Done(): //如果上下文结束,关闭整个连接
			c.Close()
			return
		}

	}
}

//发送数据
func (c *TcpClient) Write(data []byte) error {
	if c.conn == nil {
		return errors.New("未连接服务器")
	}
	_, err := c.conn.Write(c.packingData(data))
	return err
}

//编码数据
func (c *TcpClient) packingData(msg []byte) []byte {
	if c.config.Protocol != nil {
		return c.config.Protocol.Packing(msg)
	}
	return msg
}

//解析返回收到的有效内容
func (c *TcpClient) processMsg() {
	defer recoverPainc(c.config.Log, c.processMsg)
	if c.config.Protocol != nil {
		for {
			v := <-c.config.Protocol.ReadMsg()
			if c.config.MessageHandler != nil {
				c.config.MessageHandler(v)
			}
		}
	}
}

//关闭连接
func (c *TcpClient) Close() {
	c.doClose = true
	c.connectClose()
}

func (c *TcpClient) connectClose() {
	if c.connectSucc {
		c.connectSucc = false
		if c.config.CloseHandler != nil {
			c.config.CloseHandler()
		}
	}
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

//心跳包
func (c *TcpClient) heartbeat() {
	defer recoverPainc(c.config.Log, c.heartbeat)
	if c.config.HeartBeatData == nil { //没有心跳需求
		return
	}
	t := time.NewTicker(c.config.HeartBeatTime)
	for {
		select {
		case <-t.C:
			if c.conn == nil {
				continue
			}
			_, err := c.conn.Write(c.packingData(c.config.HeartBeatData))
			if err != nil {
				if c.config.ErrorHandler != nil {
					c.config.ErrorHandler(SendErr, errors.New("心跳发送失败"))
				}
				c.Close()   //关闭之前连接
				c.Connect() //重新连接
			}
		case <-c.ctx.Done():
			return
		}
	}
}