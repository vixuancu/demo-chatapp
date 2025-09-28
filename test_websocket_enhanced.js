const WebSocket = require('ws');

// Configuration
const WS_URL = 'ws://localhost:8080/api/v1/chat/ws';
const API_BASE = 'http://localhost:8080/api/v1';

// Test users tokens (you need to get these from login)
const USER_TOKENS = {
    user1: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mjc1MTMxOTEsImlhdCI6MTcyNzUwOTU5MSwidXNlcl91dWlkIjoiNzg1MDg0ZjUtM2E5Ni00ZTcyLTk4MzAtZjg1NDVhZDQzOTEzIn0.Ir8x2k11P0NW8jKnxgnxdABMbXNq_I8tB3vp__IOQZQ', // Replace with actual token
    user2: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mjc1MTMwMzAsImlhdCI6MTcyNzUwOTQzMCwidXNlcl91dWlkIjoiM2YzNmQwODUtNzZhNi00NDdlLThiNzEtZThjNjVmODdkZGE5In0.dAElg4L3Xokz6AtLeU0PAa6epvk3cmyXXGpnMA2rOiQ', // Replace with actual token
};

const ROOM_ID = 1; // Test room ID

class WebSocketTester {
    constructor(name, token) {
        this.name = name;
        this.token = token;
        this.ws = null;
        this.messageCount = 0;
        this.receivedMessages = [];
    }

    connect() {
        return new Promise((resolve, reject) => {
            const wsUrl = `${WS_URL}?token=${this.token}&room_id=${ROOM_ID}`;
            console.log(`🔌 [${this.name}] Connecting to ${wsUrl}`);
            
            this.ws = new WebSocket(wsUrl, ['chat']);
            
            this.ws.on('open', () => {
                console.log(`✅ [${this.name}] Connected to WebSocket`);
                resolve();
            });

            this.ws.on('message', (data) => {
                try {
                    const message = JSON.parse(data);
                    this.receivedMessages.push(message);
                    console.log(`📨 [${this.name}] Received:`, message);
                    
                    if (message.type === 'new_message') {
                        this.messageCount++;
                    }
                } catch (e) {
                    console.log(`📨 [${this.name}] Received raw:`, data.toString());
                }
            });

            this.ws.on('error', (error) => {
                console.error(`❌ [${this.name}] WebSocket error:`, error);
                reject(error);
            });

            this.ws.on('close', () => {
                console.log(`🔚 [${this.name}] WebSocket closed`);
            });
        });
    }

    joinRoom(roomId) {
        const message = {
            type: 'join_room',
            room_id: roomId
        };
        console.log(`🏠 [${this.name}] Joining room ${roomId}`);
        this.ws.send(JSON.stringify(message));
    }

    leaveRoom(roomId) {
        const message = {
            type: 'leave_room',
            room_id: roomId
        };
        console.log(`🚪 [${this.name}] Leaving room ${roomId}`);
        this.ws.send(JSON.stringify(message));
    }

    sendMessage(content) {
        const message = {
            type: 'send_message',
            room_id: ROOM_ID,
            content: content
        };
        console.log(`📤 [${this.name}] Sending: "${content}"`);
        this.ws.send(JSON.stringify(message));
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
        }
    }
}

