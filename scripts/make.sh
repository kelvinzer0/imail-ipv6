#!/bin/bash

_os=`uname`
_path=`pwd`
_dir=`dirname $_path`


sed "s:{APP_PATH}:${_dir}:g" $_dir/scripts/init.d/imail.service.tpl | sudo tee /etc/systemd/system/imail.service > /dev/null

echo `dirname $_path`