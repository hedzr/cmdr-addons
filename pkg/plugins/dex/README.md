# dex

`dex` is a new daemon plugin for `cmdr`. It's compatible with both windows, linux and macOS.

For more information, see  
[the example app: service](https://github.com/hedzr/cmdr-examples/tree/master/examples/service)




## Install/Uninstall your service under linux distro

### CentOS/RedHat/Ubuntu/Debian/...

```bash
sudo /path/to/app server install
sudo /path/to/app server uninstall
```

After installed, check out the service file at /etc/systemd/system/app-service-name.service.

These files/folders are compliant to standard systemd layouts:

- Log file:
  - /var/log/`app-service-name`/`app-service-name`.out
  - /var/log/`app-service-name`/`app-service-name`.err
- Working Directory:
  - /var/lib/`app-service-name`
- Configurations:
  The configuration files must be copied to /etc ...
  - etc:
	- /etc/`app-service-name`/`app-service-name`.yml
	- /etc/`app-service-name`/conf.d/*.yml
	- /etc/`app-service-name`/certs: the certification files for web server
  - default:
	the default config options can be written to these locations
	- /etc/default/`app-service-name`
	for some distros, the location might be `/etc/sysconfig/app-service-name`.

### SELinux PRB `203/EXEC`

For CentOS and others SELinux enabled distros, you might get the error message or status about `app-service-name.service: Main process exited, code=exited, status=203/EXEC`, if you're running/starting up the service from a nonstandard location such as /home/you/go/bin/xxx.

In `/var/log/messages`, the slight clear information can be found like:

```
May  2 23:00:16 c8 systemd[21010]: my-service.service: Failed to execute command: Permission denied
May  2 23:00:16 c8 systemd[21010]: ny-service.service: Failed at step EXEC spawning /home/worker/src/backend/bin/my-service: Permission denied
```

So the reason is clear, the SELinux restricts binaries that can be used in ExecStart to paths that has `system_u:object_r:bin_t:s0` attribute set. As we know, typically those are /usr/bin /usr/sbin /usr/libexec /usr/local/bin directories but `/home/worker/src/backend/bin` is not.

The solution is:

```bash
chcon -R -t bin_t /home/you/go/bin/
```

Sometimes you might wanna revoke it:

```bash
restorecon -r -v /home/you/go/bin/
```

> 
> 


### start/stop the service

Use `systemctl`:

```bash
sudo systemctl start|stop|restart your-service.service

# start the service with OS starting?
sudo systemctl enable|disable your-service.service
```

Or start/stop/restart the service with its binary executable:

```bash
sudo /path/to/app server start|stop|restart
```

### run from console

You can run the service at console mode:

```bash
/path/to/app server run
# Or: /path/to/app server start -f
```





## dependencies

`dex` wraps ["github.com/kardianos/service"](https://github.com/kardianos/service) to [`cmdr`](https://github/com/hedzr/cmdr) system.




