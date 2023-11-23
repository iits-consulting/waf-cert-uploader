package main

import (
	"context"
	"fmt"
	"github.com/thoas/go-funk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func test(clientSet *kubernetes.Clientset) {

	pods, err := clientSet.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		panic(err.Error())
	}

	secret, err := clientSet.CoreV1().Secrets("default").Get(context.TODO(), "otc-credentials", metav1.GetOptions{})

	if err != nil {
		panic(err.Error())
	}

	fmt.Println(secret.Data)

	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	funk.ForEach(secret.Data, func(k string, v []byte) {
		fmt.Println(k + ":" + string(v))
	})
}
