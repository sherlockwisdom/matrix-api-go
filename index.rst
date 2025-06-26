.. ShortMesh documentation master file, created by
   sphinx-quickstart on Wed Jun 25 20:29:56 2025.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

ShortMesh API Documentation
==========================

Welcome to the ShortMesh API documentation! This guide will help you integrate with the ShortMesh messaging bridge API to send messages across different platforms.

.. note::
   This tutorial assumes you have already:
   
   - Created an account or logged in through the web interface
   - Retrieved your API access token from the dashboard
   - Added platforms to your devices through the web UI (which handles QR code scanning)

Getting Started
--------------

Before using the API, ensure you have:

1. **API Access Token**: Get this from your ShortMesh dashboard after logging in
2. **Platform Setup**: Add your messaging platforms (WhatsApp, Telegram, etc.) through the web interface
3. **Device Names**: Note the device names for each platform you've configured

API Base URL
^^^^^^^^^^^

All API endpoints are relative to your ShortMesh server. For local development, this is typically:

.. code-block:: text

   http://localhost:8080

Authentication
^^^^^^^^^^^^^

All API requests require authentication using a Bearer token in the Authorization header:

.. code-block:: text

   Authorization: Bearer YOUR_ACCESS_TOKEN

Adding Devices to Platforms
--------------------------

Once you've set up your platforms through the web UI, you can programmatically add devices using the API.

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
     "websocket_url": "ws://localhost:8080/ws/telegram/your_username"
   }

.. tabs::

   .. tab:: Python

      .. code-block:: python

         import requests
         import json

         # Configuration
         API_BASE_URL = "http://localhost:8080"
         ACCESS_TOKEN = "your_access_token_here"
         USERNAME = "your_username"
         PLATFORM = "wa"  # wa for WhatsApp, tg for Telegram

         # Headers
         headers = {
             "Authorization": f"Bearer {ACCESS_TOKEN}",
             "Content-Type": "application/json"
         }

         # Request payload
         payload = {
             "username": USERNAME
         }

         # Add device to platform
         response = requests.post(
             f"{API_BASE_URL}/{PLATFORM}/devices",
             headers=headers,
             json=payload
         )

         if response.status_code == 200:
             result = response.json()
             print(f"Device added successfully!")
             print(f"WebSocket URL: {result['websocket_url']}")
         else:
             print(f"Error: {response.status_code}")
             print(response.json())

   .. tab:: JavaScript (Node.js)

      .. code-block:: javascript

         const axios = require('axios');

         // Configuration
         const API_BASE_URL = 'http://localhost:8080';
         const ACCESS_TOKEN = 'your_access_token_here';
         const USERNAME = 'your_username';
         const PLATFORM = 'wa'; // wa for WhatsApp, tg for Telegram

         // Headers
         const headers = {
             'Authorization': `Bearer ${ACCESS_TOKEN}`,
             'Content-Type': 'application/json'
         };

         // Request payload
         const payload = {
             username: USERNAME
         };

         // Add device to platform
         async function addDevice() {
             try {
                 const response = await axios.post(
                     `${API_BASE_URL}/${PLATFORM}/devices`,
                     payload,
                     { headers }
                 );
                 
                 console.log('Device added successfully!');
                 console.log(`WebSocket URL: ${response.data.websocket_url}`);
             } catch (error) {
                 console.error('Error:', error.response?.status);
                 console.error(error.response?.data);
             }
         }

         addDevice();

   .. tab:: Go

      .. code-block:: go

         package main

         import (
             "bytes"
             "encoding/json"
             "fmt"
             "io"
             "net/http"
         )

         // Configuration
         const (
             API_BASE_URL = "http://localhost:8080"
             ACCESS_TOKEN = "your_access_token_here"
             USERNAME     = "your_username"
             PLATFORM     = "wa" // wa for WhatsApp, tg for Telegram
         )

         type DeviceRequest struct {
             Username string `json:"username"`
         }

         type DeviceResponse struct {
             WebsocketURL string `json:"websocket_url"`
         }

         func addDevice() error {
             // Request payload
             payload := DeviceRequest{
                 Username: USERNAME,
             }

             jsonData, err := json.Marshal(payload)
             if err != nil {
                 return fmt.Errorf("error marshaling JSON: %v", err)
             }

             // Create request
             req, err := http.NewRequest("POST", 
                 fmt.Sprintf("%s/%s/devices", API_BASE_URL, PLATFORM), 
                 bytes.NewBuffer(jsonData))
             if err != nil {
                 return fmt.Errorf("error creating request: %v", err)
             }

             // Set headers
             req.Header.Set("Authorization", "Bearer "+ACCESS_TOKEN)
             req.Header.Set("Content-Type", "application/json")

             // Send request
             client := &http.Client{}
             resp, err := client.Do(req)
             if err != nil {
                 return fmt.Errorf("error sending request: %v", err)
             }
             defer resp.Body.Close()

             if resp.StatusCode == 200 {
                 body, _ := io.ReadAll(resp.Body)
                 var result DeviceResponse
                 json.Unmarshal(body, &result)
                 fmt.Printf("Device added successfully!\n")
                 fmt.Printf("WebSocket URL: %s\n", result.WebsocketURL)
             } else {
                 body, _ := io.ReadAll(resp.Body)
                 fmt.Printf("Error: %d\n", resp.StatusCode)
                 fmt.Printf("Response: %s\n", string(body))
             }

             return nil
         }

         func main() {
             if err := addDevice(); err != nil {
                 fmt.Printf("Error: %v\n", err)
             }
         }

   .. tab:: PHP

      .. code-block:: php

         <?php

         // Configuration
         $API_BASE_URL = 'http://localhost:8080';
         $ACCESS_TOKEN = 'your_access_token_here';
         $USERNAME = 'your_username';
         $PLATFORM = 'wa'; // wa for WhatsApp, tg for Telegram

         // Request payload
         $payload = [
             'username' => $USERNAME
         ];

         // Initialize cURL
         $ch = curl_init();

         // Set cURL options
         curl_setopt($ch, CURLOPT_URL, $API_BASE_URL . '/' . $PLATFORM . '/devices');
         curl_setopt($ch, CURLOPT_POST, true);
         curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload));
         curl_setopt($ch, CURLOPT_HTTPHEADER, [
             'Authorization: Bearer ' . $ACCESS_TOKEN,
             'Content-Type: application/json'
         ]);
         curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
         curl_setopt($ch, CURLOPT_HTTP_VERSION, CURL_HTTP_VERSION_1_1);

         // Execute request
         $response = curl_exec($ch);
         $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
         curl_close($ch);

         // Handle response
         if ($httpCode == 200) {
             $result = json_decode($response, true);
             echo "Device added successfully!\n";
             echo "WebSocket URL: " . $result['websocket_url'] . "\n";
         } else {
             echo "Error: " . $httpCode . "\n";
             echo "Response: " . $response . "\n";
         }

         ?>

