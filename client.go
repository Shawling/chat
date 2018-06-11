package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type client struct {
	socket   *websocket.Conn
	send     chan *message
	room     *room
	userData map[string]interface{}
}

//读取来自 client 的消息，并加入到 room.forward 中
func (c *client) read() {
	//在函数结束时调用 socket 的 Close 方法
	defer c.socket.Close()
	for {
		msg := &message{}
		err := c.socket.ReadJSON(&msg)
		msg.Name = c.userData["name"].(string)
		if err != nil {
			log.Printf("Error reading from room %s: %s", c.userData["name"].(string), err)
			return
		}
		c.room.forward <- msg
	}
}

//向 client 写入消息
func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.send {
		if err := c.socket.WriteJSON(msg); err != nil {
			log.Printf("Error writing to room %s: %s", c.userData["name"].(string), err)
			return
		}
	}
}
