package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Message struct {
	Len  uint32
	Ver  byte
	Cmd  uint32
	Ts   uint32
	Seq  uint32
	Body []byte
}

func EncodeMessage(message *Message) []byte {
	if message.Len == 0 {
		return nil
	}

	var buf = make([]byte, message.Len)
	binary.BigEndian.PutUint32(buf[:4], message.Len)
	buf[4] = message.Ver
	binary.BigEndian.PutUint32(buf[5:9], message.Cmd)
	binary.BigEndian.PutUint32(buf[9:13], message.Ts)
	binary.BigEndian.PutUint32(buf[13:17], message.Seq)
	copy(buf[17:], message.Body)
	return buf
}

func DecodeMessage(content []byte) (*Message, error) {
	if len(content) < 4 {
		return nil, errors.New("长度缺失")
	}
	msg := &Message{}
	msg.Len = binary.BigEndian.Uint32(content[:4])
	if len(content) != int(msg.Len) {
		return nil, errors.New("内容缺失")
	}

	msg.Ver = content[4]
	msg.Cmd = binary.BigEndian.Uint32(content[5:9])
	msg.Ts = binary.BigEndian.Uint32(content[9:13])
	msg.Seq = binary.BigEndian.Uint32(content[13:17])
	msg.Body = make([]byte, msg.Len-17)
	copy(msg.Body[:], content[17:])
	return msg, nil
}

func ReadMessage(conn io.Reader) ([]byte, error) {
	var buf [4]byte
	n, err := conn.Read(buf[:])
	if err != nil {
		return nil, err
	}
	if n != 4 {
		return nil, errors.New("长度丢失")
	}

	bodyLen := binary.BigEndian.Uint32(buf[:])
	var content = make([]byte, bodyLen)
	copy(content[:4], buf[:])

	n, err = io.ReadFull(conn, content[4:])
	if err != nil {
		return nil, err
	}
	if uint32(n) != bodyLen-4 {
		return nil, errors.New("内容丢失")
	}
	return content, nil
}

type EventHandle = func(msg *Message, sender MessageSender)

type MessageSender interface {
	Input(message *Message) error
}

func NewClient(conn net.Conn) *Client {
	clt := &Client{
		conn:   conn,
		reader: make(chan *Message),
		writer: make(chan *Message),
		close:  make(chan struct{}),
		events: make(map[uint32]EventHandle),
	}
	go clt.loopReader()
	go clt.loopSender()
	return clt
}

type Client struct {
	conn     net.Conn
	reader   chan *Message
	writer   chan *Message
	close    chan struct{}
	closeFlg int32
	events   map[uint32]EventHandle
}

func (c *Client) RegEvent(cmdId uint32, event EventHandle) {
	c.events[cmdId] = event
}

func (c *Client) Loop() error {
	for {
		content, err := ReadMessage(c.conn)
		if err != nil {
			return err
		}
		msg, err := DecodeMessage(content)
		if err != nil {
			fmt.Println(err)
			continue
		}
		select {
		case <-c.close:
			return nil
		case c.reader <- msg:
		}
	}
}

func (c *Client) Input(message *Message) error {
	select {
	case <-c.close:
		return errors.New("conn is closed")
	case c.writer <- message:
		return nil
	}
}

func (c *Client) loopSender() {
	for {
		select {
		case <-c.close:
			return
		case msg := <-c.writer:
			if msg != nil {
				_, _ = c.conn.Write(EncodeMessage(msg))
			}
		}
	}
}

func (c *Client) loopReader() {
	for {
		select {
		case <-c.close:
			return
		case msg := <-c.reader:
			if fun, ok := c.events[msg.Cmd]; ok {
				go fun(msg, c)
			}
		}
	}
}

func (c *Client) Close() {
	if atomic.CompareAndSwapInt32(&c.closeFlg, 0, 1) {
		if c.conn != nil {
			_ = c.conn.Close()
			c.conn = nil
		}
		close(c.close)
	}
}

type TcpManager struct {
	mux   sync.RWMutex
	conns map[string]*Client
}

func NewTcpManager() *TcpManager {
	return &TcpManager{conns: make(map[string]*Client)}
}

func (tm *TcpManager) Pong() (uint32, EventHandle) {
	return 1001, func(msg *Message, sender MessageSender) {
		if string(msg.Body) == "ping" {
			msg.Ts = uint32(time.Now().Unix())
			msg.Body = []byte("pong")
			_ = sender.Input(msg)
		}
	}
}

func (tm *TcpManager) Send(clientId string, body []byte) error {
	tm.mux.RLock()
	defer tm.mux.RUnlock()
	if conn, ok := tm.conns[clientId]; ok {
		return conn.Input(&Message{
			Len:  uint32(17 + len(body)),
			Ver:  1,
			Cmd:  1001,
			Ts:   uint32(time.Now().Unix()),
			Seq:  0,
			Body: body,
		})
	}
	return errors.New("client not found")
}

func (tm *TcpManager) Run(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go func(conn net.Conn) {
			remoteAddr := strings.Split(conn.RemoteAddr().String(), ":")
			tm.mux.RLock()
			if conn, ok := tm.conns[remoteAddr[0]]; ok {
				conn.Close()
			}
			tm.mux.RUnlock()

			tm.mux.Lock()
			clt := NewClient(conn)
			tm.conns[remoteAddr[0]] = clt
			tm.mux.Unlock()

			// 注册自定义事件
			clt.RegEvent(tm.Pong())

			if err := clt.Loop(); err != nil {
				fmt.Println(err)
			}

			tm.mux.Lock()
			clt.Close()
			delete(tm.conns, remoteAddr[0])
			tm.mux.Unlock()
		}(conn)
	}
}

func (tm *TcpManager) Close() {
	tm.mux.RLock()
	defer tm.mux.RUnlock()
	for _, v := range tm.conns {
		v.Close()
	}
}

func main() {
	tm := NewTcpManager()
	go tm.Run("127.0.0.1:8134")
	go func() {
		time.Sleep(time.Second * 2)
		fmt.Println("发送:", tm.Send("127.0.0.1", []byte("你好")))
	}()
	go func() {
		appSignal := make(chan os.Signal, 3)
		signal.Notify(appSignal, os.Interrupt)

		go func() {
			select {
			case <-appSignal:
				tm.Close()
				fmt.Println("安全退出")
				time.Sleep(time.Second * 1)
				os.Exit(-2)
			}
		}()

	}()
	time.Sleep(time.Second * 1)

	conn, err := net.Dial("tcp", "127.0.0.1:8134")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	pingMsg := &Message{
		Len:  21,
		Ver:  1,
		Cmd:  1001,
		Ts:   uint32(time.Now().Unix()),
		Seq:  0,
		Body: []byte("ping"),
	}

	for {
		pingMsg.Ts = uint32(time.Now().Unix())
		pingMsg.Seq++
		_, err = conn.Write(EncodeMessage(pingMsg))
		if err != nil {
			fmt.Println(err)
			return
		}

		cnt, err := ReadMessage(conn)
		if err != nil {
			fmt.Println(err)
			return
		}
		pingMsg, err := DecodeMessage(cnt)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(pingMsg.Ts, pingMsg.Seq, string(pingMsg.Body))
		time.Sleep(time.Second * 2)
	}
}