Sending Messages
---------------

Once you have devices set up, you can send messages to contacts using their phone numbers in E.164 format.

**Endpoint**: ``POST /{platform}/message/{contact}``

**Headers**:
- ``Authorization: Bearer YOUR_ACCESS_TOKEN``
- ``Content-Type: application/json``

**Request Body**:
.. code-block:: json

   {
     "username": "your_username",
     "message": "Hello from ShortMesh!",
     "device_name": "wa123456789"
   }

**Response**:
.. code-block:: json

   {
     "contact": "+1234567890",
     "event_id": "$1234567890abcdef",
     "message": "Hello from ShortMesh!",
     "status": "sent"
   }

.. tabs::

   .. tab:: Python

      .. code-block:: python

         import requests
         import json

         # Configuration
         API_BASE_URL = "http://localhost:8080"
         ACCESS_TOKEN = "your_access_token_here"
         USERNAME = "your_username"
         PLATFORM = "wa"  # wa for WhatsApp, tg for Telegram
         DEVICE_NAME = "wa123456789"  # Your device name
         CONTACT = "1234567890"  # Phone number without + (E.164 format)

         # Headers
         headers = {
             "Authorization": f"Bearer {ACCESS_TOKEN}",
             "Content-Type": "application/json"
         }

         # Request payload
         payload = {
             "username": USERNAME,
             "message": "Hello from ShortMesh!",
             "device_name": DEVICE_NAME
         }

         # Send message
         response = requests.post(
             f"{API_BASE_URL}/{PLATFORM}/message/{CONTACT}",
             headers=headers,
             json=payload
         )

         if response.status_code == 200:
             result = response.json()
             print(f"Message sent successfully!")
             print(f"Contact: {result['contact']}")
             print(f"Event ID: {result['event_id']}")
             print(f"Status: {result['status']}")
         else:
             print(f"Error: {response.status_code}")
             print(response.json())

   .. tab:: JavaScript (Node.js)

      .. code-block:: javascript

         const axios = require('axios');

         // Configuration
         const API_BASE_URL = 'http://localhost:8080';
         const ACCESS_TOKEN = 'your_access_token_here';
         const USERNAME = 'your_username';
         const PLATFORM = 'wa'; // wa for WhatsApp, tg for Telegram
         const DEVICE_NAME = 'wa123456789'; // Your device name
         const CONTACT = '1234567890'; // Phone number without + (E.164 format)

         // Headers
         const headers = {
             'Authorization': `Bearer ${ACCESS_TOKEN}`,
             'Content-Type': 'application/json'
         };

         // Request payload
         const payload = {
             username: USERNAME,
             message: 'Hello from ShortMesh!',
             device_name: DEVICE_NAME
         };

         // Send message
         async function sendMessage() {
             try {
                 const response = await axios.post(
                     `${API_BASE_URL}/${PLATFORM}/message/${CONTACT}`,
                     payload,
                     { headers }
                 );
                 
                 console.log('Message sent successfully!');
                 console.log(`Contact: ${response.data.contact}`);
                 console.log(`Event ID: ${response.data.event_id}`);
                 console.log(`Status: ${response.data.status}`);
             } catch (error) {
                 console.error('Error:', error.response?.status);
                 console.error(error.response?.data);
             }
         }

         sendMessage();

   .. tab:: Go

      .. code-block:: go

         package main

         import (
             "bytes"
             "encoding/json"
             "fmt"
             "io"
             "net/http"
         )

         // Configuration
         const (
             API_BASE_URL = "http://localhost:8080"
             ACCESS_TOKEN = "your_access_token_here"
             USERNAME     = "your_username"
             PLATFORM     = "wa" // wa for WhatsApp, tg for Telegram
             DEVICE_NAME  = "wa123456789" // Your device name
             CONTACT      = "1234567890"  // Phone number without + (E.164 format)
         )

         type MessageRequest struct {
             Username   string `json:"username"`
             Message    string `json:"message"`
             DeviceName string `json:"device_name"`
         }

         type MessageResponse struct {
             Contact string `json:"contact"`
             EventID string `json:"event_id"`
             Message string `json:"message"`
             Status  string `json:"status"`
         }

         func sendMessage() error {
             // Request payload
             payload := MessageRequest{
                 Username:   USERNAME,
                 Message:    "Hello from ShortMesh!",
                 DeviceName: DEVICE_NAME,
             }

             jsonData, err := json.Marshal(payload)
             if err != nil {
                 return fmt.Errorf("error marshaling JSON: %v", err)
             }

             // Create request
             req, err := http.NewRequest("POST", 
                 fmt.Sprintf("%s/%s/message/%s", API_BASE_URL, PLATFORM, CONTACT), 
                 bytes.NewBuffer(jsonData))
             if err != nil {
                 return fmt.Errorf("error creating request: %v", err)
             }

             // Set headers
             req.Header.Set("Authorization", "Bearer "+ACCESS_TOKEN)
             req.Header.Set("Content-Type", "application/json")

             // Send request
             client := &http.Client{}
             resp, err := client.Do(req)
             if err != nil {
                 return fmt.Errorf("error sending request: %v", err)
             }
             defer resp.Body.Close()

             if resp.StatusCode == 200 {
                 body, _ := io.ReadAll(resp.Body)
                 var result MessageResponse
                 json.Unmarshal(body, &result)
                 fmt.Printf("Message sent successfully!\n")
                 fmt.Printf("Contact: %s\n", result.Contact)
                 fmt.Printf("Event ID: %s\n", result.EventID)
                 fmt.Printf("Status: %s\n", result.Status)
             } else {
                 body, _ := io.ReadAll(resp.Body)
                 fmt.Printf("Error: %d\n", resp.StatusCode)
                 fmt.Printf("Response: %s\n", string(body))
             }

             return nil
         }

         func main() {
             if err := sendMessage(); err != nil {
                 fmt.Printf("Error: %v\n", err)
             }
         }

   .. tab:: PHP

      .. code-block:: php

         <?php

         // Configuration
         $API_BASE_URL = 'http://localhost:8080';
         $ACCESS_TOKEN = 'your_access_token_here';
         $USERNAME = 'your_username';
         $PLATFORM = 'wa'; // wa for WhatsApp, tg for Telegram
         $DEVICE_NAME = 'wa123456789'; // Your device name
         $CONTACT = '1234567890'; // Phone number without + (E.164 format)

         // Request payload
         $payload = [
             'username' => $USERNAME,
             'message' => 'Hello from ShortMesh!',
             'device_name' => $DEVICE_NAME
         ];

         // Initialize cURL
         $ch = curl_init();

         // Set cURL options
         curl_setopt($ch, CURLOPT_URL, $API_BASE_URL . '/' . $PLATFORM . '/message/' . $CONTACT);
         curl_setopt($ch, CURLOPT_POST, true);
         curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload));
         curl_setopt($ch, CURLOPT_HTTPHEADER, [
             'Authorization: Bearer ' . $ACCESS_TOKEN,
             'Content-Type: application/json'
         ]);
         curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
         curl_setopt($ch, CURLOPT_HTTP_VERSION, CURL_HTTP_VERSION_1_1);

         // Execute request
         $response = curl_exec($ch);
         $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
         curl_close($ch);

         // Handle response
         if ($httpCode == 200) {
             $result = json_decode($response, true);
             echo "Message sent successfully!\n";
             echo "Contact: " . $result['contact'] . "\n";
             echo "Event ID: " . $result['event_id'] . "\n";
             echo "Status: " . $result['status'] . "\n";
         } else {
             echo "Error: " . $httpCode . "\n";
             echo "Response: " . $response . "\n";
         }

         ?>

