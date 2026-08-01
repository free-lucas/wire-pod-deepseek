# wire-pod 在树莓派服务器上的部署测试指南

这份文档的目标是帮你把当前仓库部署到你自己的树莓派服务器上做测试，尽量用“照着执行”的方式来写。

适用服务器：

- SSH 地址：`pi@192.168.1.38`
- 服务器信息来源：`/Users/yaodong/Documents/trae_projects/wire-pod/pi服务器.md`

为了避免把密码扩散到更多文件里，这里不重复记录密码。需要时直接查看你本地的 `pi服务器.md`。

## 1. 这次部署要测什么

如果你当前主要是在验证“小智，小智”这次改造，建议把测试目标分成 3 层：

1. **服务能不能在树莓派上正常跑起来**
2. **网页和 SDK 页面文案是否已经切到中文语义**
3. **如果手边有机器人，再继续测真实唤醒和问答链路**

对应参考文档：

- [specs/001-xiaozhi-wakeword/quickstart.md](file:///Users/yaodong/Documents/trae_projects/wire-pod/specs/001-xiaozhi-wakeword/quickstart.md)

## 2. 推荐的部署方式

对于你现在这个阶段，我建议优先用下面这套方式：

1. 本机改代码
2. 把仓库同步到树莓派
3. 在树莓派上直接运行 `sudo ./chipper/start.sh`
4. 用浏览器访问页面验证
5. 如果要长期挂着，再切换到 `systemd`

原因很简单：

- 先手动跑，最容易看报错
- 当前你还在频繁改代码，没必要一上来就做成复杂部署

## 3. 第一次连接服务器

在你本机终端执行：

```bash
ssh pi@192.168.1.38
```

首次连接如果提示确认指纹，输入：

```bash
yes
```

连上以后，先做一轮基础检查：

```bash
uname -a
pwd
whoami
df -h
free -h || vm_stat
```

如果你想确认 Go、Git 是否可用：

```bash
go version
git --version
```

## 4. 把代码传到树莓派

有两种常见做法。

### 方式 A：树莓派上直接 `git clone` 或 `git pull`

适合你已经在树莓派上放过这个仓库。

首次：

```bash
git clone <你的仓库地址> ~/wire-pod
```

后续更新：

```bash
cd ~/wire-pod
git pull
```

### 方式 B：从本机把当前工作目录同步过去

适合你本机有还没推远端的改动，想直接拿本地版本测试。

在你本机执行：

```bash
rsync -avz --delete \
  --exclude ".git" \
  --exclude "node_modules" \
  --exclude ".DS_Store" \
  "/Users/yaodong/Documents/trae_projects/wire-pod/" \
  pi@192.168.1.38:~/wire-pod/
```

如果你机器上没有 `rsync`，也可以先压缩再传，或者直接用 `scp`。

## 5. 登录树莓派后先检查仓库状态

```bash
cd ~/wire-pod
git status --short
ls
ls chipper
```

重点看这几个文件是不是都在：

- `setup.sh`
- `chipper/start.sh`
- `chipper/source.sh`
- `chipper/apiConfig.json`（可能首次还没生成完整内容）

## 6. 最关键的前置条件

`chipper/start.sh` 的实际要求，仓库里已经写死了两件事：

1. **必须用 root 跑**
2. **必须存在 `chipper/source.sh`**

也就是：

```bash
cd ~/wire-pod
sudo ./chipper/start.sh
```

如果报错：

```text
You need to make a source.sh file.
```

说明这台树莓派还没完成初始安装，先执行：

```bash
cd ~/wire-pod
sudo ./setup.sh
```

然后按脚本提示把 `source.sh` 和依赖准备好。

## 7. 启动前建议先确认的配置

先看 `source.sh`：

```bash
cd ~/wire-pod/chipper
sed -n '1,200p' source.sh
```

重点关注：

- `STT_SERVICE`
- BLE 相关配置
- 端口相关配置

这一步很重要，因为 `start.sh` 会根据 `STT_SERVICE` 走不同入口：

- `vosk`
- `whisper.cpp`
- `coqui`
- 其他实验选项

## 8. 不同 STT 的注意点

### `vosk`

如果你用的是 `vosk`，通常要确认这些目录或库是可用的：

```bash
ls ~/.vosk
ls ~/.vosk/libvosk
```

### `whisper.cpp`

如果你用的是 `whisper.cpp`，要确认仓库同级或预期位置的 `whisper.cpp` 相关文件存在，并且库路径没问题。

### 这次“小智，小智”改造的建议

如果你只是先验证页面和配置链路，优先保证：

1. `wire-pod` 能启动
2. setup 页面可访问
3. STT 页面里能看到 `zh-CN`

## 9. 手动启动 wire-pod

在树莓派上执行：

```bash
cd ~/wire-pod
sudo ./chipper/start.sh
```

如果正常，你应该会看到服务启动日志持续输出。

如果你只想先看启动是否通过，也可以另开一个终端窗口再看端口：

```bash
ss -lntp | grep -E "80|8080|8084"
```

## 10. 浏览器里怎么访问

假设树莓派 IP 还是 `192.168.1.38`，常用入口一般是：

### 配置页

```text
http://192.168.1.38:8080
```

### setup 页面

```text
http://192.168.1.38:8080/setup.html
```

### SDK 页面

```text
http://192.168.1.38/sdkapp
```

如果 SDK 页面端口访问不通，再检查服务是否真的监听了 `:80`。

## 11. 这次需求建议怎么测

### 第 1 步：先测 setup 页面

打开：

```text
http://192.168.1.38:8080/setup.html
```

检查：

1. STT 语言里是否有 `Chinese (CN)` / `zh-CN`
2. 切换后是否提示成功，或提示下载模型

### 第 2 步：再测 SDK 页面

打开：

```text
http://192.168.1.38/sdkapp
```

检查：

1. 默认文案是否已经从 `Hey Vector` 变成 `小智，小智`
2. locale 区域是否有 `zh-CN`
3. button wakeword 当前显示是否变成 `小智，小智`

### 第 3 步：如果机器人在线，再测真实链路

按这里执行：

- [specs/001-xiaozhi-wakeword/quickstart.md](file:///Users/yaodong/Documents/trae_projects/wire-pod/specs/001-xiaozhi-wakeword/quickstart.md)

重点记住一件事：

- 页面文案变了，不等于固件里的真实热词一定已经换掉

也就是说，如果页面已经对了，但说“小智，小智”还是不唤醒，不一定是你部署错了，也可能是机器人固件或协议限制。

## 12. 怎么看日志

### 如果你是前台直接运行

日志就在当前终端里看。

### 如果你想单独保存日志

```bash
cd ~/wire-pod
sudo ./chipper/start.sh 2>&1 | tee ~/wire-pod-run.log
```

### 如果你想过滤关键字

```bash
grep -i "error\\|panic\\|warn\\|zh-CN\\|locale" ~/wire-pod-run.log
```

## 13. 如果要做成后台服务

仓库已经带了 systemd 安装逻辑，可以在树莓派上执行：

```bash
cd ~/wire-pod
sudo ./setup.sh daemon-enable
```

启用后常用命令：

```bash
sudo systemctl start wire-pod
sudo systemctl stop wire-pod
sudo systemctl restart wire-pod
sudo systemctl status wire-pod
journalctl -u wire-pod -f
```

这个方式适合“服务已经基本稳定，准备长期挂着测试”时再用。

## 14. 当前最容易遇到的 5 个问题

### 1. `start.sh` 提示必须 root

处理：

```bash
sudo ./chipper/start.sh
```

### 2. `source.sh` 不存在

处理：

```bash
sudo ./setup.sh
```

先完成初始化。

### 3. 页面能开，但语音不工作

先不要急着怀疑页面代码，按这个顺序排查：

1. `STT_SERVICE` 是否配置对了
2. 相关 STT 依赖是否已经装好
3. setup 页面里的 STT 语言是否已经切到 `zh-CN`
4. SDK 页面里的 locale 是否已经同步
5. 机器人是否真的在线

### 4. SDK 页面打不开

检查：

```bash
ss -lntp | grep ':80'
```

如果没监听，优先回头看启动日志。

### 5. “小智，小智”页面已经显示了，但真实语音不唤醒

这次需求里，这是一个**很可能出现的正常阻塞**。

原因可能是：

1. 只是 host 侧 UI 和 locale 改了
2. `vector-cloud` 里没有真正的中文 `LanguageCode`
3. 机器人固件 / 上游热词识别并没有真的支持新短语

所以这类情况要记录为“能力边界”或“外部限制”，不是简单记成“部署失败”。

## 15. 我建议你的实际执行顺序

在你当前阶段，直接按这个顺序最稳：

1. 本机把代码同步到树莓派
2. SSH 登录树莓派
3. `cd ~/wire-pod`
4. 检查 `chipper/source.sh`
5. `sudo ./chipper/start.sh`
6. 浏览器打开 `http://192.168.1.38:8080/setup.html`
7. 再打开 `http://192.168.1.38/sdkapp`
8. 如果页面都正常，再开始机器人实测
9. 把实测结果填回 `specs/001-xiaozhi-wakeword/quickstart.md`

## 16. 你可以直接复制的命令清单

### 一键上传并重启

仓库里现在已经加好了脚本：

- `scripts/deploy_to_escapepod.sh`

直接在你本机执行：

```bash
cd /Users/yaodong/Documents/trae_projects/wire-pod
./scripts/deploy_to_escapepod.sh
```

默认行为：

1. 把本机仓库同步到 `pi@escapepod:~/wire-pod`
2. 保留树莓派上的运行时文件，例如 `chipper/source.sh`、`jdocs/`、`session-certs/`
3. 如果远端有 `wire-pod.service`，就用 `systemd restart`
4. 如果没有，就退回到 `nohup sudo ./chipper/start.sh`

常用变体：

```bash
./scripts/deploy_to_escapepod.sh --host pi@192.168.1.38
./scripts/deploy_to_escapepod.sh --dry-run
./scripts/deploy_to_escapepod.sh --skip-sync
./scripts/deploy_to_escapepod.sh --skip-restart
```

### 本机同步代码到树莓派

```bash
rsync -avz --delete \
  --exclude ".git" \
  --exclude "node_modules" \
  --exclude ".DS_Store" \
  "/Users/yaodong/Documents/trae_projects/wire-pod/" \
  pi@192.168.1.38:~/wire-pod/
```

### 登录树莓派

```bash
ssh pi@192.168.1.38
```

### 启动服务

```bash
cd ~/wire-pod
sudo ./chipper/start.sh
```

### 看端口

```bash
ss -lntp | grep -E "80|8080|8084"
```

### 后台服务方式

```bash
cd ~/wire-pod
sudo ./setup.sh daemon-enable
sudo systemctl restart wire-pod
journalctl -u wire-pod -f
```
