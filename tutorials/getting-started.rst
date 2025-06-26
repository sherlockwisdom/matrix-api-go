Getting Started
==============

Welcome to the ShortMesh API documentation! This tutorial will guide you through the basic steps to get started with ShortMesh messaging bridge API.

Overview
--------

ShortMesh allows you to send messages across different messaging platforms through a unified API. This tutorial covers the essential steps to set up and use ShortMesh.

Step 1: Login and Access
------------------------

1. **Visit the ShortMesh Web UI**: Go to https://shortmesh.com
2. **Create an account or log in**: Use your credentials to access the web interface
3. **Navigate to the dashboard**: Once logged in, you'll see your ShortMesh dashboard

The web UI provides a user-friendly interface for managing your account, devices, and API access.

Step 2: Add Devices
-------------------

1. **Access the Devices section**: In the web UI, navigate to the Devices area
2. **Add your messaging platforms**: Connect your WhatsApp, Signal, or other supported platforms
3. **Follow platform-specific setup**: Each platform has its own authentication process (QR codes, etc.)
4. **Verify device status**: Ensure your devices are properly connected and ready

For detailed instructions on adding devices, see :doc:`adding-devices`.

Step 3: Get Your API Access Token
--------------------------------

1. **Navigate to API settings**: In the web UI, find the API or Developer section
2. **Generate or copy your access token**: This token will be used to authenticate your API requests
3. **Store the token securely**: Keep this token safe as it provides access to your account

**Important**: Never share your access token publicly or commit it to version control.

Step 4: Send Messages
---------------------

Once you have your devices set up and access token ready, you can start sending messages through the API.

For detailed examples of how to send messages, see :doc:`sending-messages`.

For information on listing your devices via API, see :doc:`listing-devices`.

What's Next?
------------

After completing these basic steps, you can:

- **Explore the API endpoints** for more advanced features
- **Integrate ShortMesh into your applications** using the provided code examples
- **Set up automated messaging workflows** for your business needs
- **Monitor your message delivery** through the web UI

All device management and authentication is handled through the web UI at https://shortmesh.com, while the API is used for sending messages and retrieving device information. 