Getting Started
==============

Welcome to the ShortMesh API documentation! This guide will help you integrate with the ShortMesh messaging bridge API to send messages across different platforms.

.. note::
   This tutorial assumes you have already:
   
   - Created an account or logged in through the web interface
   - Retrieved your API access token from the dashboard
   - Added platforms to your devices through the web UI (which handles QR code scanning)

Prerequisites
------------

Before using the API, ensure you have:

1. **API Access Token**: Get this from your ShortMesh dashboard after logging in
2. **Platform Setup**: Add your messaging platforms (WhatsApp, Signal, etc.) through the web interface
3. **Device Names**: Note the device names for each platform you've configured

API Base URL
-----------

All API endpoints are relative to the ShortMesh API server:

.. code-block:: text

   https://api.shortmesh.com

Authentication
--------------

All API requests require authentication using a Bearer token in the Authorization header:

.. code-block:: text

   Authorization: Bearer YOUR_ACCESS_TOKEN

Platform Codes
-------------

The following platform codes are supported:

- ``wa`` - WhatsApp
- ``signal`` - Signal

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

Basic Setup Examples
-------------------

.. tabs::

   .. tab:: Python

      .. code-block:: python

         import os
         import requests
         from dotenv import load_dotenv

         # Load environment variables
         load_dotenv()

         # Configuration from environment variables
         API_BASE_URL = os.getenv('SHORTMESH_API_URL', 'https://api.shortmesh.com')
         ACCESS_TOKEN = os.getenv('SHORTMESH_ACCESS_TOKEN')
         USERNAME = os.getenv('SHORTMESH_USERNAME')

         # Validate required configuration
         if not ACCESS_TOKEN:
             raise ValueError("SHORTMESH_ACCESS_TOKEN environment variable is required")
         if not USERNAME:
             raise ValueError("SHORTMESH_USERNAME environment variable is required")

         # Common headers for all requests
         def get_headers():
             return {
                 "Authorization": f"Bearer {ACCESS_TOKEN}",
                 "Content-Type": "application/json"
             }

         # Example: Test API connection
         def test_connection():
             try:
                 # Use the listing devices endpoint as a simple test
                 response = requests.post(
                     f"{API_BASE_URL}/wa/list/devices",
                     headers=get_headers(),
                     json={"username": USERNAME}
                 )
                 
                 if response.status_code == 200:
                     print("✅ API connection successful!")
                     return True
                 else:
                     print(f"❌ API connection failed: {response.status_code}")
                     print(response.json())
                     return False
             except Exception as e:
                 print(f"❌ Connection error: {e}")
                 return False

         if __name__ == "__main__":
             test_connection()

   .. tab:: JavaScript (Node.js)

      .. code-block:: javascript

         require('dotenv').config();
         const axios = require('axios');

         // Configuration from environment variables
         const API_BASE_URL = process.env.SHORTMESH_API_URL || 'https://api.shortmesh.com';
         const ACCESS_TOKEN = process.env.SHORTMESH_ACCESS_TOKEN;
         const USERNAME = process.env.SHORTMESH_USERNAME;

         // Validate required configuration
         if (!ACCESS_TOKEN) {
             throw new Error('SHORTMESH_ACCESS_TOKEN environment variable is required');
         }
         if (!USERNAME) {
             throw new Error('SHORTMESH_USERNAME environment variable is required');
         }

         // Common headers for all requests
         function getHeaders() {
             return {
                 'Authorization': `Bearer ${ACCESS_TOKEN}`,
                 'Content-Type': 'application/json'
             };
         }

         // Example: Test API connection
         async function testConnection() {
             try {
                 // Use the listing devices endpoint as a simple test
                 const response = await axios.post(
                     `${API_BASE_URL}/wa/list/devices`,
                     { username: USERNAME },
                     { headers: getHeaders() }
                 );
                 
                 console.log('✅ API connection successful!');
                 return true;
             } catch (error) {
                 console.error('❌ API connection failed:', error.response?.status);
                 if (error.response?.data) {
                     console.error(error.response.data);
                 }
                 return false;
             }
         }

         // Run the test
         testConnection();

   .. tab:: PHP

      .. code-block:: php

         <?php

         // Load environment variables (requires vlucas/phpdotenv package)
         $dotenv = Dotenv\Dotenv::createImmutable(__DIR__);
         $dotenv->load();

         // Configuration from environment variables
         $API_BASE_URL = $_ENV['SHORTMESH_API_URL'] ?? 'https://api.shortmesh.com';
         $ACCESS_TOKEN = $_ENV['SHORTMESH_ACCESS_TOKEN'] ?? null;
         $USERNAME = $_ENV['SHORTMESH_USERNAME'] ?? null;

         // Validate required configuration
         if (!$ACCESS_TOKEN) {
             throw new Exception('SHORTMESH_ACCESS_TOKEN environment variable is required');
         }
         if (!$USERNAME) {
             throw new Exception('SHORTMESH_USERNAME environment variable is required');
         }

         // Common headers for all requests
         function getHeaders($accessToken) {
             return [
                 'Authorization: Bearer ' . $accessToken,
                 'Content-Type: application/json'
             ];
         }

         // Example: Test API connection
         function testConnection($apiBaseUrl, $accessToken, $username) {
             $ch = curl_init();
             
             curl_setopt($ch, CURLOPT_URL, $apiBaseUrl . '/wa/list/devices');
             curl_setopt($ch, CURLOPT_POST, true);
             curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode(['username' => $username]));
             curl_setopt($ch, CURLOPT_HTTPHEADER, getHeaders($accessToken));
             curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
             curl_setopt($ch, CURLOPT_HTTP_VERSION, CURL_HTTP_VERSION_1_1);
             
             $response = curl_exec($ch);
             $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
             curl_close($ch);
             
             if ($httpCode == 200) {
                 echo "✅ API connection successful!\n";
                 return true;
             } else {
                 echo "❌ API connection failed: " . $httpCode . "\n";
                 echo "Response: " . $response . "\n";
                 return false;
             }
         }

         // Run the test
         testConnection($API_BASE_URL, $ACCESS_TOKEN, $USERNAME);

         ?>

Example Environment Setup
------------------------

.. code-block:: bash

   # .env file
   SHORTMESH_API_URL=https://api.shortmesh.com
   SHORTMESH_ACCESS_TOKEN=your_access_token_here
   SHORTMESH_USERNAME=your_username 