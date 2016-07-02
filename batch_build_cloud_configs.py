#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""Batch build cloud-configs

This program is used to batch build multiple CoreOS
initialized with the cloud-configs configuration file.
Because differentiation exists for each cloud-config identity.
Uses pyyaml and jinja2 two extension modules.

@Create on : 16/7/2 10:55
@Author    : Li Wen Shan
@version   : 0.1
"""

from yaml import load
from jinja2 import Environment, FileSystemLoader
import os
import sys
import logging

# Setting logger
logging.basicConfig(format='[%(asctime)s](%(levelname)s)%(name)s: %(message)s', level=logging.INFO)
logger = logging.getLogger('builder')

# Read the configuration file
logger.info("Read the configuration file...")
if not os.path.isfile('build_config.yml'):
    logger.error('build_config.yml file not found!')
    sys.exit(1)

with open('build_config.yml', 'r') as f:
    b = f.read()
    build_config = load(b)

# Start batch create
for k, v in build_config.iteritems():
    own_ip = k
    mac_address = v['MAC']
    hostname = v['hostname']
    etcd2 = v['etcd2']
    nic_name = v['nic_name']
    logger.info("Write file %s.yml ..." % mac_address)
    env = Environment(loader=FileSystemLoader('./'))
    body = env.get_template('cloud-config.template').render(own_ip=own_ip,
                                                            mac_address=mac_address,
                                                            hostname=hostname,
                                                            etcd2=etcd2,
                                                            nic_name=nic_name)
    with open("%s.yml" % mac_address, "w") as f:
        f.write(body)

logger.info("Batch create configuration file is completed.")
