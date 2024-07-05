/**
 *
 */
package main

import (
	"context"
	"fmt"
	"time"
)

func someHandler() {
	// 创建继承Background的子节点Context
	//ctx, cancel := context.WithCancel(context.Background())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	go doSth(ctx)

	//模拟程序运行 - Sleep 5秒
	time.Sleep(5 * time.Second)
	cancel()
	fmt.Println("Father call cancel")
	time.Sleep(20 * time.Second)
}

// 每1秒work一下，同时会判断ctx是否被取消，如果是就退出
func doSth(ctx context.Context) {
	var i = 1
	for {
		time.Sleep(1 * time.Second) //！！！模拟实际干活的过程！！！
		select {
		case <-ctx.Done():
			fmt.Println("done")
			return
		default:
			fmt.Printf("work %d seconds: \n", i)
		}
		i++
	}
}

func main() {
	fmt.Println("start...")
	someHandler()
	fmt.Println("end.")
}
