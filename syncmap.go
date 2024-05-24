package main

import (
	"fmt"
	"sync"
)

type mapStruct struct {
	sync.Mutex
	clients map[int]*Client
}

func (m *mapStruct) Get(id int) *Client {
	m.Lock()
	defer m.Unlock()

	if v, ok := m.clients[id]; ok {
		return v
	}
	return nil
}

func (m *mapStruct) Set(id int, client *Client) {
	m.Lock()
	defer m.Unlock()
	m.clients[id] = client
	fmt.Println("Client with Id ", id, " added.")
	fmt.Println("Clients: ", m.clients)
}

func (m *mapStruct) Delete(id int) {
	m.Lock()
	defer m.Unlock()
	if client, ok := m.clients[id]; ok {
		delete(m.clients, id)
		client.conn.Close()
		fmt.Printf("Deleted: %v\n", id)
		fmt.Println("Clients: ", m.clients)
	}
}

func MakeMap() mapStruct {
	return mapStruct{
		clients: make(map[int]*Client),
	}
}
