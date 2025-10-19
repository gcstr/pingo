#!/bin/sh
# Create data directories
mkdir -p /var/lib/pingo

# Reload systemd
systemctl daemon-reload

# Enable and start service
systemctl enable pingo
systemctl start pingo

echo "Pingo has been installed and started."
echo "Dashboard: http://localhost:7777"
echo "Config: /etc/pingo/config.toml.example"
echo "Service: systemctl status pingo"
