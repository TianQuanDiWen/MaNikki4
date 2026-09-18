package launcher

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/TianQuanDiWen/MaNikki4/agent/internal/runtimepath"
)

const (
	defaultPackageName = "com.papegames.nn4.cn"
	defaultADBAddress  = "127.0.0.1:16384"
)

var knownExecutables = []string{
	"MuMuNxMain.exe",
	"MuMuPlayer.exe",
	"MuMuManager.exe",
	"NemuPlayer.exe",
}

// Run 启动供 MXU Controller 前置执行的预任务。
func Run(args []string) error {
	flags := flag.NewFlagSet("pretask", flag.ContinueOnError)
	root := flags.String("root", ".", "project root containing maafw and resource")
	if err := flags.Parse(args); err != nil {
		return err
	}

	paths, err := runtimepath.Resolve(*root)
	if err != nil {
		return fmt.Errorf("resolve runtime path: %w", err)
	}

	// 1. 读取既有配置中定义的 ADB 端口、路径与实例序号
	cfgADBAddress, cfgADBPath, vmIndex := loadADBConfig(paths.Root)
	adbAddress := defaultADBAddress
	if cfgADBAddress != "" {
		adbAddress = cfgADBAddress
	}

	// 2. 解析 MXU 传入的 option 参数
	var inputPath string
	remaining := flags.Args()
	if len(remaining) > 0 {
		inputPath = extractMuMuPathFromJSON(remaining[len(remaining)-1])
	}

	// 3. 定位 MuMu 路径（显式输入 -> 正在运行 -> 注册表关联 -> 默认目录 -> 置顶弹窗）
	mumuPath, wasAutoDetected, err := resolveMuMuPath(inputPath)
	if err != nil {
		return fmt.Errorf("定位 MuMu 模拟器失败: %w", err)
	}
	fmt.Printf("[Launcher] 确认 MuMu 路径: %s\n", mumuPath)

	// 4. 自动检测或弹窗选择时写回配置供 UI 回显
	if wasAutoDetected || inputPath == "" {
		if err := savePathToConfig(paths.Root, mumuPath); err != nil {
			fmt.Printf("[Launcher] 警告: 写回配置文件失败: %v\n", err)
		} else {
			fmt.Println("[Launcher] 模拟器路径已写入 MXU 配置文件")
		}
	}

	// 5. 定位 adb.exe
	adbExe := resolveADBPath(paths.Lib, mumuPath, cfgADBPath)
	if adbExe == "" {
		return errors.New("未找到可用的 adb.exe，请检查环境")
	}

	// 6. 确保模拟器已启动并且 Android 系统完全就绪
	if err := ensureEmulatorRunning(mumuPath, adbExe, adbAddress, vmIndex); err != nil {
		return fmt.Errorf("启动/连接模拟器失败: %w", err)
	}

	// 7. 拉起闪耀暖暖并确保其运行就绪
	if err := launchGameApp(mumuPath, adbExe, adbAddress, vmIndex); err != nil {
		return fmt.Errorf("拉起游戏应用失败: %w", err)
	}

	fmt.Println("[Launcher] 启动准备完成，无缝移交 Controller")
	return nil
}

// extractMuMuPathFromJSON 从 MXU 传入的 JSON 字符串中提取 mumu_path
func extractMuMuPathFromJSON(rawJSON string) string {
	if !strings.HasPrefix(strings.TrimSpace(rawJSON), "{") {
		return ""
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(rawJSON), &data); err != nil {
		return ""
	}
	if val, ok := data["mumu_path"].(string); ok && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	if sub, ok := data["MuMuConfig"].(map[string]any); ok {
		if val, ok := sub["mumu_path"].(string); ok && strings.TrimSpace(val) != "" {
			return strings.TrimSpace(val)
		}
	}
	return ""
}

// resolveMuMuPath 通用解析 MuMu 可执行文件绝对路径
func resolveMuMuPath(inputPath string) (string, bool, error) {
	if inputPath != "" {
		clean := filepath.Clean(inputPath)
		if fileExists(clean) {
			return clean, false, nil
		}
		if exe := searchKnownExeInDir(clean); exe != "" {
			return exe, false, nil
		}
	}

	if p := findPathFromRunningProcess(); p != "" {
		return p, true, nil
	}
	if p := findPathFromRegistry(); p != "" {
		return p, true, nil
	}
	if p := findPathFromDefaultProgramFiles(); p != "" {
		return p, true, nil
	}

	fmt.Println("[Launcher] 未自动检测到默认安装路径，正在呼出置顶文件选择框...")
	selected, err := openFileDialog()
	if err != nil || selected == "" {
		return "", false, errors.New("未选择模拟器路径")
	}
	return selected, true, nil
}

