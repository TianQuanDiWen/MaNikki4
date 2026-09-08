# MaNikki4

基于 [MaaFramework](https://github.com/MaaXYZ/MaaFramework) 开发、使用 MXU 前端运行的暖暖系列游戏自动化项目。

项目通过 ADB 连接 MuMu 模拟器，使用图像识别和 OCR 完成截图、点击及滑动等操作。目前主要包含《闪耀暖暖》的部分日常任务。

## 已支持任务

### 闪耀暖暖

- 启动游戏
- 好友送心、领取体力及日常采购
- 组队副本：不落的帷幕、印象航旅、时光钟表铺
- 钻石竞技场
- 心之门免费抽取及分享奖励
- 联盟捐献、机密任务及联盟福利

任务仍在持续完善，实际可用情况可能受游戏版本、区服、活动界面和网络状态影响。

## 运行环境

- Windows 10 或 Windows 11（x64）
- MuMu 模拟器
- 模拟器分辨率设置为竖屏 `1080 × 1920`
- 已在 MuMu 模拟器中开启 ADB 调试
- 游戏区服及画面与当前识别资源匹配
- [Microsoft Edge WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/)
- [Microsoft Visual C++ 2015–2022 Redistributable（x64）](https://aka.ms/vs/17/release/vc_redist.x64.exe)

## 使用方法

1. 启动 MuMu 模拟器，并确认目标游戏可以正常运行。
2. 打开 MXU 前端。
3. 选择“MuMu模拟器”控制器。
4. 选择 MXU 检测到的 ADB 设备和“官服”资源。
5. 勾选需要执行的任务并开始运行。

MaaFramework 会自动选择可用的 ADB 截图和触控方式。若 MXU 未检测到模拟器，请先确认 MuMu 的 ADB 调试已开启，并检查模拟器是否已被 ADB 正常识别。

“启动闪暖”任务通过识别模拟器桌面上的游戏图标启动《闪耀暖暖》，运行该任务前请先返回模拟器桌面。其他闪暖任务通常需要游戏已进入主菜单。

## 目录结构

```text
assets/interface.json              # 项目、控制器、资源及任务入口配置
assets/resource/pipeline/          # 自动化任务流水线
assets/resource/image/             # 图像识别模板
assets/MaaCommonAssets/             # 公共 OCR 资源（Git 子模块）
agent/                              # 自定义 Agent 扩展
deps/                               # 本地 MaaFramework 开发依赖
tools/                              # 安装、配置及 Schema 校验工具
maatools.config.mts                 # MaaTools 配置
build.ps1                           # 本地构筑入口
build.config.json                   # 本地与发布构筑配置
```

## 开发与校验

安装 Node.js 依赖后，可以运行 MaaTools 检查接口配置及任务流水线：

```powershell
npm install
npx.cmd --no-install @nekosu/maa-tools check
```

新增或删除任务流水线时，需要同步修改 `assets/interface.json` 中的 `task` 列表。图像模板应放在 `assets/resource/image/` 下，并在流水线中使用相对路径引用。

克隆项目时需要同时初始化公共资源子模块：

```powershell
git clone --recurse-submodules <仓库地址>
```

## 免责声明

本项目为非官方自动化工具，与游戏开发商及运营商无关。使用自动化脚本可能违反游戏的用户协议或运营规则，并可能导致账号受到限制、暂停或永久封禁。

使用者应在使用前自行了解并遵守相关规则，充分评估风险。使用本项目产生的一切后果由使用者自行承担，项目作者及贡献者不对账号损失或其他直接、间接损失承担责任。

## 鸣谢

- [MaaFramework](https://github.com/MaaXYZ/MaaFramework)
- [MXU](https://github.com/MistEO/MXU)
