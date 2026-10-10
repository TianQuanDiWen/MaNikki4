<!-- markdownlint-disable MD033 MD041 -->
<p align="center">
  <img alt="LOGO" src="assets/logo.png" width="256" height="256" />
</p>
<div align="center">

# MaNikki4

<!-- prettier-ignore-start -->
<!-- markdownlint-disable-next-line MD036 -->
_✨ 基于 MaaFramework 的《闪耀暖暖》全自动日常减负辅助工具 ✨_
<!-- prettier-ignore-end -->

</div>

<p align="center">
  <img alt="Go" src="https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go">
  <img alt="platform" src="https://img.shields.io/badge/platform-Windows%2010%2B%20x64-informational">
  <a href="https://github.com/MaaXYZ/MaaFramework" target="_blank"><img alt="framework" src="https://img.shields.io/badge/Powered%20by-MaaFramework-blueviolet"></a>
  <br>
  <a href="LICENSE"><img alt="license" src="https://img.shields.io/github/license/TianQuanDiWen/MaNikki4?color=blue"></a>
  <img alt="activity" src="https://img.shields.io/github/commit-activity/m/TianQuanDiWen/MaNikki4?color=%23ff69b4">
  <a href="https://github.com/TianQuanDiWen/MaNikki4/stargazers"><img alt="stars" src="https://img.shields.io/github/stars/TianQuanDiWen/MaNikki4?style=social"></a>
</p>

## 简介

