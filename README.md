# MaNikki4

[![Status](https://img.shields.io/badge/Status-In%20Development-orange?style=flat-square)](https://github.com/TianQuanDiWen/MaNikki4)
[![Platform](https://img.shields.io/badge/Platform-Windows%2010%2B%20x64-informational?style=flat-square)](https://github.com/TianQuanDiWen/MaNikki4)
[![Framework](https://img.shields.io/badge/Powered%20by-MaaFramework-blueviolet?style=flat-square)](https://github.com/MaaXYZ/MaaFramework)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)

> 🎮 **基于 [MaaFramework](https://github.com/MaaXYZ/MaaFramework) 的《闪耀暖暖》（Shining Nikki）全自动日常减负辅助工具。**  
> 采用高性能原生 Go Agent 与自维护现代化 MXU 前端，专注日常打卡清理，一键减负护肝。

> [!NOTE]
> 🚧 **开发中提示（Work In Progress）**：本项目当前处于积极开发与功能调试阶段，暂未发布正式预编译 Release 包。目前主要面向开发者与尝鲜用户通过源码本地构建运行。

---

## 🎯 定位与免责声明

* **专注日常减负**：纯粹为自动化清理每日重复打卡任务而生（送礼、组队本、竞技场、抽卡、联盟等），不提供任何形式的破坏游戏平衡功能。
* **完全免费开源**：非盈利开源工具，完全免费，严禁倒卖。
* **免责声明**：本项目为非官方开源自动化工具，与游戏开发商（叠纸游戏 Papergames）及运营商无关。仅供技术交流与日常减负使用，使用风险由使用者自行评估并承担。

---

## 📋 功能清单

| 任务名称 | 功能覆盖 | 自动化特性 / 异常处理 | 当前状态 |
| :--- | :--- | :--- | :---: |
| **🎮 启动模拟器与游戏** | 模拟器及游戏拉起（预任务） | 智能探测路径与端口，冷启动就绪检测与配置自动回写 | `已就绪` |
| **🚀 登陆游戏** | 自动登录进入游戏主界面 | 若游戏已在运行或主菜单则自动接管 | `已就绪` |
| **💌 闪暖送礼** | 好友送心、领体力与采购 | 自动购买流光阁/常驻免费物资 | `已就绪` |
| **👥 闪暖组队本** | 不落的帷幕/印象航旅/时光钟表铺 | 自动扫荡每日挑战次数并结算流转 | `已就绪` |
| **⚔️ 闪暖竞技场** | 钻石竞技场全自动对决 | **原生 Go Agent 动态 OCR 比对**：单次缓存自身战力，智能挑低战力对手挑战，打不过自动金币换一批 | `已就绪` |
| **🚪 闪暖抽卡** | 心之门免费抽取与分享 | 自动抽取幻之海/谜之海，领取分享奖励 | `已就绪` |
| **🏛️ 联盟任务** | 联盟日常贡献与福利 | 金币/暗语捐献，机密任务扫荡，领红包/补给 | `已就绪` |
| **🌊 忆海心阶** | 每日心阶挑战 | 自动完成关卡推进 | `已就绪` |
| **💅 美甲店铺** | 美甲日常经营 | 自动雇佣店员并提取店铺收益 | `已就绪` |
| **🛑 退出模拟器** | 安全优雅关闭模拟器 | 首选 `mumu-cli` 安全关机，未运行静默跳过 | `已就绪` |


---

## ✨ 架构亮点

* **纯 Go 原生扩展 Agent**：彻底脱离庞大的 Python 解释器，内存占用仅约 10MB，冷启动极速；通过 MaaFramework 官方 Go 绑定与底层 C++ 视觉引擎通过本地 Socket 高效通信。
* **高内聚解耦架构**：`runtimepath`（环境自适应感知）与 `agentserver`（通用服务生命周期）作为独立基础设施，游戏业务逻辑物理隔离在 `arena` 中，架构具备良好的跨项目复用性。
* **智能战力比对决策**：进入竞技场单次 OCR 暂存自身战力，后续循环复用；识别区域（ROI）完全委托给 Pipeline 节点配置，零坐标硬编码，内置字距消空防截断容错。
* **自维护 MXU 前端**：深度适配自维护的 [MXU_tqdw](https://github.com/TianQuanDiWen/MXU_tqdw) 前端，提供轻量、现代化的桌面交互体验。

---

## 🚀 本地开发与构建

### 运行环境
* **操作系统**：Windows 10 / 11 x64
* **模拟器**：MuMu 模拟器 12（竖屏 `720 × 1280` 或 `1080 × 1920`，需在模拟器设置中开启 **ADB 调试**）
* **基础依赖**：[WebView2 Runtime](https://developer.microsoft.com/zh-cn/microsoft-edge/webview2/)、[VC++ 2015-2022 x64](https://aka.ms/vs/17/release/vc_redist.x64.exe)
* **工具链**：PowerShell 5.1+、Go 1.24+、[uv](https://github.com/astral-sh/uv)（或 Python 3.10+）

### 构建步骤
```powershell
# 1. 克隆仓库及子模块
git clone --recurse-submodules https://github.com/TianQuanDiWen/MaNikki4.git
cd MaNikki4

# 2. 一键本地构建（自动拉取依赖、编译 Go Agent 并输出至 install/ 目录）
.\build.ps1 -Action Build

# 3. 运行调试
# 构建完成后，直接运行 install/MaNikki4.exe 启动前端测试
```

---

## 🙏 鸣谢

* 核心底层引擎由 [MaaFramework](https://github.com/MaaXYZ/MaaFramework) 强力驱动。
* 前端 UI 基于 [MistEO/MXU](https://github.com/MistEO/MXU) 及自维护定制分支 [TianQuanDiWen/MXU_tqdw](https://github.com/TianQuanDiWen/MXU_tqdw)。
* OCR 通用模型资源基于 [MaaCommonAssets](https://github.com/MaaXYZ/MaaCommonAssets)。
