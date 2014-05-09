#!/bin/bash
set -x
go build dasheen.go
scp dasheen atomic@radiator:dasheen/dasheen
scp web/* atomic@radiator:dasheen/web/
scp com.atomicobject.dasheen.plist atomic@radiator:dasheen/
ssh atomic@radiator 'sudo mv dasheen/com.atomicobject.dasheen.plist /Library/LaunchDaemons/com.atomicobject.dasheen.plist'
ssh atomic@radiator 'sudo chown root:wheel /Library/LaunchDaemons/com.atomicobject.dasheen.plist'
ssh atomic@radiator 'sudo chmod 644 /Library/LaunchDaemons/com.atomicobject.dasheen.plist'
ssh atomic@radiator 'sudo launchctl unload /Library/LaunchDaemons/com.atomicobject.dasheen.plist'
ssh atomic@radiator 'sudo launchctl load /Library/LaunchDaemons/com.atomicobject.dasheen.plist'
