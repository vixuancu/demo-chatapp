const WebSocket = require('ws');

// Test script để verify WebSocket functionality
async function testWebSocket() {
    console.log('🧪 Testing WebSocket functionality...');
    
    // Get token first
    const userResponse = await fetch('http://localhost:8081/api/v1/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            user_fullname: 'WebSocket Test User',
            user_email: 'ws_test@example.com',
            user_password: 'password123'
        })
    }).catch(() => {
        // User might already exist, try login
        return fetch('http://localhost:8081/api/v1/auth/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                user_email: 'ws_test@example.com',
                user_password: 'password123'
            })
        });
    });
    
    if (!userResponse.ok) {
        console.error('❌ Failed to get user token');
        return;
    }
    
    const userData = await userResponse.json();
    const token = userData.data.token;
    console.log('✅ Got user token');
    
    // Create a room
    const roomResponse = await fetch('http://localhost:8081/api/v1/rooms', {
        method: 'POST',
        headers: { 
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
            room_name: 'WebSocket Test Room',
            room_description: 'Testing WebSocket'
        })
    });
    
    if (!roomResponse.ok) {
        console.error('❌ Failed to create room');
        return;
    }
    
    const roomData = await roomResponse.json();
    const roomId = roomData.data.room_id;
    console.log(`✅ Created room: ${roomId}`);
    
    // Test WebSocket connection
    const wsUrl = `ws://localhost:8081/api/v1/chat/ws?token=${token}&room_id=${roomId}`;
    const ws = new WebSocket(wsUrl);
    
    ws.on('open', () => {
        console.log('✅ WebSocket connected');
        
        // Send a test message
        const message = {
            type: 'send_message',
            room_id: roomId,
            content: 'Hello from WebSocket test!'
        };
        
        console.log('📤 Sending message:', message);
        ws.send(JSON.stringify(message));
    });
    
    ws.on('message', (data) => {
        try {
            const message = JSON.parse(data.toString());
            console.log('📥 Received message:', message);
            
            if (message.type === 'new_message') {
                console.log('✅ Message broadcast received successfully');
                ws.close();
            }
        } catch (err) {
            console.error('❌ Failed to parse message:', err);
        }
    });
    
    ws.on('error', (error) => {
        console.error('❌ WebSocket error:', error.message);
    });
    
    ws.on('close', () => {
        console.log('🔌 WebSocket connection closed');
        console.log('✅ Test completed successfully!');
        process.exit(0);
    });
    
    // Timeout after 10 seconds
    setTimeout(() => {
        console.log('⏰ Test timeout');
        ws.close();
        process.exit(1);
    }, 10000);
}

testWebSocket().catch(console.error);