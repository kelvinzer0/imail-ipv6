#!/bin/bash

# Stop and disable the imail service
sudo systemctl stop imail || true
sudo systemctl disable imail || true

# Remove systemd service file
sudo rm -f /etc/systemd/system/imail.service

# Remove init.d script (if it exists)
if [ -f /etc/init.d/imail ]; then
    sudo rm -f /etc/init.d/imail
fi

# Remove installation directory
sudo rm -rf /usr/local/imail

# Reload systemd daemon
sudo systemctl daemon-reload

echo "imail uninstalled successfully."