// findPathFromRunningProcess 查询系统中正在运行的 MuMu 进程路径
func findPathFromRunningProcess() string {
	psCmd := `Get-Process | Where-Object { $_.ProcessName -match '^(MuMuNxMain|MuMuPlayer|MuMuManager|NemuPlayer)$' } | Select-Object -ExpandProperty Path -First 1`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	if out, err := cmd.Output(); err == nil {
		p := strings.TrimSpace(string(out))
		if p != "" && fileExists(p) {
			return p
		}
	}
	return ""
}

// findPathFromRegistry 查询 Windows 注册表中的关联指令与安装目录
func findPathFromRegistry() string {
	assocKeys := []string{
		`HKLM\SOFTWARE\Classes\MuMuPlayer.apk\shell\open\command`,
		`HKCU\Software\Classes\MuMuPlayer.apk\shell\open\command`,
		`HKLM\SOFTWARE\Classes\Applications\Nemux.exe\shell\open\command`,
		`HKCU\Software\Classes\Applications\Nemux.exe\shell\open\command`,
	}
	for _, key := range assocKeys {
		if out, err := exec.Command("reg", "query", key, "/ve").Output(); err == nil {
			if exe := extractExeFromCmd(string(out)); exe != "" {
				return exe
			}
		}
	}

	regKeys := []string{
		`HKLM\SOFTWARE\NetEase\MuMuPlayer-12.0`,
		`HKCU\Software\NetEase\MuMuPlayer-12.0`,
		`HKLM\SOFTWARE\NetEase\MuMuPlayer`,
		`HKCU\Software\NetEase\MuMuPlayer`,
	}
	for _, key := range regKeys {
		for _, v := range []string{"InstallDir", "AppPath"} {
			if out, err := exec.Command("reg", "query", key, "/v", v).Output(); err == nil {
				val := parseRegValue(string(out))
				if fileExists(val) {
					return val
				}
				if exe := searchKnownExeInDir(val); exe != "" {
					return exe
				}
			}
		}
	}
	return ""
}

// extractExeFromCmd 从注册表命令串（如 "C:\path\app.exe" -i "%1"）提取有效 exe
func extractExeFromCmd(output string) string {
	val := parseRegValue(output)
	if val == "" {
		return ""
	}
	val = strings.Trim(val, `"`)
	lower := strings.ToLower(val)
	if idx := strings.Index(lower, ".exe"); idx != -1 {
		candidate := strings.Trim(val[:idx+4], `"`)
		if fileExists(candidate) {
			return candidate
		}
	}
	return ""
}

func parseRegValue(output string) string {
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && (fields[1] == "REG_SZ" || fields[1] == "REG_EXPAND_SZ") {
			return strings.Join(fields[2:], " ")
		}
	}
	return ""
}

// findPathFromDefaultProgramFiles 检查标准 ProgramFiles 路径
func findPathFromDefaultProgramFiles() string {
	progRoots := []string{os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)"), os.Getenv("ProgramW6432")}
	subDirs := []string{`Netease\MuMuPlayer-12.0`, `Netease\MuMu Player 12`, `Netease\MuMuPlayer`, `MuMu\emulator\nemu9`}
	for _, root := range progRoots {
		if root == "" {
			continue
		}
		for _, sub := range subDirs {
			if exe := searchKnownExeInDir(filepath.Join(root, sub)); exe != "" {
				return exe
			}
		}
	}
	return ""
}

// searchFileInDirs 在 baseDir 及其各级子目录/相对路径中探测目标文件
func searchFileInDirs(baseDir, fileName string, relativeDirs ...string) string {
	for _, rel := range append([]string{""}, relativeDirs...) {
		p := filepath.Clean(filepath.Join(baseDir, rel, fileName))
		if fileExists(p) {
			return p
		}
	}
	return ""
}

func searchKnownExeInDir(dir string) string {
	for _, name := range knownExecutables {
		if p := searchFileInDirs(dir, name, "nx_main", "shell", "EmulatorShell"); p != "" {
			return p
		}
	}
	return ""
}

func findMuMuCli(mumuPath string) string {
	return searchFileInDirs(filepath.Dir(mumuPath), "mumu-cli.exe", "..", "nx_main", "../nx_main", "../../nx_main")
}

func resolveADBPath(libDir, mumuPath, cfgADBPath string) string {
	if cfgADBPath != "" && fileExists(cfgADBPath) {
		return cfgADBPath
	}
	if p := searchFileInDirs(filepath.Dir(mumuPath), "adb.exe", "nx_main", "shell", "../nx_main", "../shell"); p != "" {
		return p
	}
	if p := filepath.Join(libDir, "adb.exe"); fileExists(p) {
		return p
	}
	if p, err := exec.LookPath("adb.exe"); err == nil {
		return p
	}
	return ""
}

