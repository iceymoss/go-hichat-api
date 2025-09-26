//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"time"
)

// 简单的测试任务
type SimpleTestTask struct{}

func (t *SimpleTestTask) GetName() string {
	return "simple_test_task"
}

func (t *SimpleTestTask) GetSpec() string {
	return "*/5 * * * * *" // 每5秒执行一次
}

func (t *SimpleTestTask) GetDescription() string {
	return "简单测试任务"
}

func (t *SimpleTestTask) GetTimeout() time.Duration {
	return 10 * time.Second
}

func (t *SimpleTestTask) Execute(ctx context.Context) error {
	fmt.Printf("Simple test task executed at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	return nil
}

func main() {
	fmt.Println("Testing cron task framework...")

	// 创建测试任务
	task := &SimpleTestTask{}

	// 模拟任务执行
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		fmt.Printf("Executing task %d...\n", i+1)
		if err := task.Execute(ctx); err != nil {
			fmt.Printf("Task execution failed: %v\n", err)
		}
		time.Sleep(2 * time.Second)
	}

	fmt.Println("Test completed successfully!")
}
