// Copyright (c) 2023, NVIDIA
// All rights reserved//
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
	"strings"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetNodes(c *Config) ([]Node, error) {
	var nodes []Node
	var nodeNames []string

	if c.Wkload != "None" {
		wkload := strings.Split(c.Wkload, ":")
		if len(wkload) != 2 || wkload[0] == "" || wkload[1] == "" {
			return nil, fmt.Errorf("invalid job definition, must be namespace:label=value. Got %s", c.Wkload)
		}
		wkloadNs := wkload[0]
		wkloadLabel := wkload[1]
		pods, err := c.clientset.CoreV1().Pods(wkloadNs).List(context.Background(), metav1.ListOptions{LabelSelector: wkloadLabel})
		if err != nil {
			return nil, fmt.Errorf("exception when calling CoreV1Api->list_namespaced_pod: %w", err)
		}
		if len(pods.Items) == 0 {
			return nil, fmt.Errorf("no workload labeled with %s - ABORT", wkloadLabel)
		}
		for _, pod := range pods.Items {
			nodeNames = append(nodeNames, pod.Spec.NodeName)
		}
	}

	if c.NodeLabel != "None" {
		labeledNodes, err := c.clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{LabelSelector: c.NodeLabel})
		if err != nil {
			return nil, fmt.Errorf("exception when calling CoreV1Api->list_node: %w", err)
		}
		if len(labeledNodes.Items) == 0 {
			return nil, fmt.Errorf("no node is labeled with %s - ABORT", c.NodeLabel)
		}
		for _, labeledNode := range labeledNodes.Items {
			nodeNames = append(nodeNames, labeledNode.Metadata.Name)
		}
	}

	if c.Nodes != "all" {
		nodeNames = append(nodeNames, strings.Split(c.Nodes, ",")...)
	}

	endpoints, err := c.clientset.CoreV1().Endpoints(c.Namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("exception when calling CoreV1Api->list_namespaced_endpoints: %w", err)
	}

	for _, endpointslice := range endpoints.Items {
		if endpointslice.ObjectMeta.Name == c.Service {
			for _, subset := range endpointslice.Subsets {
				for _, address := range subset.Addresses {
					if len(nodeNames) == 0 || contains(nodeNames, *address.NodeName) {
						nodes = append(nodes, Node{
							Name: *address.NodeName,
							IP:   address.IP,
						})
					}
				}
			}
		}
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes found")
	}

	return nodes, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func createURL(address string, daemonNode string, checks string, dcgmR string, wkload string, nodelabel string, nodes string) string {
	if checks == "all" {
		return fmt.Sprintf("http://%s:3333/status?host=%s", address, daemonNode)
	}
	extraParams := ""
	if strings.Contains(checks, "ping") {
		if wkload != "None" {
			extraParams += "&job=" + wkload
		}
		if nodelabel != "None" {
			extraParams += "&nodelabel=" + nodelabel
		}
		if nodes != "all" {
			extraParams += "&pingnodes=" + nodes
		}
	}
	if strings.Contains(checks, "dcgm") {
		extraParams += "&r=" + dcgmR
	}
	return fmt.Sprintf("http://%s:3333/status?host=%s&check=%s%s", address, daemonNode, checks, extraParams)
}

func getNodeStatus(response string) []string {
	var nodeStatusList []string
	responseList := strings.Split(response, "\n")
	for _, line := range responseList {
		if strings.Contains(line, "FAIL") || strings.Contains(line, "ABORT") {
			if strings.Contains(line, "PCIE") {
				nodeStatusList = append(nodeStatusList, "PCIE Failed")
			} else if strings.Contains(line, "REMAPPED ROWS") {
				nodeStatusList = append(nodeStatusList, "REMAPPED ROWS Failed")
			} else if strings.Contains(line, "DCGM") {
				nodeStatusList = append(nodeStatusList, "DCGM Failed")
			} else if strings.Contains(line, "GPU POWER") {
				nodeStatusList = append(nodeStatusList, "GPU POWER Failed")
			} else if strings.Contains(line, "PING") {
				nodeStatusList = append(nodeStatusList, "PING Failed")
			} else if strings.Contains(line, "GPU-MEM") {
				nodeStatusList = append(nodeStatusList, "GPU MEM Test Failed")
			} else if strings.Contains(line, "PVC") {
				nodeStatusList = append(nodeStatusList, "PVC Create-Delete Test Failed")
			} else if strings.Contains(line, "Disconnected") {
				nodeStatusList = append(nodeStatusList, "Connection to Server Failed")
			}
		}
	}
	if len(nodeStatusList) < 1 {
		nodeStatusList = append(nodeStatusList, "OK")
	}
	return nodeStatusList
}