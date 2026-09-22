package agentserver

import (
	"fmt"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/TianQuanDiWen/MaNikki4/agent/internal/arena"
	"github.com/TianQuanDiWen/MaNikki4/agent/internal/emulator"
	"github.com/TianQuanDiWen/MaNikki4/agent/internal/matcher"
)

// Registry 保存项目提供给 MaaFramework 的自定义识别和自定义动作。
type Registry struct {
	recognitions map[string]maa.CustomRecognitionRunner
	actions      map[string]maa.CustomActionRunner
}

// NewRegistry 创建一个空的扩展注册表。
func NewRegistry() *Registry {
	return &Registry{
		recognitions: make(map[string]maa.CustomRecognitionRunner),
		actions:      make(map[string]maa.CustomActionRunner),
	}
}

// AddRecognition 添加一个 MaaFramework 自定义识别实现。
func (r *Registry) AddRecognition(name string, runner maa.CustomRecognitionRunner) error {
	if name == "" || runner == nil {
		return fmt.Errorf("custom recognition requires a name and runner")
	}
	if _, exists := r.recognitions[name]; exists {
		return fmt.Errorf("custom recognition %q is already registered", name)
	}
	r.recognitions[name] = runner
	return nil
}

// AddAction 添加一个 MaaFramework 自定义动作实现。
func (r *Registry) AddAction(name string, runner maa.CustomActionRunner) error {
	if name == "" || runner == nil {
		return fmt.Errorf("custom action requires a name and runner")
	}
	if _, exists := r.actions[name]; exists {
		return fmt.Errorf("custom action %q is already registered", name)
	}
	r.actions[name] = runner
	return nil
}

// RegisterAgentServer 将当前注册表一次性挂载到 MaaFramework AgentServer。
func (r *Registry) RegisterAgentServer() error {
	for name, runner := range r.recognitions {
		if err := maa.AgentServerRegisterCustomRecognition(name, runner); err != nil {
			return fmt.Errorf("register custom recognition %q: %w", name, err)
		}
	}
	for name, runner := range r.actions {
		if err := maa.AgentServerRegisterCustomAction(name, runner); err != nil {
			return fmt.Errorf("register custom action %q: %w", name, err)
		}
	}
	return nil
}

// BuildRegistry 是项目自定义能力的统一装配入口。
func BuildRegistry(projectRoot string) (*Registry, error) {
	registry := NewRegistry()

	checker := arena.NewPowerChecker()

	// 注册自身战力记录识别器（进入竞技场时执行一次，暂存自身战力）
	recordRunner := arena.NewArenaRecordSelfPowerRunner(checker)
	if err := registry.AddRecognition("ArenaRecordSelfPower", recordRunner); err != nil {
		return nil, fmt.Errorf("register ArenaRecordSelfPower: %w", err)
	}

	// 注册通用竞技场战力比对条件识别器（区域完全由节点自身的 ROI 决定）
	runner := arena.NewArenaCanBeatRunner(checker)
	if err := registry.AddRecognition("ArenaCanBeat", runner); err != nil {
		return nil, fmt.Errorf("register ArenaCanBeat: %w", err)
	}

	// 注册通用行按钮关联识别器（根据 keyword 锁定同行右侧按钮，支持 | 管道符）
	if err := registry.AddRecognition("MatchRowButton", matcher.NewMatchRowButtonRunner()); err != nil {
		return nil, fmt.Errorf("register MatchRowButton: %w", err)
	}

	// 注册退出模拟器自定义动作
	shutdownRunner := maa.CustomActionFunc(func(ctx *maa.Context, arg *maa.CustomActionArg) bool {
		fmt.Println("[Agent] 收到退出模拟器指令，正在安全关闭...")
		if err := emulator.ShutdownEmulator(projectRoot); err != nil {
			fmt.Printf("[Agent] 关闭模拟器失败: %v\n", err)
			return false
		}
		return true
	})
	if err := registry.AddAction("ShutdownEmulator", shutdownRunner); err != nil {
		return nil, fmt.Errorf("register ShutdownEmulator: %w", err)
	}

	return registry, nil
}
