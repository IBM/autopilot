// Copyright (c) 2023, NVIDIA
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its
//    contributors may be used to endorse or promote products derived from
//    this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package healthcheck

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"autopilot/pkg/utils"

	"github.com/NVIDIA/go-dcgm/pkg/dcgm"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func RunHealthChecks(c *cli.Context) error {
	config, err := NewConfig(c)
	if err != nil {
		return err
	}

	nodes, err := GetNodes(config)
	if err != nil {
		config.log.Fatalf("Error getting nodes: %v", err)
	}

	var wg sync.WaitGroup
	nodeStatus := make(map[string][]string)
	var mu sync.Mutex

	batchSize := config.BatchSize
	if batchSize == 0 || batchSize > len(nodes) {
		batchSize = len(nodes)
	}

	for i := 0; i < len(nodes); i += batchSize {
		end := i + batchSize
		if end > len(nodes) {
			end = len(nodes)
		}
		batch := nodes[i:end]

		for _, node := range batch {
			wg.Add(1)
			go func(node Node) {
				defer wg.Done()
				url := createURL(node.IP, node.Name, config.Checks, config.DcgmR, config.Wkload, config.NodeLabel, config.Nodes)
				config.log.Infof("Initiated connection to %s.", url)

				client := &http.Client{Timeout: 24 * time.Hour}
				resp, err := client.Get(url)
				var reply string
				if err != nil {
					config.log.Errorf("Error connecting to server: %v", err)
					reply = "Server Disconnected. ABORT"
				} else {
					defer resp.Body.Close()
					body, err := utils.ReadAll(resp.Body)
					if err != nil {
						config.log.Errorf("Error reading response body: %v", err)
						reply = "Error reading response body. ABORT"
					} else {
						reply = string(body)
					}
				}

				status := getNodeStatus(reply)
				mu.Lock()
				nodeStatus[node.Name] = status
				mu.Unlock()

				config.log.Infof("\nAutopilot Endpoint: %s\nNode: %s\nurl(s): %s\nResponse:\n%s\nNode Status: %s\n-------------------------------------\n", node.IP, node.Name, url, reply, strings.Join(status, ", "))
			}(node)
		}
		wg.Wait()
	}

	config.log.Info("Node Summary:\n")
	for node, status := range nodeStatus {
		config.log.Infof("%s: %s", node, strings.Join(status, ", "))
	}

	return nil
}

func NewConfig(c *cli.Context) (*Config, error) {
	config := &Config{
		Service:   c.String("service"),
		Namespace: c.String("namespace"),
		Nodes:     c.String("nodes"),
		Checks:    c.String("check"),
		BatchSize: c.Int("batchSize"),
		Wkload:    c.String("wkload"),
		DcgmR:     c.String("dcgmR"),
		NodeLabel: c.String("nodelabel"),
		log:       logrus.NewEntry(logrus.New()),
	}

	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting in-cluster config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating clientset: %w", err)
	}
	config.clientset = clientset

	return config, nil
}