package emulator

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

	pkgCN   = "com.papegames.nn4.cn"
	pkgTW   = "com.shining.nikki4.tw"
	pkgTWGP = "com.papegames.nn4.tw"
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

	// 1. 读取既有配置中定义的 ADB 端口、路径、模拟器路径、资源服务器与实例序号
	cfgADBAddress, cfgADBPath, cfgMuMuPath, cfgResource, vmIndex := loadConfig(paths.Root)
	adbAddress := defaultADBAddress
	if cfgADBAddress != "" {
		adbAddress = cfgADBAddress
	}

	// 2. 解析 MXU 传入的 option 参数与目标区服
	var inputPath string
	var rawJSON string
	remaining := flags.Args()
	if len(remaining) > 0 {
		rawJSON = remaining[len(remaining)-1]
		inputPath = extractMuMuPathFromJSON(rawJSON)
	}
	if inputPath == "" {
		inputPath = cfgMuMuPath
	}

	serverOpt := detectServer(rawJSON, cfgResource)
	fmt.Printf("[Emulator] 确认运行区服: %s\n", serverOpt)

	// 3. 定位 MuMu 路径（显式输入 -> 正在运行 -> 注册表关联 -> 默认目录 -> 置顶弹窗）
	mumuPath, wasAutoDetected, err := resolveMuMuPath(inputPath, true)
	if err != nil {
		return fmt.Errorf("定位 MuMu 模拟器失败: %w", err)
	}
	fmt.Printf("[Emulator] 确认 MuMu 路径: %s\n", mumuPath)

	// 4. 自动检测或弹窗选择时写回配置供 UI 回显
	if wasAutoDetected || inputPath == "" {
		if err := savePathToConfig(paths.Root, mumuPath); err != nil {
			fmt.Printf("[Emulator] 警告: 写回配置文件失败: %v\n", err)
		} else {
			fmt.Println("[Emulator] 模拟器路径已写入 MXU 配置文件")
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
	if err := launchGameApp(mumuPath, adbExe, adbAddress, vmIndex, serverOpt); err != nil {
		return fmt.Errorf("拉起游戏应用失败: %w", err)
	}

	fmt.Println("[Emulator] 启动准备完成，无缝移交 Controller")
	return nil
}

// extractServerFromJSON 从 MXU 传入的 option JSON 或配置文件中解析 ServerOption (CN / TW / TW_GP)
func extractServerFromJSON(rawJSON string) string {
	if !strings.HasPrefix(strings.TrimSpace(rawJSON), "{") {
		return ""
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(rawJSON), &data); err != nil {
		return ""
	}
	parseVal := func(target any) string {
		if val, ok := target.(string); ok {
			return strings.TrimSpace(val)
		}
		if sub, ok := target.(map[string]any); ok {
			for _, k := range []string{"caseName", "case", "value", "name"} {
				if v, ok := sub[k].(string); ok && strings.TrimSpace(v) != "" {
					return strings.TrimSpace(v)
				}
			}
		}
		return ""
	}
	if v := parseVal(data["ServerOption"]); v != "" {
		return v
	}
	if parseVal(data["ServerOption_TW_GP"]) == "Yes" {
		return "TW_GP"
	}
	if parseVal(data["ServerOption_TW"]) == "Yes" {
		return "TW"
	}
	if opt, ok := data["option"].(map[string]any); ok {
		if v := parseVal(opt["ServerOption"]); v != "" {
			return v
		}
		if parseVal(opt["ServerOption_TW_GP"]) == "Yes" {
			return "TW_GP"
		}
		if parseVal(opt["ServerOption_TW"]) == "Yes" {
			return "TW"
		}
	}
	return ""
}

// detectServer 解析目标区服（CLI 选项优先，其次为 PI_RESOURCE 环境变量，再次为配置中的 resource）
func detectServer(rawJSON, cfgResource string) string {
	if s := extractServerFromJSON(rawJSON); s != "" {
		return s
	}
	if env := os.Getenv("PI_RESOURCE"); env != "" {
		if strings.Contains(env, "台服") || strings.EqualFold(env, "TW") {
			return "TW"
		}
		if strings.Contains(env, "官服") || strings.Contains(env, "国服") || strings.EqualFold(env, "CN") {
			return "CN"
		}
	}
	if strings.Contains(cfgResource, "台服") || strings.EqualFold(cfgResource, "TW") {
		return "TW"
	}
	return "CN"
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

// resolveMuMuPath 通用解析 MuMu 可执行文件绝对路径。allowDialog 控制在所有自动探测未命中时是否呼出 WinForms 置顶弹窗。
func resolveMuMuPath(inputPath string, allowDialog bool) (string, bool, error) {
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

	if !allowDialog {
		return "", false, nil
	}

	fmt.Println("[Emulator] 未自动检测到默认安装路径，正在呼出置顶文件选择框...")
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
		fmt.Println("[Emulator] 检测到 MuMu 模拟器已处于运行就绪状态")
		return nil
	}

	fmt.Println("[Emulator] 正在启动 MuMu 模拟器...")
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

	fmt.Printf("[Emulator] 等待模拟器系统与 ADB 就绪 (%s)...\n", adbAddress)
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(1500 * time.Millisecond)
		_ = exec.Command(adbExe, "connect", adbAddress).Run()
		if isEmulatorReady(adbExe, adbAddress) {
			fmt.Println("[Emulator] MuMu 模拟器系统已完全就绪！")
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

func detectPackageName(adbExe, adbAddress, server string) string {
	installed := make(map[string]bool)
	var all []string
	if out, err := exec.Command(adbExe, "-s", adbAddress, "shell", "pm", "list", "packages").Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimPrefix(strings.TrimSpace(line), "package:")
			if line == pkgTW || line == pkgTWGP || line == pkgCN || strings.Contains(line, "nn4") {
				installed[line] = true
				all = append(all, line)
			}
		}
	}

	if server == "TW_GP" {
		if installed[pkgTWGP] {
			return pkgTWGP
		}
		if installed[pkgTW] {
			return pkgTW
		}
	} else if server == "TW" {
		if installed[pkgTW] {
			return pkgTW
		}
		if installed[pkgTWGP] {
			return pkgTWGP
		}
	} else {
		if installed[pkgCN] {
			return pkgCN
		}
	}

	if len(all) > 0 {
		return all[0]
	}
	if server == "TW_GP" {
		return pkgTWGP
	}
	if server == "TW" {
		return pkgTW
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

func launchGameApp(mumuPath, adbExe, adbAddress string, vmIndex int, server string) error {
	pkg := detectPackageName(adbExe, adbAddress, server)
	fmt.Printf("[Emulator] 目标游戏包名: %s\n", pkg)

	if isAppRunning(adbExe, adbAddress, pkg) {
		fmt.Println("[Emulator] 检测到游戏进程已在运行，唤起至前台...")
		launchOrFocusApp(mumuPath, adbExe, adbAddress, vmIndex, pkg, false)
		time.Sleep(2 * time.Second)
		return nil
	}

	fmt.Println("[Emulator] 正在拉起《闪耀暖暖》游戏应用...")
	launchOrFocusApp(mumuPath, adbExe, adbAddress, vmIndex, pkg, true)

	fmt.Println("[Emulator] 等待游戏进程就绪...")
	deadline := time.Now().Add(20 * time.Second)
	retried := false
	startTime := time.Now()
	for time.Now().Before(deadline) {
		time.Sleep(1 * time.Second)
		if isAppRunning(adbExe, adbAddress, pkg) {
			fmt.Println("[Emulator] 游戏进程已确立，预留缓冲移交 Controller...")
			time.Sleep(3 * time.Second)
			return nil
		}
		if !retried && time.Since(startTime) >= 10*time.Second {
			fmt.Println("[Emulator] 启动用时较长，重新尝试发送拉起指令...")
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

// resolveMXUResource 从 MXU 运行时配置（mxu-MaNikki4.json）中解析用户当前激活的资源包名称
func resolveMXUResource(projectRoot string) string {
	candidates := []string{
		filepath.Join(projectRoot, "config", "mxu-MaNikki4.json"),
		filepath.Join(projectRoot, "assets", "config", "mxu-MaNikki4.json"),
	}
	for _, p := range candidates {
		if data, err := os.ReadFile(p); err == nil {
			var mxuCfg struct {
				Instances []struct {
					ID           string `json:"id"`
					ResourceName string `json:"resourceName"`
				} `json:"instances"`
				LastActiveInstanceID string `json:"lastActiveInstanceId"`
			}
			if err := json.Unmarshal(data, &mxuCfg); err == nil {
				for _, inst := range mxuCfg.Instances {
					if inst.ID == mxuCfg.LastActiveInstanceID && strings.TrimSpace(inst.ResourceName) != "" {
						return strings.TrimSpace(inst.ResourceName)
					}
				}
				if len(mxuCfg.Instances) > 0 && strings.TrimSpace(mxuCfg.Instances[0].ResourceName) != "" {
					return strings.TrimSpace(mxuCfg.Instances[0].ResourceName)
				}
			}
		}
	}
	return ""
}

// loadConfig 从现有配置文件加载已配置的 ADB 地址、路径、模拟器路径与实例序号
func loadConfig(projectRoot string) (string, string, string, string, int) {
	cfgPath := resolveConfigPath(projectRoot, false)
	if !fileExists(cfgPath) {
		mxuRes := resolveMXUResource(projectRoot)
		return "", "", "", mxuRes, 0
	}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		mxuRes := resolveMXUResource(projectRoot)
		return "", "", "", mxuRes, 0
	}
	var root struct {
		Resource string `json:"resource"`
		ADB      struct {
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
		Option struct {
			MuMuPath   string `json:"mumu_path"`
			MuMuConfig struct {
				MuMuPath string `json:"mumu_path"`
			} `json:"MuMuConfig"`
		} `json:"option"`
	}
	if err := json.Unmarshal(stripJSONComments(data), &root); err == nil {
		mumuPath := root.Option.MuMuPath
		if mumuPath == "" {
			mumuPath = root.Option.MuMuConfig.MuMuPath
		}
		res := strings.TrimSpace(root.Resource)
		if mxuRes := resolveMXUResource(projectRoot); mxuRes != "" {
			res = mxuRes
		}
		return strings.TrimSpace(root.ADB.Address), strings.TrimSpace(root.ADB.ADBPath), strings.TrimSpace(mumuPath), res, root.ADB.Config.Extras.MuMu.Index
	}
	mxuRes := resolveMXUResource(projectRoot)
	return "", "", "", mxuRes, 0
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

// ShutdownEmulator 安全优雅地关闭 MuMu 模拟器实例，防止残留唤醒且绝不误伤其他多开实例。
// closeLauncherOpt 控制是否在模拟器实例安全退出后直接结束 MuMu 启动器主面板（默认开启 true）。
func ShutdownEmulator(projectRoot string, closeLauncherOpt ...bool) error {
	closeLauncher := true
	if len(closeLauncherOpt) > 0 {
		closeLauncher = closeLauncherOpt[0]
	}

	cfgADBAddress, cfgADBPath, cfgMuMuPath, _, vmIndex := loadConfig(projectRoot)
	adbAddress := defaultADBAddress
	if cfgADBAddress != "" {
		adbAddress = cfgADBAddress
	}

	mumuPath, _, err := resolveMuMuPath(cfgMuMuPath, false)
	if err != nil {
		return fmt.Errorf("定位 MuMu 模拟器失败: %w", err)
	}
	if mumuPath == "" {
		fmt.Println("[Emulator] 未检测到正在运行或已安装的模拟器，静默跳过关机")
		return nil
	}

	cliExe := findMuMuCli(mumuPath)
	adbExe := resolveADBPath("", mumuPath, cfgADBPath)

	// 1. 获取目标实例初始状态与专属 PID (供定向关机与超时兜底，防止误伤多开)
	var initialInfo mumuVMInfo
	if cliExe != "" {
		info, err := queryVMInfo(cliExe, vmIndex)
		if err == nil {
			initialInfo = info
			if !initialInfo.IsProcessStarted {
				fmt.Printf("[Emulator] 实例 %d 当前未在运行，无需关机\n", vmIndex)
				if adbExe != "" {
					_ = exec.Command(adbExe, "disconnect", adbAddress).Run()
				}
				return nil
			}
		}
	}

	// 2. 发送安全关机指令，确保安卓虚拟机数据正常落盘
	if cliExe != "" {
		fmt.Printf("[Emulator] 正在使用 mumu-cli 发送关机指令 (实例: %d)...\n", vmIndex)
		cmd := exec.Command(cliExe, "control", "-v", fmt.Sprintf("%d", vmIndex), "shutdown")
		cmd.Dir = filepath.Dir(cliExe)
		_ = cmd.Run()
	} else if adbExe != "" {
		// 降级方案：若未找到 mumu-cli，尝试通过 ADB 发送软关机指令通知安卓系统落盘
		fmt.Println("[Emulator] 未检测到 mumu-cli，尝试通过 ADB 发送软关机指令...")
		_ = exec.Command(adbExe, "-s", adbAddress, "shell", "reboot", "-p").Run()
	}

	// 3. 切断 ADB 连接，杜绝守护机制因端口探测轮询而再次唤醒模拟器
	if adbExe != "" {
		_ = exec.Command(adbExe, "disconnect", adbAddress).Run()
	}

	// 4. 定向等待该实例退出（最长 15 秒，通过 CLI info 轮询，保障 Android 充分落盘且不派生 tasklist 外部进程）
	vmTerminated := false
	lastInfo := initialInfo
	if cliExe != "" {
		deadline := time.Now().Add(15 * time.Second)
		for time.Now().Before(deadline) {
			time.Sleep(500 * time.Millisecond)
			if info, err := queryVMInfo(cliExe, vmIndex); err == nil {
				lastInfo = info
				if !info.IsProcessStarted {
					vmTerminated = true
					break
				}
			}
		}
	} else {
		// 无 CLI 场景退化：留出 3 秒基础落盘缓冲后退出
		time.Sleep(3 * time.Second)
	}

	// 5. 若目标实例超时未退出，仅针对该实例专属 PID 定向强杀，绝不误伤多开
	if !vmTerminated && cliExe != "" {
		fmt.Printf("[Emulator] 实例 %d 关机超时，执行专属 PID 定向兜底清理...\n", vmIndex)
		if lastInfo.PID > 0 {
			killPID(lastInfo.PID)
		}
		if lastInfo.HeadlessPID > 0 {
			killPID(lastInfo.HeadlessPID)
		}
	}

	// 6. 检查是否存在其他正在运行的实例：仅在用户开启选项且无其他多开实例时直接结束启动器主面板（绝不触碰 MuMuNxService 等系统底层服务）
	if cliExe != "" && closeLauncher {
		if !hasOtherRunningInstances(cliExe, vmIndex) {
			fmt.Println("[Emulator] 已开启关闭启动器，正在退出 MuMu 启动器主面板...")
			launcherExe := filepath.Base(mumuPath)
			if launcherExe == "" {
				launcherExe = "MuMuNxMain.exe"
			}
			_ = exec.Command("taskkill", "/F", "/IM", launcherExe).Run()
		} else {
			fmt.Println("[Emulator] 检测到其他模拟器实例仍在运行，保留 MuMu 启动器环境")
		}
	}

	fmt.Println("[Emulator] MuMu 模拟器已安全关闭")
	return nil
}

type mumuVMInfo struct {
	Index            any  `json:"index"`
	IsProcessStarted bool `json:"is_process_started"`
	PID              int  `json:"pid"`
	HeadlessPID      int  `json:"headless_pid"`
}

// queryVMInfo 通过 mumu-cli 查询特定实例的运行状态与专属 PID
func queryVMInfo(cliExe string, vmIndex int) (mumuVMInfo, error) {
	cmd := exec.Command(cliExe, "info", "-v", fmt.Sprintf("%d", vmIndex))
	cmd.Dir = filepath.Dir(cliExe)
	out, err := cmd.Output()
	if err != nil {
		return mumuVMInfo{}, err
	}
	var info mumuVMInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return mumuVMInfo{}, err
	}
	return info, nil
}

// hasOtherRunningInstances 检查系统中除当前实例外是否还有其他 MuMu 实例正在运行（保护多开）
func hasOtherRunningInstances(cliExe string, currentVMIndex int) bool {
	cmd := exec.Command(cliExe, "info", "-v", "all")
	cmd.Dir = filepath.Dir(cliExe)
	out, err := cmd.Output()
	if err != nil {
		// 命令异常时采取 Fail-Safe 保守策略，默认判定有其他实例，防止误杀主面板
		return true
	}
	return parseOtherRunningInstances(out, currentVMIndex)
}

func parseOtherRunningInstances(out []byte, currentVMIndex int) bool {
	targetIndexStr := fmt.Sprintf("%d", currentVMIndex)

	// 尝试反序列化为数组（多实例场景）
	var list []mumuVMInfo
	if err := json.Unmarshal(out, &list); err == nil {
		for _, vm := range list {
			if fmt.Sprintf("%v", vm.Index) != targetIndexStr && vm.IsProcessStarted {
				return true
			}
		}
		return false
	}

	// 尝试反序列化为单个对象（单实例场景）
	var single mumuVMInfo
	if err := json.Unmarshal(out, &single); err == nil {
		if fmt.Sprintf("%v", single.Index) != targetIndexStr && single.IsProcessStarted {
			return true
		}
		return false
	}

	// 解析完全异常时，同样采取 Fail-Safe 保守策略
	return true
}

// killPID 定向强制终止指定 PID 进程
func killPID(pid int) {
	if pid <= 0 {
		return
	}
	_ = exec.Command("taskkill", "/F", "/PID", fmt.Sprintf("%d", pid)).Run()
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
