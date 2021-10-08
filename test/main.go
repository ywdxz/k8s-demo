package main

import (
	"context"
	//"encoding/json"
	"fmt"

	"net/http"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	//"k8s.io/metrics/pkg/apis/metrics/v1alpha1"
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

	/*
		data, err := k8sClient.RESTClient().Get().AbsPath("apis/metrics.k8s.io/v1beta1/pods").DoRaw(context.Background())
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Println(string(data))

		var podMetrics v1alpha1.PodMetricsList
		json.Unmarshal(data, &podMetrics)
		for _, item := range podMetrics.Items {
			container := item.Containers[0]
			msg := fmt.Sprintf("Container Name: %s \n CPU usage: %s \n Memory usage: %d", item.Name, container.Usage.Cpu().String(), container.Usage.Memory().Value())
			fmt.Println(msg)
		}
	*/

	HTTPServe(k8sClient)
}

func HTTPServe(c *kubernetes.Clientset) {

	http.HandleFunc("/pods", func(w http.ResponseWriter, req *http.Request) {
		data, err := c.RESTClient().Get().AbsPath("apis/metrics.k8s.io/v1beta1/pods").DoRaw(context.Background())
		if err != nil {
			fmt.Fprintln(w, err.Error())
		} else {
			fmt.Fprintln(w, string(data))
		}
	})
	fmt.Println("HTTPServe start:")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err.Error())
	}
}
