// Copyright 2019 Orange
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// 	You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// 	See the License for the specific language governing permissions and
// limitations under the License.
package models

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Client is the k8s client to use to connect to each kubernetes
type Client struct {
	Name   string
	Client client.Client
}

// Clients defined each client (master & remotes) to access to k8s cluster.
type Clients struct {
	Local  *Client
	Remotes []*Client
}

// Convert master & remotes clients to a merged list.
// Simplify loop on clients
func (clients *Clients) FlatClients() []*Client {
	flatClients := append([]*Client{clients.Local}, clients.Remotes...)
	return flatClients
}
