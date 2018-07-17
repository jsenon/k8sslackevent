// Copyright © 2018 Julien SENON <julien.senon@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	r "runtime"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Launch Event watcher",
	Long: `Launch event watcher
in order to get OOM signal
`,
	Run: func(cmd *cobra.Command, args []string) {
		Serve()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

// Serve launch command serve
func Serve() {
	var kubeconfig *string
	var podsStore cache.Store
	var podStorekube cache.Store
	var nodesStore cache.Store
	var eventStore cache.Store

	ctx := context.Background()

	kubeconfig = flag.String("kubeconfig", filepath.Join(homeDir(), ".kube", "config"), "(optional) absolute path to the kubeconfig file")

	client, err := Connect(kubeconfig)
	if err != nil {
		fmt.Println(err)
	}

	pods, err := client.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	nodes, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	fmt.Printf("There are %d nodes in the cluster\n", len(nodes.Items))

	_, err = getNode(ctx, client)
	if err != nil {
		panic(err.Error())
	}
	go eventPod(ctx, client, podsStore, "default")
	go eventPod(ctx, client, podStorekube, "kube-system")
	go eventNode(ctx, client, nodesStore)
	go event(ctx, client, eventStore, "default")

	fmt.Println("** Watcher started - Waiting events **")
	r.Goexit()

}

// Connect will connect to k8s cluster
func Connect(filePath *string) (clientset *kubernetes.Clientset, err error) {

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *filePath)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset, err
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func eventPod(ctx context.Context, client *kubernetes.Clientset, store cache.Store, namespace string) cache.Store {

	//Define what we want to look for (Pods)
	watchlist := cache.NewListWatchFromClient(client.CoreV1().RESTClient(), "pods", namespace, fields.Everything())
	fmt.Println("Namespace :", namespace)
	if namespace == "default" {
		fmt.Println("Namespace :", namespace)
		watchlist = cache.NewListWatchFromClient(client.CoreV1().RESTClient(), "pods", v1.NamespaceDefault, fields.Everything())
	}
	resyncPeriod := 5 * time.Minute

	//Setup an informer to call functions when the watchlist changes
	eStore, eController := cache.NewInformer(
		watchlist,
		&v1.Pod{},
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod := obj.(*v1.Pod)
				fmt.Println("Add Pod:", pod.GetName(), "on ", namespace)
				msg := "New Pod added: " + pod.GetName() + namespace
				publish(msg)
			},
			DeleteFunc: func(obj interface{}) {
				pod := obj.(*v1.Pod)
				fmt.Println("Delete Pod:", pod.GetName(), "on ", namespace)
				msg := "Deleted Pod: " + pod.GetName() + namespace
				publish(msg)
			},
		},
	)
	eController.Run(ctx.Done())
	return eStore
}

func eventNode(ctx context.Context, client *kubernetes.Clientset, store cache.Store) cache.Store {
	resyncPeriod := 30 * time.Minute

	//Setup an informer to call functions when the watchlist changes
	eStore, eController := cache.NewInformer(
		// watchlist,
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return client.CoreV1().Nodes().List(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().Nodes().Watch(lo)
			},
		},
		&v1.Node{},
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				node := obj.(*v1.Node)
				fmt.Println("New Node:", node.GetName())
				// msg := "New Node added: " + node.GetName()
				// publish(msg)
			},
			DeleteFunc: func(obj interface{}) {
				node := obj.(*v1.Node)
				fmt.Println("Deleted Node:", node)
				// msg := "Deleted Node: " + node.GetName()
				// publish(msg)
			},
			UpdateFunc: nil,
			// func(objold interface{}, objnew interface{}) {
			// 	nodeold := objold.(*v1.Node)
			// 	nodenew := objnew.(*v1.Node)
			// 	fmt.Println("Updated Node:", nodeold.GetName(), "to:", nodenew)
			// },
		},
	)
	eController.Run(ctx.Done())
	return eStore
	// ctx is not canceled, continue immediately
}

func getNode(ctx context.Context, client *kubernetes.Clientset) (cache.Store, error) {
	for {
		a, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, n := range a.Items {
			fmt.Println(n.GetName())
		}
		select {
		case <-ctx.Done():
			// ctx is canceled
			return nil, ctx.Err()
		default:
			return nil, nil
			// ctx is not canceled, continue immediately
		}
	}
}

