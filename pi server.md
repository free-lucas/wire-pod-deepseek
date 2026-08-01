# 佳兆业
ssh pi@192.168.1.38
密码：123456

# v2raya账号
http://192.168.1.38:2017/
pi
密码：123456


# 查看最近日志
sudo journalctl -u wire-pod.service --no-pager -n 50

# 实时追踪日志（类似 tail -f）
sudo journalctl -u wire-pod.service -f

# 查看最近 5 分钟的日志
sudo journalctl -u wire-pod.service --no-pager --since '5 minutes ago'

# 只看含有关键词的日志（如 error、LLM、DeepSeek）
sudo journalctl -u wire-pod.service --no-pager | grep -i 'error\|llm\|deepseek'