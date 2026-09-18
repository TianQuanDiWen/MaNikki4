package main

import (
	"fmt"
	"os"

	"github.com/TianQuanDiWen/MaNikki4/agent/internal/agentserver"
	"github.com/TianQuanDiWen/MaNikki4/agent/internal/launcher"
)

// main 将命令行参数交给模式分发器，并将错误写入标准错误后返回非零退出码。
func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "MaNikki4 Agent failed: %v\n", err)
		os.Exit(1)
	}
}

// run 根据第一个参数选择运行模式。
func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing mode: expected agent or pretask")
	}

	switch args[0] {
	case "agent":
		return agentserver.Run(args[1:])
	case "pretask":
		return launcher.Run(args[1:])
	default:
		return fmt.Errorf("unknown mode %q: expected agent or pretask", args[0])
	}
}
