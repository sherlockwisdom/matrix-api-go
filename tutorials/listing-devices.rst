Listing Devices
==============

You can list all devices for a specific platform to see what devices are available for messaging.

**Endpoint**: ``POST /{platform}/list/devices``

**Headers**:
- ``Authorization: Bearer YOUR_ACCESS_TOKEN``
- ``Content-Type: application/json``

**Request Body**: 
The request body contains your username. For detailed request body specifications, see the complete API documentation at https://api.shortmesh.com/

**Response**: 
The response contains an array of devices with their IDs, platforms, and status. For detailed response specifications, see the complete API documentation at https://api.shortmesh.com/

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
         const API_BASE_URL = 'https://api.shortmesh.com';
         const ACCESS_TOKEN = 'your_access_token_here';
         const USERNAME = 'your_username';
         const PLATFORM = 'wa'; // wa for WhatsApp, signal for Signal

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
             API_BASE_URL = "https://api.shortmesh.com"
             ACCESS_TOKEN = "your_access_token_here"
             USERNAME     = "your_username"
             PLATFORM     = "wa" // wa for WhatsApp, signal for Signal
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
         $API_BASE_URL = 'https://api.shortmesh.com';
         $ACCESS_TOKEN = 'your_access_token_here';
         $USERNAME = 'your_username';
         $PLATFORM = 'wa'; // wa for WhatsApp, signal for Signal

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