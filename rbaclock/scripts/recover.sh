#!/bin/bash
VBoxManage controlvm "126-vbox" poweroff
VBoxManage controlvm "126-vbox-m02" poweroff
VBoxManage controlvm "126-vbox-m03" poweroff
VBoxManage controlvm "126-vbox-m04" poweroff
VBoxManage controlvm "126-vbox-m05" poweroff
VBoxManage controlvm "126-vbox-m06" poweroff
VBoxManage controlvm "126-vbox-m07" poweroff
VBoxManage controlvm "126-vbox-m08" poweroff
VBoxManage snapshot "126-vbox" restore "init"
VBoxManage snapshot "126-vbox-m02" restore "init"
VBoxManage snapshot "126-vbox-m03" restore "init"
VBoxManage snapshot "126-vbox-m04" restore "init"
VBoxManage snapshot "126-vbox-m05" restore "init"
VBoxManage snapshot "126-vbox-m06" restore "init"
VBoxManage snapshot "126-vbox-m07" restore "init"
VBoxManage snapshot "126-vbox-m08" restore "init"
VBoxManage startvm "126-vbox" --type headless
VBoxManage startvm "126-vbox-m02" --type headless
VBoxManage startvm "126-vbox-m03" --type headless
VBoxManage startvm "126-vbox-m04" --type headless
VBoxManage startvm "126-vbox-m05" --type headless
VBoxManage startvm "126-vbox-m06" --type headless
VBoxManage startvm "126-vbox-m07" --type headless
VBoxManage startvm "126-vbox-m08" --type headless