// Test scenarios
async function runTests() {
    console.log('🧪 Starting WebSocket Chat Tests...\n');

    const user1 = new WebSocketTester('User1', USER_TOKENS.user1);
    const user2 = new WebSocketTester('User2', USER_TOKENS.user2);

    try {
        // Test 1: Connect both users
        console.log('=== Test 1: Connection ===');
        await user1.connect();
        await user2.connect();
        await sleep(1000);

        // Test 2: Join room
        console.log('\n=== Test 2: Join Room ===');
        user1.joinRoom(ROOM_ID);
        user2.joinRoom(ROOM_ID);
        await sleep(2000);

        // Test 3: Send messages
        console.log('\n=== Test 3: Single Messages ===');
        user1.sendMessage('Hello from User1!');
        await sleep(500);
        user2.sendMessage('Hi User1, this is User2!');
        await sleep(1000);

        // Test 4: Rapid fire messages (test ordering)
        console.log('\n=== Test 4: Rapid Fire Messages (Testing Order) ===');
        for (let i = 1; i <= 5; i++) {
            user1.sendMessage(`User1 message #${i}`);
            user2.sendMessage(`User2 message #${i}`);
            await sleep(100); // Small delay to simulate real usage
        }
        await sleep(2000);

        // Test 5: Leave and rejoin
        console.log('\n=== Test 5: Leave and Rejoin ===');
        user1.leaveRoom(ROOM_ID);
        await sleep(500);
        user2.sendMessage('User1 should not receive this');
        await sleep(500);
        user1.joinRoom(ROOM_ID);
        await sleep(500);
        user2.sendMessage('User1 should receive this after rejoining');
        await sleep(1000);

        // Test 6: Concurrent messaging
        console.log('\n=== Test 6: Concurrent Messages ===');
        const promises = [];
        for (let i = 1; i <= 3; i++) {
            promises.push(
                (async () => {
                    user1.sendMessage(`Concurrent User1 #${i}`);
                })()
            );
            promises.push(
                (async () => {
                    user2.sendMessage(`Concurrent User2 #${i}`);
                })()
            );
        }
        await Promise.all(promises);
        await sleep(2000);

        // Results
        console.log('\n=== Test Results ===');
        console.log(`User1 received ${user1.messageCount} messages`);
        console.log(`User2 received ${user2.messageCount} messages`);
        
        // Check for message loss
        const user1Messages = user1.receivedMessages.filter(m => m.type === 'new_message');
        const user2Messages = user2.receivedMessages.filter(m => m.type === 'new_message');
        
        console.log(`\nUser1 message IDs:`, user1Messages.map(m => m.message_id || 'no-id'));
        console.log(`User2 message IDs:`, user2Messages.map(m => m.message_id || 'no-id'));

    } catch (error) {
        console.error('❌ Test failed:', error);
    } finally {
        user1.disconnect();
        user2.disconnect();
        console.log('\n✅ Tests completed');
    }
}

// Test multiple rooms
async function testMultipleRooms() {
    console.log('\n🏢 Testing Multiple Rooms...\n');

    const user1 = new WebSocketTester('User1', USER_TOKENS.user1);
    const user2 = new WebSocketTester('User2', USER_TOKENS.user2);
    const user3 = new WebSocketTester('User3', USER_TOKENS.user1); // Same user, different connection

    try {
        await user1.connect();
        await user2.connect();
        await user3.connect();

        // User1 and User2 join room 1
        user1.joinRoom(1);
        user2.joinRoom(1);
        
        // User3 joins room 2
        user3.joinRoom(2);
        
        await sleep(1000);

        // Send messages in different rooms
        console.log('📤 Sending messages to room 1');
        user1.sendMessage('Message in room 1 from user1');
        user2.sendMessage('Message in room 1 from user2');

        console.log('📤 Sending message to room 2');
        user3.sendMessage('Message in room 2 from user3');

        await sleep(2000);

        console.log('\n=== Multi-Room Results ===');
        console.log(`User1 (room 1) received ${user1.messageCount} messages`);
        console.log(`User2 (room 1) received ${user2.messageCount} messages`);
        console.log(`User3 (room 2) received ${user3.messageCount} messages`);

    } finally {
        user1.disconnect();
        user2.disconnect();
        user3.disconnect();
    }
}

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

// Run all tests
async function main() {
    try {
        await runTests();
        await sleep(1000);
        await testMultipleRooms();
    } catch (error) {
        console.error('Test suite failed:', error);
    }
    process.exit(0);
}

main();