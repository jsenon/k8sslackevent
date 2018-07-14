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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
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

	watchlist := cache.NewListWatchFromClient(client.Core().RESTClient(), "pods", v1.NamespaceDefault,
		fields.Everything())
	_, controller := cache.NewInformer(
		watchlist,
		&v1.Pod{},
		time.Second*10,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod := obj.(*v1.Pod)
				fmt.Println("add Pod:", pod.GetName())
			},
			DeleteFunc: func(obj interface{}) {
				pod := obj.(*v1.Pod)
				fmt.Printf("delete Pod: %s \n", pod.GetName())
			},
			// UpdateFunc: func(oldObj, newObj interface{}) {
			// 	podold := oldObj.(*v1.Pod)
			// 	podnew := newObj.(*v1.Pod)
			// 	fmt.Printf("old: %s, new: %s \n", podold.GetName(), podnew.GetName())
			// },
		},
	)
	stop := make(chan struct{})
	go controller.Run(stop)

	// watchList2 := cache.NewListWatchFromClient(client.Core().RESTClient(), "nodes", v1.NamespaceDefault,
	// 	fields.Everything())
	// _, controller2 := cache.NewInformer(
	// 	watchList2,
	// 	&v1.Node{},
	// 	time.Second*30,
	// 	cache.ResourceEventHandlerFuncs{
	// 		AddFunc: func(obj interface{}) {
	// 			fmt.Printf("add: %s \n", obj)
	// 		},
	// 		UpdateFunc: func(oldObj, newObj interface{}) {
	// 			fmt.Printf("old: %s, new: %s \n", oldObj, newObj)
	// 		},
	// 	},
	// )

	// stop2 := make(chan struct{})
	// go controller2.Run(stop2)

	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	fmt.Printf("There are %d nodes in the cluster\n", len(nodes.Items))
	fmt.Println("** Waiting event **")
	for {
		time.Sleep(time.Second)
	}

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
