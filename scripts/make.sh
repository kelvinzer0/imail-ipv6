#!/bin/bash

_os=`uname`
_path=`pwd`
_dir=`dirname $_path`


sed "s:{APP_PATH}:${_dir}:g" $_dir/scripts/init.d/imail.service.tpl > /etc/systemd/system/imail.service

echo `dirname $_path`