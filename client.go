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
		if err != nil {
			log.Printf("Error reading from room %s: %s", c.userData["name"].(string), err)
			return
		}
		msg.Name = c.userData["name"].(string)
		if msg.AvatarURL, err = c.room.avatar.GetAvatarURL(c); err != nil {
			msg.AvatarURL = "http://img.hb.aicdn.com/99eab0f202688dbe7dedd09dfc69cace1201cf851f206-tlBb0W_fw658"
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
