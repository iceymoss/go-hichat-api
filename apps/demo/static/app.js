// WebRTC 配置
const configuration = {
    iceServers: [
        { urls: 'stun:stun.l.google.com:19302' },
        { urls: 'stun:stun1.l.google.com:19302' }
    ]
};

// 全局变量
let ws = null;
let pc = null;
let localStream = null;
let clientId = null;
let peerId = null;
let isMuted = false;
let isVideoOff = false;

// DOM 元素
const statusDot = document.getElementById('statusDot');
const statusText = document.getElementById('statusText');
const clientIdElem = document.getElementById('clientId');
const peerIdElem = document.getElementById('peerId');
const localVideo = document.getElementById('localVideo');
const remoteVideo = document.getElementById('remoteVideo');
const remoteVideoWrapper = document.getElementById('remoteVideoWrapper');
const remoteVideoPlaceholder = document.getElementById('remoteVideoPlaceholder');
const messages = document.getElementById('messages');
const startBtn = document.getElementById('startBtn');
const hangupBtn = document.getElementById('hangupBtn');
const muteBtn = document.getElementById('muteBtn');
const videoBtn = document.getElementById('videoBtn');

// 添加日志消息
function addMessage(text, type = 'info') {
    const msg = document.createElement('div');
    msg.className = `message ${type}`;
    msg.textContent = `[${new Date().toLocaleTimeString()}] ${text}`;
    messages.appendChild(msg);
    messages.scrollTop = messages.scrollHeight;
}

// 更新状态
function updateStatus(status, statusClass) {
    statusText.textContent = status;
    statusDot.className = `status-dot ${statusClass}`;
}

// 开始通话
async function startCall() {
    try {
        addMessage('正在获取摄像头和麦克风权限...', 'info');
        
        // 获取本地媒体流
        localStream = await navigator.mediaDevices.getUserMedia({
            video: {
                width: { ideal: 1280 },
                height: { ideal: 720 }
            },
            audio: true
        });

        localVideo.srcObject = localStream;
        addMessage('✅ 成功获取本地媒体流', 'success');

        // 连接信令服务器
        connectSignaling();

        // 更新按钮状态
        startBtn.disabled = true;
        hangupBtn.disabled = false;
        muteBtn.disabled = false;
        videoBtn.disabled = false;

    } catch (err) {
        console.error('获取媒体流失败:', err);
        addMessage('❌ 获取摄像头/麦克风失败: ' + err.message, 'error');
    }
}

// 连接信令服务器
function connectSignaling() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    
    addMessage('正在连接信令服务器...', 'info');
    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
        addMessage('✅ 已连接到信令服务器', 'success');
        updateStatus('已连接 - 等待配对', 'connected');
    };

    ws.onmessage = async (event) => {
        const message = JSON.parse(event.data);
        console.log('收到消息:', message);

        switch (message.type) {
            case 'welcome':
                clientId = message.fromId;
                clientIdElem.textContent = clientId.substring(0, 8);
                addMessage(`客户端 ID: ${clientId.substring(0, 8)}`, 'info');
                break;

            case 'matched':
                peerId = message.toId;
                peerIdElem.textContent = peerId.substring(0, 8);
                addMessage('✅ 配对成功！正在建立连接...', 'success');
                updateStatus('已配对 - 建立连接中', 'matching');
                
                // 创建 PeerConnection
                await createPeerConnection();
                
                // 只有呼叫方（caller）才创建 offer
                if (message.sdp === 'caller') {
                    addMessage('作为呼叫方发起连接...', 'info');
                    const offer = await pc.createOffer();
                    await pc.setLocalDescription(offer);
                    
                    ws.send(JSON.stringify({
                        type: 'offer',
                        sdp: offer.sdp
                    }));
                } else {
                    addMessage('等待对方发起连接...', 'info');
                }
                break;

            case 'offer':
                addMessage('收到对方的连接请求', 'info');
                await createPeerConnection();
                
                await pc.setRemoteDescription(new RTCSessionDescription({
                    type: 'offer',
                    sdp: message.sdp
                }));
                
                const answer = await pc.createAnswer();
                await pc.setLocalDescription(answer);
                
                ws.send(JSON.stringify({
                    type: 'answer',
                    sdp: answer.sdp
                }));
                break;

            case 'answer':
                addMessage('收到对方的应答', 'info');
                await pc.setRemoteDescription(new RTCSessionDescription({
                    type: 'answer',
                    sdp: message.sdp
                }));
                break;

            case 'candidate':
                if (message.candidate && pc) {
                    try {
                        await pc.addIceCandidate(new RTCIceCandidate(JSON.parse(message.candidate)));
                    } catch (err) {
                        console.error('添加 ICE candidate 失败:', err);
                    }
                }
                break;

            case 'peer-disconnected':
                addMessage('⚠️ 对方已断开连接', 'error');
                handlePeerDisconnected();
                break;
        }
    };

    ws.onerror = (error) => {
        console.error('WebSocket 错误:', error);
        addMessage('❌ 信令服务器连接错误', 'error');
        updateStatus('连接错误', '');
    };

    ws.onclose = () => {
        addMessage('⚠️ 信令服务器连接已关闭', 'error');
        updateStatus('未连接', '');
    };
}

