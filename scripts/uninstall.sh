#!/bin/bash

# Stop and disable the imail service
systemctl stop imail || true
systemctl disable imail || true

# Remove systemd service file
rm -f /etc/systemd/system/imail.service

# Remove init.d script (if it exists)
if [ -f /etc/init.d/imail ]; then
    rm -f /etc/init.d/imail
fi

# Remove installation directory
rm -rf /usr/local/imail

# Reload systemd daemon
systemctl daemon-reload

echo "imail uninstalled successfully."