func event(ctx context.Context, client *kubernetes.Clientset, store cache.Store, namespace string) cache.Store {

	resyncPeriod := 30 * time.Minute

	//Setup an informer to call functions when the watchlist changes
	eStore, eController := cache.NewInformer(
		// watchlist,
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return client.CoreV1().Events(namespace).List(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().Events(namespace).Watch(lo)
			},
		},
		&v1.Event{},
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				event := obj.(*v1.Event)
				fmt.Println("New Event:", event.Reason, "", event.Message, "on ", event.Name)
				fmt.Println("Debug", event)
				msg := "New Event: " + event.Reason + "\n" + event.Message
				publish(msg)
				err := findPodKilled(ctx, client, "all", 1)
				if err != nil {
					fmt.Println(err)
				}
			},
			DeleteFunc: func(obj interface{}) {
				event := obj.(*v1.Event)
				fmt.Println("Deleted event:", event.Reason, "", event.Message)
				msg := "New Event: " + event.Reason + "\n" + event.Message
				publish(msg)
				err := findPodKilled(ctx, client, "all", 1)
				if err != nil {
					fmt.Println(err)
				}
				// fmt.Println("Debug", event)
			},
			UpdateFunc: nil,
			// func(objold interface{}, objnew interface{}) {
			// 	eventold := objold.(*v1.Node)
			// 	eventnew := objnew.(*v1.Node)
			// 	fmt.Println("Updated Event:", eventold.GetName(), "to:", eventnew)
			// },
		},
	)
	eController.Run(ctx.Done())
	return eStore
	// ctx is not canceled, continue immediately
}

func publish(msg string) {
	url := os.Getenv("SLACK_URL")
	// fmt.Println("Slack url", url)

	values := map[string]string{"text": msg}
	b, _ := json.Marshal(values)
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	httpclient := &http.Client{Transport: tr}
	rs, err := httpclient.Post(url, "application/json", bytes.NewBuffer(b))
	// fmt.Println("Body", b, "rs", rs)
	if err != nil {
		panic(err)
	}
	defer rs.Body.Close() // nolint: errcheck
}

// nolint: gocyclo
func findPodKilled(ctx context.Context, client *kubernetes.Clientset, namespace string, offset uint32) error {
	if namespace == "all" {
		fmt.Println("all namespace")
		a, err := client.CoreV1().Pods(v1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		for _, n := range a.Items {
			for _, m := range n.Status.ContainerStatuses {
				if m.LastTerminationState.Terminated != nil {
					if m.LastTerminationState.Terminated.Reason == "OOMKilled" {
						fmt.Println("Pod ", n.GetName(), "Container", m.Name, "has been restarted ", m.RestartCount, "time", "due to ", m.LastTerminationState.Terminated.Reason, "at ", m.LastTerminationState.Terminated.FinishedAt)
						msg := ("Pod " + n.GetName() + "Container" + m.Name + "has been restarted " + conv(m.RestartCount) + "time" + "due to " + m.LastTerminationState.Terminated.Reason + "at " + m.LastTerminationState.Terminated.FinishedAt.String())
						publish(msg)
					} else {
						fmt.Println("Debug: No container OOMKilled")
					}
					fmt.Println("Debug: No container terminated")
				}
			}
		}
		return nil
	}

	fmt.Println("namespace: ", namespace)
	a, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, n := range a.Items {
		// fmt.Println("POD: ", n.GetName())
		for _, m := range n.Status.ContainerStatuses {
			if m.LastTerminationState.Terminated != nil {
				if m.LastTerminationState.Terminated.Reason == "OOMKilled" {
					fmt.Println("Pod ", n.GetName(), "Container", m.Name, "has been restarted ", m.RestartCount, "time", "due to ", m.LastTerminationState.Terminated.Reason, "at ", m.LastTerminationState.Terminated.FinishedAt)
				} else {
					fmt.Println("no container OOMKilled")
				}
				fmt.Println("No container terminated")
			}
		}
	}
	return nil
}

func conv(n int32) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
}
