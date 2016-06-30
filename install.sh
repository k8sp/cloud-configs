#!/bin/sh

mac_addr=`ifconfig | grep -A2 'broadcast' | grep -o '..:..:..:..:..:..'`
wget http://10.10.10.182/cloud-configs/${mac_addr}.yml
sudo coreos-install -d /dev/sda -c ${mac_addr}.yml -b http://10.10.10.192
sudo reboot
