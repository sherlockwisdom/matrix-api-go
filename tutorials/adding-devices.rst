Adding Devices
==============

The ShortMesh API allows you to add new device connections for different messaging platforms. This guide covers how to add devices and establish WebSocket connections for real-time communication.

**Endpoint**: ``POST /{platform}/devices``

**Headers**:
- ``Authorization: Bearer YOUR_ACCESS_TOKEN``
- ``Content-Type: application/json``

**Request Body**:
.. code-block:: json

   {
     "username": "your_username"
   }

**Response**:
.. code-block:: json

   {
     "websocket_url": "ws://localhost:8080/ws/wa/john_doe"
   }

Supported Platforms
------------------

The following platforms are supported:

- ``wa`` - WhatsApp
- ``signal`` - Signal (coming soon)

WebSocket Connection
-------------------

When you add a device, the API returns a WebSocket URL that you can use to establish a real-time connection. The WebSocket connection will:

- Receive media/images from the platform bridge
- Handle login synchronization events
- Send existing active sessions if available
- Close connection when receiving nil data (indicating end of session or error)

Code Examples
-------------

.. tabs::

   .. tab:: Python

      .. code-block:: python

         import requests
         import json
         import os
         import websocket
         import threading
         from dotenv import load_dotenv

         # Load environment variables
         load_dotenv()

         # Configuration
         API_BASE_URL = os.getenv('SHORTMESH_API_URL', 'https://api.shortmesh.com')
         ACCESS_TOKEN = os.getenv('SHORTMESH_ACCESS_TOKEN')
         USERNAME = os.getenv('SHORTMESH_USERNAME')
         PLATFORM = "wa"  # wa for WhatsApp, signal for Signal

         # Validate required configuration
         if not ACCESS_TOKEN:
             raise ValueError("SHORTMESH_ACCESS_TOKEN environment variable is required")
         if not USERNAME:
             raise ValueError("SHORTMESH_USERNAME environment variable is required")

         def add_device(platform, username, access_token):
             """Add a new device for the specified platform"""
             url = f"{API_BASE_URL}/{platform}/devices"
             payload = {"username": username}
             headers = {
                 "Authorization": f"Bearer {access_token}",
                 "Content-Type": "application/json"
             }
             
             try:
                 response = requests.post(url, headers=headers, json=payload)
                 
                 if response.status_code == 200:
                     result = response.json()
                     print("✅ Device added successfully!")
                     print(f"WebSocket URL: {result['websocket_url']}")
                     return result['websocket_url']
                 else:
                     print(f"❌ Failed to add device: {response.status_code}")
                     print(response.json())
                     return None
             except Exception as e:
                 print(f"❌ Error adding device: {e}")
                 return None

         def on_message(ws, message):
             """Handle incoming WebSocket messages"""
             print(f"📨 Received message: {message}")

         def on_error(ws, error):
             """Handle WebSocket errors"""
             print(f"❌ WebSocket error: {error}")

         def on_close(ws, close_status_code, close_msg):
             """Handle WebSocket connection close"""
             print("🔌 WebSocket connection closed")

         def on_open(ws):
             """Handle WebSocket connection open"""
             print("🔌 WebSocket connection established")

         def connect_websocket(websocket_url):
             """Establish WebSocket connection"""
             ws = websocket.WebSocketApp(
                 websocket_url,
                 on_open=on_open,
                 on_message=on_message,
                 on_error=on_error,
                 on_close=on_close
             )
             
             # Run WebSocket in a separate thread
             wst = threading.Thread(target=ws.run_forever)
             wst.daemon = True
             wst.start()
             
             return ws

         if __name__ == "__main__":
             # Add device
             print("=== Adding Device ===")
             websocket_url = add_device(PLATFORM, USERNAME, ACCESS_TOKEN)
             
             if websocket_url:
                 print(f"\n=== Establishing WebSocket Connection ===")
                 print(f"Connecting to: {websocket_url}")
                 
                 # Connect to WebSocket
                 ws = connect_websocket(websocket_url)
                 
                 # Keep the main thread alive
                 try:
                     while True:
                         import time
                         time.sleep(1)
                 except KeyboardInterrupt:
                     print("\n🛑 Shutting down...")
                     ws.close()

   .. tab:: JavaScript (Node.js)

      .. code-block:: javascript

         require('dotenv').config();
         const axios = require('axios');
         const WebSocket = require('ws');

         // Configuration
         const API_BASE_URL = process.env.SHORTMESH_API_URL || 'https://api.shortmesh.com';
         const ACCESS_TOKEN = process.env.SHORTMESH_ACCESS_TOKEN;
         const USERNAME = process.env.SHORTMESH_USERNAME;
         const PLATFORM = 'wa'; // wa for WhatsApp, signal for Signal

         // Validate required configuration
         if (!ACCESS_TOKEN) {
             throw new Error('SHORTMESH_ACCESS_TOKEN environment variable is required');
         }
         if (!USERNAME) {
             throw new Error('SHORTMESH_USERNAME environment variable is required');
         }

         async function addDevice(platform, username, accessToken) {
             /** Add a new device for the specified platform */
             const url = `${API_BASE_URL}/${platform}/devices`;
             const payload = { username: username };
             const headers = {
                 'Authorization': `Bearer ${accessToken}`,
                 'Content-Type': 'application/json'
             };
             
             try {
                 const response = await axios.post(url, payload, { headers });
                 
                 console.log('✅ Device added successfully!');
                 console.log(`WebSocket URL: ${response.data.websocket_url}`);
                 return response.data.websocket_url;
             } catch (error) {
                 console.error('❌ Failed to add device:', error.response?.status);
                 if (error.response?.data) {
                     console.error(error.response.data);
                 }
                 return null;
             }
         }

         function connectWebSocket(websocketUrl) {
             /** Establish WebSocket connection */
             const ws = new WebSocket(websocketUrl);
             
             ws.on('open', function open() {
                 console.log('🔌 WebSocket connection established');
             });
             
             ws.on('message', function message(data) {
                 console.log('📨 Received message:', data.toString());
             });
             
             ws.on('error', function error(err) {
                 console.error('❌ WebSocket error:', err);
             });
             
             ws.on('close', function close(code, reason) {
                 console.log('🔌 WebSocket connection closed');
             });
             
             return ws;
         }

         // Example usage
         async function main() {
             // Add device
             console.log('=== Adding Device ===');
             const websocketUrl = await addDevice(PLATFORM, USERNAME, ACCESS_TOKEN);
             
             if (websocketUrl) {
                 console.log('\n=== Establishing WebSocket Connection ===');
                 console.log(`Connecting to: ${websocketUrl}`);
                 
                 // Connect to WebSocket
                 const ws = connectWebSocket(websocketUrl);
                 
                 // Keep the process alive
                 process.on('SIGINT', () => {
                     console.log('\n🛑 Shutting down...');
                     ws.close();
                     process.exit(0);
                 });
             }
         }

         main();

   .. tab:: PHP

      .. code-block:: php

         <?php

         // Load environment variables (requires vlucas/phpdotenv package)
         $dotenv = Dotenv\Dotenv::createImmutable(__DIR__);
         $dotenv->load();

         // Configuration
         $API_BASE_URL = $_ENV['SHORTMESH_API_URL'] ?? 'https://api.shortmesh.com';
         $ACCESS_TOKEN = $_ENV['SHORTMESH_ACCESS_TOKEN'] ?? null;
         $USERNAME = $_ENV['SHORTMESH_USERNAME'] ?? null;
         $PLATFORM = 'wa'; // wa for WhatsApp, signal for Signal

         // Validate required configuration
         if (!$ACCESS_TOKEN) {
             throw new Exception('SHORTMESH_ACCESS_TOKEN environment variable is required');
         }
         if (!$USERNAME) {
             throw new Exception('SHORTMESH_USERNAME environment variable is required');
         }

         function addDevice($apiBaseUrl, $platform, $username, $accessToken) {
             /** Add a new device for the specified platform */
             $url = $apiBaseUrl . '/' . $platform . '/devices';
             $payload = ['username' => $username];
             $headers = [
                 'Authorization: Bearer ' . $accessToken,
                 'Content-Type: application/json'
             ];
             
             $ch = curl_init();
             curl_setopt($ch, CURLOPT_URL, $url);
             curl_setopt($ch, CURLOPT_POST, true);
             curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload));
             curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
             curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
             curl_setopt($ch, CURLOPT_HTTP_VERSION, CURL_HTTP_VERSION_1_1);
             
             $response = curl_exec($ch);
             $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
             curl_close($ch);
             
             if ($httpCode == 200) {
                 $result = json_decode($response, true);
                 echo "✅ Device added successfully!\n";
                 echo "WebSocket URL: " . $result['websocket_url'] . "\n";
                 return $result['websocket_url'];
             } else {
                 echo "❌ Failed to add device: " . $httpCode . "\n";
                 echo "Response: " . $response . "\n";
                 return null;
             }
         }

         function connectWebSocket($websocketUrl) {
             /** Establish WebSocket connection using Ratchet WebSocket client */
             // Note: This requires the Ratchet WebSocket library
             // composer require cboden/ratchet
             
             echo "🔌 Connecting to WebSocket: " . $websocketUrl . "\n";
             echo "Note: WebSocket implementation requires additional libraries\n";
             echo "Consider using a WebSocket client library like Ratchet\n";
             
             // Example WebSocket connection code (requires Ratchet):
             /*
             $client = new \Ratchet\Client\WebSocket($websocketUrl);
             $client->on('message', function($msg) {
                 echo "📨 Received message: " . $msg . "\n";
             });
             $client->on('close', function() {
                 echo "🔌 WebSocket connection closed\n";
             });
             $client->run();
             */
         }

         // Example usage
         echo "=== Adding Device ===\n";
         $websocketUrl = addDevice($API_BASE_URL, $PLATFORM, $USERNAME, $ACCESS_TOKEN);
         
         if ($websocketUrl) {
             echo "\n=== Establishing WebSocket Connection ===\n";
             connectWebSocket($websocketUrl);
         }

         ?>

Error Handling
-------------

The device addition endpoint returns appropriate HTTP status codes:

- ``200 OK`` - Device added successfully
- ``400 Bad Request`` - Invalid request parameters
- ``401 Unauthorized`` - Invalid or missing Bearer token
- ``500 Internal Server Error`` - Server error

Common error responses:

.. code-block:: json

   {
     "error": "Invalid request",
     "details": "Username must be 3-32 characters"
   }

.. code-block:: json

   {
     "error": "Invalid access token",
     "details": "Token validation failed"
   }

WebSocket Best Practices
-----------------------

1. **Connection Management**: Implement proper connection handling and reconnection logic
2. **Message Handling**: Process incoming messages according to your application needs
3. **Error Handling**: Handle WebSocket errors and connection drops gracefully
4. **Resource Cleanup**: Properly close WebSocket connections when done
5. **Heartbeat**: Implement heartbeat mechanisms to keep connections alive

Example Environment Setup
------------------------

.. code-block:: bash

   # .env file
   SHORTMESH_API_URL=https://api.shortmesh.com
   SHORTMESH_ACCESS_TOKEN=your_access_token_here
   SHORTMESH_USERNAME=your_username 