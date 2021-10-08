package main

import (
	"container/heap"
	"context"
	"encoding/json"
	"fmt"

	"net/http"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/apis/metrics/v1alpha1"
)

func main() {

	var kubeconfig, master string
	config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		panic(err)
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	HTTPServe(k8sClient)
}

func HTTPServe(c *kubernetes.Clientset) {

	http.HandleFunc("/pods", func(w http.ResponseWriter, req *http.Request) {
		data, err := c.RESTClient().Get().AbsPath("apis/metrics.k8s.io/v1beta1/pods").DoRaw(context.Background())
		if err != nil {
			fmt.Fprintln(w, err.Error())
			return
		}

		msg, err := ProcessingLogic(data)
		if err != nil {
			fmt.Fprintln(w, err.Error())
			return
		}

		b, err := json.Marshal(msg)
		if err != nil {
			fmt.Fprintln(w, err.Error())
			return
		}

		fmt.Fprintln(w, string(b))
	})

	fmt.Println("--------HTTPServe start--------")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err.Error())
	}
}

type PodCpuInfo struct {
	Name string `json:"name"`
	CPU  string `json:"cpu"`
}

func ProcessingLogic(raw []byte) (msg []PodCpuInfo, err error) {

	var podMetrics v1alpha1.PodMetricsList
	if err = json.Unmarshal(raw, &podMetrics); err != nil {
		return
	}

	h := &Heap{}
	for _, item := range podMetrics.Items {
		// podNmae := item.ObjectMeta.Name
		container := item.Containers[0]
		t := PodCpuInfo{
			Name: item.Name,
			CPU:  container.Usage.Cpu().String(),
		}

		switch {
		case h.Len() < 10:
			heap.Push(h, t)
		case (*h)[0].CPU < t.CPU:
			heap.Pop(h)
			heap.Push(h, t)
		}
	}

	for h.Len() > 0 { // 持续推出顶部最小元素
		msg = append(msg, heap.Pop(h).(PodCpuInfo))
	}

	return
}

type Heap []PodCpuInfo

func (h Heap) Len() int { return len(h) }

func (h Heap) Less(i, j int) bool { return h[i].CPU < h[j].CPU }

func (h Heap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *Heap) Pop() interface{} {

	old := *h
	n := len(old)

	x := old[n-1]
	*h = old[0 : n-1]

	return x
}

func (h *Heap) Push(x interface{}) {

	*h = append(*h, x.(PodCpuInfo))
}