// 创建 PeerConnection
async function createPeerConnection() {
    if (pc) {
        return;
    }

    addMessage('正在创建 WebRTC 连接...', 'info');
    pc = new RTCPeerConnection(configuration);

    // 添加本地流
    localStream.getTracks().forEach(track => {
        pc.addTrack(track, localStream);
    });

    // 处理远程流
    pc.ontrack = (event) => {
        console.log('🎬 收到远程流 track:', event.track.kind, 'streams:', event.streams.length);
        
        if (event.streams && event.streams[0]) {
            const stream = event.streams[0];
            console.log('🎬 远程流 ID:', stream.id, '包含轨道:', stream.getTracks().length);
            
            // 设置远程视频源
            remoteVideo.srcObject = stream;
            remoteVideo.style.display = 'block';
            remoteVideoPlaceholder.style.display = 'none';
            remoteVideoWrapper.classList.remove('empty');
            
            // 确保视频自动播放
            remoteVideo.play().catch(err => {
                console.error('远程视频播放失败:', err);
            });
            
            addMessage('✅ 已接收到对方的视频流', 'success');
            updateStatus('通话中', 'connected');
        }
    };

    // 处理 ICE candidate
    pc.onicecandidate = (event) => {
        if (event.candidate) {
            ws.send(JSON.stringify({
                type: 'candidate',
                candidate: JSON.stringify(event.candidate)
            }));
        }
    };

    // 连接状态变化
    pc.onconnectionstatechange = () => {
        console.log('连接状态:', pc.connectionState);
        
        switch (pc.connectionState) {
            case 'connected':
                addMessage('✅ WebRTC 连接已建立', 'success');
                updateStatus('通话中', 'connected');
                break;
            case 'disconnected':
                addMessage('⚠️ 连接已断开', 'error');
                updateStatus('连接断开', '');
                break;
            case 'failed':
                addMessage('❌ 连接失败', 'error');
                updateStatus('连接失败', '');
                handlePeerDisconnected();
                break;
            case 'closed':
                addMessage('连接已关闭', 'info');
                break;
        }
    };

    // ICE 连接状态
    pc.oniceconnectionstatechange = () => {
        console.log('🧊 ICE 连接状态:', pc.iceConnectionState);
        addMessage(`ICE 状态: ${pc.iceConnectionState}`, 'info');
        
        if (pc.iceConnectionState === 'connected' || pc.iceConnectionState === 'completed') {
            addMessage('✅ ICE 连接成功', 'success');
        } else if (pc.iceConnectionState === 'failed') {
            addMessage('❌ ICE 连接失败', 'error');
        }
    };
    
    // ICE 候选收集状态
    pc.onicegatheringstatechange = () => {
        console.log('🧊 ICE 收集状态:', pc.iceGatheringState);
    };
}

// 处理对端断开
function handlePeerDisconnected() {
    if (remoteVideo.srcObject) {
        remoteVideo.srcObject.getTracks().forEach(track => track.stop());
        remoteVideo.srcObject = null;
    }
    remoteVideo.style.display = 'none';
    remoteVideoPlaceholder.style.display = 'block';
    remoteVideoPlaceholder.textContent = '对方已离开，等待新的连接...';
    remoteVideoWrapper.classList.add('empty');
    
    if (pc) {
        pc.close();
        pc = null;
    }
    
    peerId = null;
    peerIdElem.textContent = '-';
    updateStatus('已连接 - 等待配对', 'connected');
}

// 挂断
function hangup() {
    addMessage('正在结束通话...', 'info');

    // 关闭 WebSocket
    if (ws) {
        ws.close();
        ws = null;
    }

    // 关闭 PeerConnection
    if (pc) {
        pc.close();
        pc = null;
    }

    // 停止本地流
    if (localStream) {
        localStream.getTracks().forEach(track => track.stop());
        localStream = null;
    }

    // 停止远程流
    if (remoteVideo.srcObject) {
        remoteVideo.srcObject.getTracks().forEach(track => track.stop());
        remoteVideo.srcObject = null;
    }

    // 重置界面
    localVideo.srcObject = null;
    remoteVideo.srcObject = null;
    remoteVideo.style.display = 'none';
    remoteVideoPlaceholder.style.display = 'block';
    remoteVideoPlaceholder.textContent = '等待对方加入...';
    remoteVideoWrapper.classList.add('empty');

    clientId = null;
    peerId = null;
    clientIdElem.textContent = '-';
    peerIdElem.textContent = '-';

    // 重置按钮
    startBtn.disabled = false;
    hangupBtn.disabled = true;
    muteBtn.disabled = true;
    videoBtn.disabled = true;

    isMuted = false;
    isVideoOff = false;
    document.getElementById('muteText').textContent = '静音';
    document.getElementById('videoText').textContent = '关闭视频';

    updateStatus('未连接', '');
    addMessage('✅ 通话已结束', 'success');
}

// 切换静音
function toggleMute() {
    if (!localStream) return;

    const audioTrack = localStream.getAudioTracks()[0];
    if (audioTrack) {
        isMuted = !isMuted;
        audioTrack.enabled = !isMuted;
        
        document.getElementById('muteIcon').textContent = isMuted ? '🔇' : '🎤';
        document.getElementById('muteText').textContent = isMuted ? '取消静音' : '静音';
        
        addMessage(isMuted ? '🔇 已静音' : '🎤 已取消静音', 'info');
    }
}

// 切换视频
function toggleVideo() {
    if (!localStream) return;

    const videoTrack = localStream.getVideoTracks()[0];
    if (videoTrack) {
        isVideoOff = !isVideoOff;
        videoTrack.enabled = !isVideoOff;
        
        document.getElementById('videoIcon').textContent = isVideoOff ? '📵' : '📹';
        document.getElementById('videoText').textContent = isVideoOff ? '开启视频' : '关闭视频';
        
        addMessage(isVideoOff ? '📵 已关闭视频' : '📹 已开启视频', 'info');
    }
}

// 页面卸载时清理
window.addEventListener('beforeunload', () => {
    if (ws) ws.close();
    if (pc) pc.close();
    if (localStream) {
        localStream.getTracks().forEach(track => track.stop());
    }
});

