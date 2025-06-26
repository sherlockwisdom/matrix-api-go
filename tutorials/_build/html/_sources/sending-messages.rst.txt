Sending Messages
===============

Once you have devices set up through the web UI, you can send messages to contacts using their phone numbers in E.164 format.

**Endpoint**: ``POST /{platform}/message/{contact}``

**Headers**:
- ``Authorization: Bearer YOUR_ACCESS_TOKEN``
- ``Content-Type: application/json``

**Request Body**: 
The request body contains your username, message content, and device name. For detailed request body specifications, see the complete API documentation at https://api.shortmesh.com/

**Response**: 
The response contains the contact information, event ID, message content, and delivery status. For detailed response specifications, see the complete API documentation at https://api.shortmesh.com/

Code Examples
-------------

.. tabs::

   .. tab:: Python

      .. code-block:: python

         import requests
         import json

         # Configuration
         API_BASE_URL = "https://api.shortmesh.com"
         ACCESS_TOKEN = "your_access_token_here"
         USERNAME = "your_username"
         PLATFORM = "wa"  # wa for WhatsApp, signal for Signal
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
         const API_BASE_URL = 'https://api.shortmesh.com';
         const ACCESS_TOKEN = 'your_access_token_here';
         const USERNAME = 'your_username';
         const PLATFORM = 'wa'; // wa for WhatsApp, signal for Signal
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
             API_BASE_URL = "https://api.shortmesh.com"
             ACCESS_TOKEN = "your_access_token_here"
             USERNAME     = "your_username"
             PLATFORM     = "wa" // wa for WhatsApp, signal for Signal
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
         $API_BASE_URL = 'https://api.shortmesh.com';
         $ACCESS_TOKEN = 'your_access_token_here';
         $USERNAME = 'your_username';
         $PLATFORM = 'wa'; // wa for WhatsApp, signal for Signal
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