Listing Devices
--------------

You can list all devices for a specific platform to see what devices are available for messaging.

**Endpoint**: ``POST /{platform}/list/devices``

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
     "devices": [
       {
         "device_id": "wa123456789",
         "platform": "wa",
         "status": "active"
       }
     ]
   }

.. tabs::

   .. tab:: Python

      .. code-block:: python

         import requests
         import json

         # Configuration
         API_BASE_URL = "http://localhost:8080"
         ACCESS_TOKEN = "your_access_token_here"
         USERNAME = "your_username"
         PLATFORM = "wa"  # wa for WhatsApp, tg for Telegram

         # Headers
         headers = {
             "Authorization": f"Bearer {ACCESS_TOKEN}",
             "Content-Type": "application/json"
         }

         # Request payload
         payload = {
             "username": USERNAME
         }

         # List devices
         response = requests.post(
             f"{API_BASE_URL}/{PLATFORM}/list/devices",
             headers=headers,
             json=payload
         )

         if response.status_code == 200:
             result = response.json()
             print(f"Devices for {PLATFORM}:")
             for device in result['devices']:
                 print(f"- Device ID: {device['device_id']}")
                 print(f"  Status: {device['status']}")
         else:
             print(f"Error: {response.status_code}")
             print(response.json())

   .. tab:: JavaScript (Node.js)

      .. code-block:: javascript

         const axios = require('axios');

         // Configuration
         const API_BASE_URL = 'http://localhost:8080';
         const ACCESS_TOKEN = 'your_access_token_here';
         const USERNAME = 'your_username';
         const PLATFORM = 'wa'; // wa for WhatsApp, tg for Telegram

         // Headers
         const headers = {
             'Authorization': `Bearer ${ACCESS_TOKEN}`,
             'Content-Type': 'application/json'
         };

         // Request payload
         const payload = {
             username: USERNAME
         };

         // List devices
         async function listDevices() {
             try {
                 const response = await axios.post(
                     `${API_BASE_URL}/${PLATFORM}/list/devices`,
                     payload,
                     { headers }
                 );
                 
                 console.log(`Devices for ${PLATFORM}:`);
                 response.data.devices.forEach(device => {
                     console.log(`- Device ID: ${device.device_id}`);
                     console.log(`  Status: ${device.status}`);
                 });
             } catch (error) {
                 console.error('Error:', error.response?.status);
                 console.error(error.response?.data);
             }
         }

         listDevices();

   .. tab:: Go

      .. code-block:: go

         package main

         import (
             "bytes"
             "encoding/json"
             "fmt"
             "io"
             "net/http"
         )

         // Configuration
         const (
             API_BASE_URL = "http://localhost:8080"
             ACCESS_TOKEN = "your_access_token_here"
             USERNAME     = "your_username"
             PLATFORM     = "wa" // wa for WhatsApp, tg for Telegram
         )

         type DeviceListRequest struct {
             Username string `json:"username"`
         }

         type Device struct {
             DeviceID string `json:"device_id"`
             Platform string `json:"platform"`
             Status   string `json:"status"`
         }

         type DeviceListResponse struct {
             Devices []Device `json:"devices"`
         }

         func listDevices() error {
             // Request payload
             payload := DeviceListRequest{
                 Username: USERNAME,
             }

             jsonData, err := json.Marshal(payload)
             if err != nil {
                 return fmt.Errorf("error marshaling JSON: %v", err)
             }

             // Create request
             req, err := http.NewRequest("POST", 
                 fmt.Sprintf("%s/%s/list/devices", API_BASE_URL, PLATFORM), 
                 bytes.NewBuffer(jsonData))
             if err != nil {
                 return fmt.Errorf("error creating request: %v", err)
             }

             // Set headers
             req.Header.Set("Authorization", "Bearer "+ACCESS_TOKEN)
             req.Header.Set("Content-Type", "application/json")

             // Send request
             client := &http.Client{}
             resp, err := client.Do(req)
             if err != nil {
                 return fmt.Errorf("error sending request: %v", err)
             }
             defer resp.Body.Close()

             if resp.StatusCode == 200 {
                 body, _ := io.ReadAll(resp.Body)
                 var result DeviceListResponse
                 json.Unmarshal(body, &result)
                 fmt.Printf("Devices for %s:\n", PLATFORM)
                 for _, device := range result.Devices {
                     fmt.Printf("- Device ID: %s\n", device.DeviceID)
                     fmt.Printf("  Status: %s\n", device.Status)
                 }
             } else {
                 body, _ := io.ReadAll(resp.Body)
                 fmt.Printf("Error: %d\n", resp.StatusCode)
                 fmt.Printf("Response: %s\n", string(body))
             }

             return nil
         }

         func main() {
             if err := listDevices(); err != nil {
                 fmt.Printf("Error: %v\n", err)
             }
         }

   .. tab:: PHP

      .. code-block:: php

         <?php

         // Configuration
         $API_BASE_URL = 'http://localhost:8080';
         $ACCESS_TOKEN = 'your_access_token_here';
         $USERNAME = 'your_username';
         $PLATFORM = 'wa'; // wa for WhatsApp, tg for Telegram

         // Request payload
         $payload = [
             'username' => $USERNAME
         ];

         // Initialize cURL
         $ch = curl_init();

         // Set cURL options
         curl_setopt($ch, CURLOPT_URL, $API_BASE_URL . '/' . $PLATFORM . '/list/devices');
         curl_setopt($ch, CURLOPT_POST, true);
         curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload));
         curl_setopt($ch, CURLOPT_HTTPHEADER, [
             'Authorization: Bearer ' . $ACCESS_TOKEN,
             'Content-Type: application/json'
         ]);
         curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
         curl_setopt($ch, CURLOPT_HTTP_VERSION, CURL_HTTP_VERSION_1_1);

         // Execute request
         $response = curl_exec($ch);
         $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
         curl_close($ch);

         // Handle response
         if ($httpCode == 200) {
             $result = json_decode($response, true);
             echo "Devices for $PLATFORM:\n";
             foreach ($result['devices'] as $device) {
                 echo "- Device ID: " . $device['device_id'] . "\n";
                 echo "  Status: " . $device['status'] . "\n";
             }
         } else {
             echo "Error: " . $httpCode . "\n";
             echo "Response: " . $response . "\n";
         }

         ?>

