#!/bin/bash

#echo powersave >/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor




#Logging Function
SCRIPT=`basename ${BASH_SOURCE[0]}`
STX_LOG="/var/log/stratux.log"
function wLog () {
	echo "$(date +"%Y/%m/%d %H:%M:%S")  - $SCRIPT - $1" >> ${STX_LOG}
}
wLog "Running Stratux Updater Script."


# Fix for https://github.com/RPi-Distro/pi-bluetooth/issues/8
# This is a workaround for the bluetooth stack not starting properly
BleGPSEnabled=$(cat /boot/stratux.conf | jq '.BleGPSEnabled // false')
#if [ "$BleGPSEnabled" = "true" ]; then
        wLog "Restarting bluetooth stack"
		# hciconfig hci0 to check status
        hciconfig hci0 down
		bluetoothctl power on
		rfkill unblock all
		# https://stackoverflow.com/questions/24945620/excessive-bluetooth-le-timeouts-on-linux
		echo 2000 > /sys/kernel/debug/bluetooth/hci0/supervision_timeout
		systemctl restart bluetooth
#fi

SCRIPT_MASK="update*stratux*v*.sh"
TEMP_LOCATION="/boot/StratuxUpdates/$SCRIPT_MASK"
UPDATE_LOCATION="/root/$SCRIPT_MASK"

if [ -e ${TEMP_LOCATION} ]; then
	wLog "Found Update Script in $TEMP_LOCATION$SCRIPT_MASK"
	TEMP_SCRIPT=`ls -1t ${TEMP_LOCATION} | head -1`
	wLog "Moving Script $TEMP_SCRIPT"
	cp -r ${TEMP_SCRIPT} /root/
	wLog "Changing permissions to chmod a+x $UPDATE_LOCATION"
	chmod a+x ${UPDATE_LOCATION}
	wLog "Removing Update file from $TEMP_LOCATION"
	rm -rf ${TEMP_SCRIPT}
fi

# Check if we need to run an update.
if [ -e ${UPDATE_LOCATION} ]; then
	UPDATE_SCRIPT=`ls -1t ${UPDATE_LOCATION} | head -1`
	if [ -n ${UPDATE_SCRIPT} ] ; then
		# Execute the script, remove it, then reboot.
		wLog "Running update script ${UPDATE_SCRIPT}..."
		bash ${UPDATE_SCRIPT}
		wLog "Removing Update SH"
		rm -f ${UPDATE_SCRIPT}
		wLog "Finished... Rebooting... Bye"
		reboot
	fi
fi
wLog "Exited without updating anything..."
