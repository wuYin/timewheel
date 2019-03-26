# timewheel

Golang 实现的时间轮算法

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

## 场景：定时保活

在 [wuYin/tron](https://github.com/wuYin/tron) 网络框架中，一个 Server 端需对已连接的多个 Client 定时发送 Ping 心跳包，若在超时时间内收到 Pong 包则认为连接有效，若未收到则二次规避重试一定次数后主动断开连接。实现方案：

- 简单实现：为每个连接会话都分配一个 `Ticker` 定时保活，但连接过多后会占用 Server 过多内存资源
- 时间轮实现：为每个 Server 配置一个时间轮，将保活任务作为指定次数的重复任务统一管理，安全高效

## 误差

`time.Ticker ` 的粒度为 1ns，时间轮的粒度由用户指定。当新增任务时可能还未开始下次 tick，当 tick 粒度较大如 1s 时，任务执行时间将出现 `[0,1)s` 的明显误差。

为减少误差，tick 粒度向下取重复间隔 2 个量级较好。但粒度越细，时间轮占用 CPU 的频率越高，需做好权衡。