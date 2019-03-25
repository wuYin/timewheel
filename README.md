# timewheel

Golang 实现的时间轮算法

## 场景

在 [wuYin/tron](https://github.com/wuYin/tron) 网络框架中，一个 Server 端需对已连接的多个 Client 定时发送 Ping 心跳包，若在超时时间内收到 Pong 包则认为连接有效，若未收到则二次规避重试一定次数后主动断开连接。实现方案：

- 简单实现：为每个连接会话都分配一个 `Ticker` 定时保活，但连接过多后可能占用 Server 过多内存资源
- 时间轮实现：为 Server 配置指定粒度的时间轮，将所有保活 `Ticker` 放入对应轮槽后统一管理，安全高效

## 功能

- 执行定时任务
- 执行指定次数的重复任务
- 任务中断和更新

## 使用

```go
package main

import (
	"fmt"
	"github.com/wuYin/timewheel"
	"time"
)

func main() {
	tw := timewheel.NewTimeWheel(100*time.Millisecond, 600) // 周期为一分钟

	// 执行定时任务
	tid, _ := tw.After(5*time.Second, func() {
		fmt.Println("after 5 seconds, task1 executed")
	})

	// 执行指定次数的重复任务
	_, allDone := tw.Repeat(1*time.Second, 3, func() {
		fmt.Println("per 1 second, task2 executed")
	})
	<-allDone

	// 中途取消任务
	tw.Cancel(tid)
}

```



## 原理

使用双向链表存储提交的 **Task**，当 **Ticker** 扫到当前 **Slot** 后，将符合条件的 **Task** 放到新 goroutine 执行即可。

 <img src="https://images.yinzige.com/2019-03-15-tw.jpg"/>