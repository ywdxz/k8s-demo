package main

import (
	"context"
	"fmt"

	"net/http"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	var kubeconfig, master string //empty, assuming inClusterConfig
	config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		panic(err)
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// var podMetrics v1alpha1.PodMetricsList
	data, err := k8sClient.RESTClient().Get().AbsPath("apis/metrics.k8s.io/v1beta1/pods").DoRaw(context.Background())
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(string(data))
	HTTPServe(string(data))

	// json.Unmarshal(data, &podMetrics)
	// for _, item := range podMetrics.Items {
	// 	container := item.Containers[0]
	// 	msg := fmt.Sprintf("Container Name: %s \n CPU usage: %s \n Memory usage: %d", item.Name, container.Usage.Cpu().String(), container.Usage.Memory().Value())
	// 	fmt.Println(msg)
	// }
}

func HTTPServe(data string) {

	http.HandleFunc("/list", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, data)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err.Error())
	}
}