// openFileDialog 借助轻量 PowerShell WinForms 呼出原生置顶文件选择框（TopMost=true）
func openFileDialog() (string, error) {
	psCmd := `Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.OpenFileDialog; $f.Title = '请选择 MuMu 模拟器主程序 (如 MuMuNxMain.exe 或 MuMuPlayer.exe)'; $f.Filter = 'MuMu可执行程序 (*.exe)|MuMu*.exe;Nemu*.exe;*.exe|所有文件 (*.*)|*.*'; $t = New-Object System.Windows.Forms.Form; $t.TopMost = $true; if ($f.ShowDialog($t) -eq [System.Windows.Forms.DialogResult]::OK) { Write-Output $f.FileName }`
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	res := strings.TrimSpace(string(out))
	if res == "" {
		return "", errors.New("用户取消选择模拟器路径")
	}
	return res, nil
}

// ensureEmulatorRunning 确保模拟器启动并在指定端口就绪（严格配置 cmd.Dir 避免闪退）
func ensureEmulatorRunning(mumuPath, adbExe, adbAddress string, vmIndex int) error {
	_ = exec.Command(adbExe, "connect", adbAddress).Run()
	if isEmulatorReady(adbExe, adbAddress) {
		fmt.Println("[Launcher] 检测到 MuMu 模拟器已处于运行就绪状态")
		return nil
	}

	fmt.Println("[Launcher] 正在启动 MuMu 模拟器...")
	cliExe := findMuMuCli(mumuPath)
	if cliExe != "" {
		cmd := exec.Command(cliExe, "control", "-v", fmt.Sprintf("%d", vmIndex), "launch")
		cmd.Dir = filepath.Dir(cliExe)
		_ = cmd.Start()
	} else {
		cmd := exec.Command(mumuPath)
		cmd.Dir = filepath.Dir(mumuPath)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("启动进程失败: %w", err)
		}
	}

	fmt.Printf("[Launcher] 等待模拟器系统与 ADB 就绪 (%s)...\n", adbAddress)
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(1500 * time.Millisecond)
		_ = exec.Command(adbExe, "connect", adbAddress).Run()
		if isEmulatorReady(adbExe, adbAddress) {
			fmt.Println("[Launcher] MuMu 模拟器系统已完全就绪！")
			return nil
		}
	}
	return fmt.Errorf("等待 MuMu 模拟器启动就绪超时 (60s, 目标: %s)", adbAddress)
}

func isEmulatorReady(adbExe, adbAddress string) bool {
	out, err := exec.Command(adbExe, "-s", adbAddress, "get-state").CombinedOutput()
	if err != nil || strings.TrimSpace(string(out)) != "device" {
		return false
	}
	bootOut, err := exec.Command(adbExe, "-s", adbAddress, "shell", "getprop", "sys.boot_completed").Output()
	return err == nil && strings.TrimSpace(string(bootOut)) == "1"
}

func detectPackageName(adbExe, adbAddress string) string {
	if out, err := exec.Command(adbExe, "-s", adbAddress, "shell", "pm", "list", "packages").Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimPrefix(strings.TrimSpace(line), "package:")
			if strings.Contains(line, "papegames") || strings.Contains(line, "nn4") {
				return line
			}
		}
	}
	return defaultPackageName
}

func isAppRunning(adbExe, adbAddress, pkg string) bool {
	out, err := exec.Command(adbExe, "-s", adbAddress, "shell", "pidof", pkg).Output()
	return err == nil && strings.TrimSpace(string(out)) != ""
}

// launchOrFocusApp 统一启动或将应用激活至前台（mumu-cli -> am start -> monkey 三级降级）
func launchOrFocusApp(mumuPath, adbExe, adbAddress string, vmIndex int, pkg string, allowMonkey bool) {
	if cliExe := findMuMuCli(mumuPath); cliExe != "" {
		cmd := exec.Command(cliExe, "control", "-v", fmt.Sprintf("%d", vmIndex), "app", "launch", "--package", pkg)
		cmd.Dir = filepath.Dir(cliExe)
		if err := cmd.Run(); err == nil {
			return
		}
	}
	cmd := exec.Command(adbExe, "-s", adbAddress, "shell", "am", "start", "-n", pkg+"/com.nikki.nn4lib.NN4PlayerActivity")
	if err := cmd.Run(); err == nil || !allowMonkey {
		return
	}
	_ = exec.Command(adbExe, "-s", adbAddress, "shell", "monkey", "-p", pkg, "1").Run()
}