Platform Codes
-------------

The following platform codes are supported:

- ``wa`` - WhatsApp
- ``tg`` - Telegram
- ``sg`` - Signal (coming soon)

Error Handling
-------------

The API returns appropriate HTTP status codes and error messages:

- ``200 OK`` - Request successful
- ``400 Bad Request`` - Invalid request parameters
- ``401 Unauthorized`` - Invalid or missing access token
- ``500 Internal Server Error`` - Server error

Error responses include details about what went wrong:

.. code-block:: json

   {
     "error": "Invalid request",
     "details": "Username must be 3-32 characters"
   }

Best Practices
-------------

1. **Store Access Tokens Securely**: Never hardcode access tokens in your source code
2. **Use Environment Variables**: Store sensitive configuration in environment variables
3. **Handle Errors Gracefully**: Always check response status codes and handle errors appropriately
4. **Validate Input**: Ensure phone numbers are in E.164 format (without + prefix)
5. **Rate Limiting**: Be mindful of API rate limits in production environments

Example Environment Setup
^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: bash

   # .env file
   SHORTMESH_API_URL=http://localhost:8080
   SHORTMESH_ACCESS_TOKEN=your_access_token_here
   SHORTMESH_USERNAME=your_username

.. toctree::
   :maxdepth: 2
   :caption: Contents:
