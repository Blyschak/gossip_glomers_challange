package main

import (
	"encoding/json"
	"fmt"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type State struct {
	mu            sync.Mutex
	next_local_id uint
	messages      map[uint]struct{}
	neighbors     []string
}

func NewState() State {
	return State{
		next_local_id: 0,
		messages:      make(map[uint]struct{}),
		neighbors:     make([]string, 0),
	}
}

func (s *State) HandleEcho(n *maelstrom.Node, msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	body["type"] = "echo_ok"

	return n.Reply(msg, body)
}

func (s *State) HandleGenerate(n *maelstrom.Node, msg maelstrom.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	body["type"] = "generate_ok"
	body["id"] = s.generateIdGlobal(n)

	return n.Reply(msg, body)
}

func (s *State) HandleBroadcast(n *maelstrom.Node, msg maelstrom.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	var m, ok = body["message"].(float64)
	if !ok {
		return fmt.Errorf("%v is not valid", body["message"])
	}

	var message = uint(m)

	if _, ok = s.messages[message]; !ok {
		s.messages[uint(message)] = struct{}{}
		for i := range s.neighbors {
			n.RPC(s.neighbors[i], body, func(msg maelstrom.Message) error { return nil })
		}
	}

	var response = make(map[string]any)
	response["type"] = "broadcast_ok"

	return n.Reply(msg, response)
}

func (s *State) HandleRead(n *maelstrom.Node, msg maelstrom.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	messages := make([]uint, 0, len(s.messages))
	for message := range s.messages {
		messages = append(messages, message)
	}

	body["messages"] = messages
	body["type"] = "read_ok"

	return n.Reply(msg, body)
}

func (s *State) HandleTopology(n *maelstrom.Node, msg maelstrom.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	topology, ok := body["topology"].(map[string]any)
	if !ok {
		return fmt.Errorf("Topology is not of expected type %v", body["topology"])
	}

	neighbors, ok := topology[n.ID()].([]interface{})
	if !ok {
		return fmt.Errorf("Missing info about nodes %v neighbors %v (%T)", n.ID(), topology[n.ID()], topology[n.ID()])
	}

	for i := range neighbors {
		s.neighbors = append(s.neighbors, neighbors[i].(string))
	}

	var response = make(map[string]any)
	response["type"] = "topology_ok"

	return n.Reply(msg, response)
}

func (s *State) generateIdLocal() uint {
	var id = s.next_local_id
	s.next_local_id += 1
	return id
}

func (s *State) generateIdGlobal(n *maelstrom.Node) string {
	return fmt.Sprintf("%s_%d", n.ID(), s.generateIdLocal())
}
