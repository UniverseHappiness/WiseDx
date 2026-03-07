// Package main provides a tool to visualize the dependency injection graph
// Usage: go run cmd/visualize/main.go [output_file]
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/dig"

	"github.com/UniverseHappiness/WiseDx/internal/container"
	"github.com/UniverseHappiness/WiseDx/internal/runtime"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v (将使用系统环境变量)", err)
	}

	// 确定输出文件名
	output := "dependency_graph.dot"
	if len(os.Args) > 1 {
		output = os.Args[1]
	}

	// 构建容器
	log.Println("正在构建依赖注入容器...")
	c := container.BuildContainer(runtime.GetContainer())

	// 创建输出文件
	f, err := os.Create(output)
	if err != nil {
		log.Fatalf("无法创建输出文件: %v", err)
	}
	defer f.Close()

	// 输出依赖图
	log.Println("正在生成依赖图...")
	if err := dig.Visualize(c, f); err != nil {
		log.Fatalf("生成依赖图失败: %v", err)
	}

	fmt.Printf("依赖图已输出到: %s\n", output)
	fmt.Println("使用以下命令转换为图片:")
	fmt.Printf("  dot -Tpng %s -o dependency_graph.png\n", output)
	fmt.Printf("  dot -Tsvg %s -o dependency_graph.svg\n", output)
}