**MaNikki4** 是基于 **[MaaFramework](https://github.com/MaaXYZ/MaaFramework)** 深度定制的《闪耀暖暖》（Shining Nikki）全自动日常减负辅助工具。

采用纯 Go 原生常驻 Agent 驱动与自维护现代化桌面端 UI（[MXU](https://github.com/TianQuanDiWen/MXU_tqdw)），专注一键清理繁琐日常打卡与深度制衣管理，为搭配师打造极速、稳定、低占用的自动化减负体验。

> [!NOTE]
> 🚧 **开发状态提示**：本项目当前处于积极开发与迭代维护中。

---

## 🎯 定位与免责声明

* **专注日常减负**：本工具纯粹用于自动化执行游戏内重复性的日常打卡（送心、组队本、竞技场、制衣扫荡、抽卡、联盟等），不包含任何破坏游戏平衡、逆向解密或篡改游戏内存的功能。
* **完全开源免费**：本项目为非营利开源软件，遵循 AGPL-3.0 开源协议，完全免费，严禁任何形式的倒卖与商用。
* **免责声明**：本项目为第三方非官方自动化工具，与《闪耀暖暖》官方运营商及开发商（叠纸游戏 Papergames）无关。使用过程中请遵守游戏用户协议，使用风险由使用者自行评估承担。

---

## 📋 功能清单

| 任务模块 | 功能说明 |
| :--- | :--- |
| **🎮 启动模拟器与游戏** | 自动检测并启动 MuMu 模拟器，拉起闪耀暖暖（支持国服与台服客户端） |
| **🚀 登陆游戏** | 自动跳过开屏公告与确认弹窗，进入游戏大厅 |
| **💌 送礼** | 好友互赠与领取爱心、暖暖家园送礼物和零食 |
| **👥 组队本** | 自动扫荡不落的帷幕、印象航旅、时光钟表铺每日次数 |
| **⚔️ 竞技场** | 自动配置挑战套装，智能挑战低战力对手，战力过高自动金币换一批 |
| **🚪 抽卡** | 自动完成心之门（幻之海 / 谜之海）每日免费感应 |
| **🏛️ 联盟任务** | 自动完成每日联盟捐献、机密关卡扫荡、领取红包与补给 |
| **🌊 忆海心阶** | 自动进入并推进心阶挑战，自动搭配，完成周常/日常结算（可选：境遇挑战 🧪测试中） |
| **💅 美甲店铺** | 自动雇佣店员打理美甲店，一键提取营业收益 |
| **💎 分享** | 每日分享领粉钻，支持自选 APP（QQ / 微信 / 微博 / 小红书 / LINE / Instagram），分享后返回游戏 |
| **🏆 搭配评选赛** | 自动完成搭配评选赛，快速完成每日点赞投票，可选自动参赛投稿 |
| **👗 时空回廊** | 自动挑战时空回廊关卡（需配置到**制衣引导第 1 套**） |
| **🧵 主线制衣** | 自动扫荡普通主线关卡与购买服装店材料（需配置到**制衣引导第 2 套**） |
| **📋 任务奖励** | 全部日常完成后，一键领满每日活跃度宝箱，自动领取月卡奖励与签到（可选：充值特权与耀星钻卡 🧪测试中） |
| **🛑 退出模拟器** | 全部任务执行完毕后，安全优雅关闭 MuMu 模拟器 |

---

## 📖 使用说明

### 1. 基准运行环境
* **模拟器**：**MuMu 模拟器 12**（需在模拟器设置中开启 **ADB 本地调试**，程序会自动探测连接）。
* **分辨率**：**竖屏 9:16**，推荐基准分辨率 **`720 × 1280`**（底层支持等比自适应）。
* **游戏主题**：主界面 UI 须设置为 **“暖调回忆”**（自动化流程基于此主题开发，避免换肤引起图标识别不准）。
* **客户端与语言**：支持**国服**与**台服**客户端（官网直下版 / Google Play 版）。**台服用户请注意：游戏内语言须保持为简体中文**（当前 OCR 识别词库基于简中）。

### 2. 制衣引导配置
游戏内前往 **【设计中心】 ➔ 【制衣引导】** 配置目标套装，脚本即可全自动寻路扫荡：

| 游戏内槽位 | 对应任务 | 配置与行为说明 |
| :--- | :--- | :--- |
| **第 1 套（左槽）** | **👗 时空回廊** | 放置正在刷的【时空回廊】套装。可选开启“挑战主线关卡”（默认关闭）。 |
| **第 2 套（右槽）** | **🧵 主线制衣** | 放置正在刷的【主线/工坊】套装，自动挑战对应主线关卡。 |

### 3. 可选开关与配置
* **🌐 连接设置 ➔ 游戏区服**：  
  在 MXU 连接设置的资源（Resource）下拉菜单中直接选择“官服”或“台服”对应版本。
* **💎 分享 ➔ 分享目标 APP**：  
  国服推荐选择 QQ / 微信 / 小红书，台服推荐选择 LINE / Instagram（需模拟器中预装对应应用）。
* **🌊 忆海心阶 ➔ 境遇挑战（测试中，默认关闭）**：  
  在完成常规真我之境挑战后，额外自动进行心阶境遇关卡挑战。该功能目前处于测试中，可在任务右侧选项中按需开启。
* **📋 任务奖励 ➔ 充值特权与耀星钻卡（测试中，默认关闭）**：  
  在每日返钻流程中，额外检测并领取**耀星钻卡**每日返钻以及**充值特权**每周免费礼包。该功能目前处于测试中，需要该特权的搭配师可在任务右侧选项中手动开启。
* **🏆 搭配评选赛 ➔ 自动投稿参赛**：  
  若本周尚未参赛，自动随机穿搭集齐套装并完成投稿（默认开启；关闭时仅点赞投票）。
* **🛑 退出模拟器 ➔ 同时关闭启动器主面板**：  
  任务结束后彻底关闭 MuMu 启动器后台（默认开启）。

---

## ✨ 核心架构与技术亮点

本项目采用 MaaFramework 图像识别推理引擎与 Go 原生常驻 Agent 深度协同，兼具精准感知与极致性能。

### 1. 自动化双眼：OCR 智能识字与模板精准搜图
* **🔍 OCR 光学字符识别（智能读懂语义）**：  
  就像为自动化程序配备了“能识字的眼睛”。不仅能自动识别关卡名称、挑战剩余次数、对手战力与材料名称，还结合了灵活的正则表达式匹配。即使游戏文案轻微更新或列表上下滚动，也能自适应精准捕捉目标，告别僵硬的死板坐标。
* **🖼️ 模板匹配 Template Matching（像素级精准搜图）**：  
  针对游戏内没有文字的固定图标（如返回上一页左箭头、粉钻与金币图标、勾选确认框等），采用像素级图像模板比对技术。配合绿色掩码（Green Mask）屏蔽动态粒子与呼吸光效干扰，实现毫秒级快速抗噪定位。

### 2. 原创动态对齐识别器：`MatchRowButton`（看文字自动找按钮）
针对游戏内材料列表“多行并存、顺序动态变化、带有滚动条”的痛点，自研了 `MatchRowButton` 行级对齐识别组件：
* 在指定区域内同时扫描所有材料关键字与操作按钮；
* **基于水平几何容差算法（$|Y_{btn} - Y_{kw}| \le \text{tolerance}$）**，自动将目标文字（如“时空回廊”、“主线 6-10”、“服装店”）与同高度右侧的“前往 / 购买”按钮动态配对；
* 无论材料列表如何动态插入新材料或翻页滚动，都能精准找到对应的那一行点击，彻底取代易错位的绝对坐标死点方案。

### 3. 纯 Go 原生极速 Agent（超轻量、免环境折腾）
* 基于 MaaFramework 官方 Go 绑定开发，原生编译为二进制常驻后台。
* 运行内存占用仅 **10MB 左右**，启动仅需数毫秒，彻底告别数百兆庞大且易损坏的 Python 虚拟环境。
* 通过本地安全 Socket 与底层 C++ 视觉推理引擎双向通信，任务事件调度与状态反馈几乎零延迟。

---

## 🛠️ 本地开发与构建

### 运行环境要求
* **操作系统**：Windows 10 / 11 x64
* **模拟器**：MuMu 模拟器 12（竖屏分辨率推荐 `720 × 1280`，需在模拟器设置中开启 **ADB 调试**）
* **基础运行库**：[WebView2 Runtime](https://developer.microsoft.com/zh-cn/microsoft-edge/webview2/)、[VC++ 2015-2022 x64](https://aka.ms/vs/17/release/vc_redist.x64.exe)
* **编译工具链**：PowerShell 5.1+、Go 1.24+、[uv](https://github.com/astral-sh/uv)（或 Python 3.10+）

### 一键构建步骤

```powershell
# 1. 克隆代码仓库及子模块
git clone --recurse-submodules https://github.com/TianQuanDiWen/MaNikki4.git
cd MaNikki4

# 2. 执行自动化构建脚本（自动拉取依赖、编译 Go Agent 并打包输出至 install/ 目录）
.\build.ps1 -Action Build

# 3. 运行调试
# 进入 install 目录，启动 MaNikki4.exe 即可直接测试
cd install
.\MaNikki4.exe
```

---

## 📂 项目结构概览

```
MaNikki4/
├── agent/                       # 原生 Go Agent 核心代码
│   ├── cmd/manikki-agent/       # Agent 主入口程序
│   └── internal/
│       ├── agentserver/         # Agent 生命周期服务与管道注册
│       ├── arena/               # 竞技场自动挑对手与战力决策业务
│       ├── emulator/            # MuMu 模拟器路径探测与 CLI 管理
│       ├── matcher/             # MatchRowButton 行容差动态按钮对齐识别器
│       └── runtimepath/         # 运行时路径自适应嗅探
├── assets/                      # 资源与流水线定义
│   ├── interface.json           # MXU 任务入口与配置项 schema
│   └── resource/
│       ├── image/               # 模板匹配图标（返回箭头、粉钻等）
│       └── pipeline/            # MaaFramework 任务流水线
│           ├── 登陆游戏.json
│           ├── 公共流程.json
│           ├── 送礼.json
│           ├── 组队本.json
│           ├── 竞技场.json
│           ├── 抽卡.json
│           ├── 联盟任务.json
│           ├── 忆海心阶.json
│           ├── 美甲.json
│           ├── 分享.json        # 独立分享与外部 APP 自愈
│           ├── 评选赛.json      # 搭配评选赛自动当评委投票
│           ├── 制衣.json        # 时空回廊/主线双合一极简制衣
│           └── 任务.json        # 每日活跃度结算
├── build.ps1                    # 全自动跨组件构建脚本
└── README.md
```

---

## 🙏 鸣谢与生态

* 核心底层图像处理与自动化控制引擎由 [MaaFramework](https://github.com/MaaXYZ/MaaFramework) 强力驱动。
* 桌面端 UI 基于 [MistEO/MXU](https://github.com/MistEO/MXU) 及其自维护定制分支 [TianQuanDiWen/MXU_tqdw](https://github.com/TianQuanDiWen/MXU_tqdw)。
* OCR 基础模型与字典资源来自 [MaaCommonAssets](https://github.com/MaaXYZ/MaaCommonAssets)。

---

## 📄 开源许可证

本项目采用 [GNU Affero General Public License v3.0 (AGPL-3.0)](LICENSE) 许可证开源。

Copyright (c) 2024-2026 TianQuanDiWen <https://github.com/TianQuanDiWen>  
Copyright (c) 2024 MaaXYZ