func launchGameApp(mumuPath, adbExe, adbAddress string, vmIndex int) error {
	pkg := detectPackageName(adbExe, adbAddress)
	fmt.Printf("[Launcher] 目标游戏包名: %s\n", pkg)

	if isAppRunning(adbExe, adbAddress, pkg) {
		fmt.Println("[Launcher] 检测到游戏进程已在运行，唤起至前台...")
		launchOrFocusApp(mumuPath, adbExe, adbAddress, vmIndex, pkg, false)
		time.Sleep(2 * time.Second)
		return nil
	}

	fmt.Println("[Launcher] 正在拉起《闪耀暖暖》游戏应用...")
	launchOrFocusApp(mumuPath, adbExe, adbAddress, vmIndex, pkg, true)

	fmt.Println("[Launcher] 等待游戏进程就绪...")
	deadline := time.Now().Add(20 * time.Second)
	retried := false
	startTime := time.Now()
	for time.Now().Before(deadline) {
		time.Sleep(1 * time.Second)
		if isAppRunning(adbExe, adbAddress, pkg) {
			fmt.Println("[Launcher] 游戏进程已确立，预留缓冲移交 Controller...")
			time.Sleep(3 * time.Second)
			return nil
		}
		if !retried && time.Since(startTime) >= 10*time.Second {
			fmt.Println("[Launcher] 启动用时较长，重新尝试发送拉起指令...")
			launchOrFocusApp(mumuPath, adbExe, adbAddress, vmIndex, pkg, true)
			retried = true
		}
	}
	return fmt.Errorf("等待游戏应用启动超时 (20s, 包名: %s)", pkg)
}

// resolveConfigPath 统一管理配置文件优先级探测与目标目录创建
func resolveConfigPath(projectRoot string, ensureDir bool) string {
	candidates := []string{
		filepath.Join(projectRoot, "config", "maa_pi_config.json"),
		filepath.Join(projectRoot, "assets", "config", "maa_pi_config.json"),
	}
	for _, p := range candidates {
		if fileExists(p) {
			return p
		}
	}
	target := candidates[0]
	if ensureDir {
		_ = os.MkdirAll(filepath.Dir(target), 0o755)
	}
	return target
}

// loadADBConfig 从现有配置文件加载已配置的 ADB 地址和路径
func loadADBConfig(projectRoot string) (string, string, int) {
	cfgPath := resolveConfigPath(projectRoot, false)
	if !fileExists(cfgPath) {
		return "", "", 0
	}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return "", "", 0
	}
	var root struct {
		ADB struct {
			Address string `json:"address"`
			ADBPath string `json:"adb_path"`
			Config  struct {
				Extras struct {
					MuMu struct {
						Index int `json:"index"`
					} `json:"mumu"`
				} `json:"extras"`
			} `json:"config"`
		} `json:"adb"`
	}
	if err := json.Unmarshal(stripJSONComments(data), &root); err == nil {
		return strings.TrimSpace(root.ADB.Address), strings.TrimSpace(root.ADB.ADBPath), root.ADB.Config.Extras.MuMu.Index
	}
	return "", "", 0
}

// savePathToConfig 写入或更新配置至 maa_pi_config.json
func savePathToConfig(projectRoot, mumuPath string) error {
	targetPath := resolveConfigPath(projectRoot, true)
	root := make(map[string]any)
	if fileExists(targetPath) {
		if data, err := os.ReadFile(targetPath); err == nil {
			_ = json.Unmarshal(stripJSONComments(data), &root)
		}
	}

	optMap, ok := root["option"].(map[string]any)
	if !ok {
		optMap = make(map[string]any)
		root["option"] = optMap
	}
	cfgSub, ok := optMap["MuMuConfig"].(map[string]any)
	if !ok {
		cfgSub = make(map[string]any)
		optMap["MuMuConfig"] = cfgSub
	}
	cfgSub["mumu_path"] = filepath.ToSlash(mumuPath)

	newBytes, err := json.MarshalIndent(root, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(targetPath, newBytes, 0o644)
}

// stripJSONComments 轻量剥离 JSONC 中的 // 注释，仅做基础双引号包裹判断以防误伤 URL
func stripJSONComments(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		clean := strings.TrimRight(line, "\r")
		for idx := strings.Index(clean, "//"); idx != -1; {
			if strings.Count(clean[:idx], `"`)%2 == 0 {
				clean = clean[:idx]
				break
			}
			next := strings.Index(clean[idx+2:], "//")
			if next == -1 {
				break
			}
			idx += 2 + next
		}
		lines[i] = clean
	}
	return []byte(strings.Join(lines, "\n"))
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
