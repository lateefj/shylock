#!/usr/bin/env python3
"""
Script to render build configuration files
"""

import os
import string
import datetime


def render_file(c, src, dest):
    """
    Simple function to rendering packaging files and scripts
    """
    f = open(src).read()
    t = string.Template(f)
    open(dest, 'w').write(t.substitute(c))

# Read version from file
version = open('VERSION').read()

# Base configruation
conf = {
        "version": version,
        "app_name":"shylock", 
        "year": datetime.date.year,
        "description": "shylock is an attempt to bring sanity to distribute system API's that act just like a file system"
        }

# deb / apt configuration
deb = {
        "distro_box":"ubuntu/xenial64",
        "package_manager":"deb"
        }

# List of package managers custom configurations
managers = [deb]

# Generate all the package manager custom files
for pm in managers:
    tmp = {}
    tmp.update(conf)
    tmp.update(pm)

    build_path = 'build/{0}'.format(tmp['package_manager'])
    if not os.path.exists(build_path):
        os.makedirs(build_path)
    # Vagrant file render  
    render_file(tmp, 'packaging/Vagrantfile', '{0}/Vagrantfile'.format(build_path))
    # JSON render file
    render_file(tmp, 'packaging/pm.json', '{0}/{1}.json'.format(build_path, tmp['package_manager']))
    # Setup script
    render_file(tmp, 'packaging/{0}_setup.sh'.format(tmp['package_manager']), '{0}/{1}_setup.sh'.format(build_path, tmp['package_manager']))

