#!/bin/bash

echo "🚀 启动视频通话 Demo 服务..."
echo ""

# 切换到脚本所在目录
cd "$(dirname "$0")"

# 检查配置文件
if [ ! -f "etc/demo.yaml" ]; then
    echo "❌ 错误: 找不到配置文件 etc/demo.yaml"
    exit 1
fi

# 停止旧服务
echo "⏹️  停止旧服务..."
pkill -f "demo.go" 2>/dev/null || true
sleep 1

# 启动服务
echo "▶️  启动新服务..."
nohup go run demo.go -f etc/demo.yaml > /tmp/demo.log 2>&1 &
sleep 2

# 检查服务状态
if lsof -i :8890 | grep LISTEN > /dev/null; then
    echo ""
    echo "✅ 服务启动成功！"
    echo ""
    echo "🎥 视频通话服务启动成功!"
    echo "📡 信令服务器: ws://localhost:8890/ws"
    echo "🌐 访问地址: http://localhost:8890"
    echo "🔧 状态接口: http://localhost:8890/status"
    echo "📝 日志文件: /tmp/demo.log"
    echo ""
    echo "💡 提示: 打开两个浏览器窗口访问 http://localhost:8890 开始测试！"
else
    echo ""
    echo "❌ 服务启动失败，请查看日志: /tmp/demo.log"
    exit 1
fi

