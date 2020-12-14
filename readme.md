Usage:
```bash
go get github.com/getevo/watchdog
cd $GOPATH/src/github.com/getevo/watchdog
go build
sudo ./watchdog install
sudo service watchdog status
```
For each service to check put your command like this example:
```yaml
#Sample YML

name: "lvf-api"
command: "service myservice status | grep Active:"
contains: "active (running)"
#regex: ".*"
#onmatch: "echo hi"
else: "service myservice restart"
interval: "3s"
debug: "false"
```