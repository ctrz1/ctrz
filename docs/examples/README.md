This folder provides an example webserver that can be used to test/showcase `ctrz`. Download the `webserver` binary and run:

```bash
ctrz run --name demo -p 8443:8443 --cpu 50 -d -- webserver
```

This command will start the webserver in an isolated container. Alternatively to passing several command line options, a config file can be used to describe a desired container. E.g.:

```yaml
container:
  demo:
    ports:
    - "8443:8443"
    cpu: 50
    detached: true
    command:
    - "webserver"
```

This is fully optional and not required to run a container. By default `ctrz` will check for a file called `ctrz.yaml` in the current directory. If the file is not found, no error will be thrown. If a config file lives in a different path, it can be used with the `--config` flag. The config hierarchy is `CLI > ctrz.yaml > defaults`. For every flag specified both in the cli and the yaml config, the cli parameter will be used.

You can run `ctrz ps` to see all runing, stopped and dangling containers. You should see output similar to:

```
Name    PID     IP             Created                Command        Status
demo    2434    10.200.1.19    01/09/2026 19:13:19    [webserver]    running
```

You can now visit `localhost:8443` to see the homepage. 

Run `ctrz daemon start` to start the ctrz daemon process. This is a simple daemon process that will collect stats of all running containers. To see live stats, run `ctrz status demo`. In a different window, visit (or curl) `localhost:8443/burn`. This will start burning CPU for 60 seconds (the webserver also supports query parameters to indicate how long to burn CPU and how many workers to use `/burn?seconds=30&workers=3`). On the `ctrz status` tab, notice how the CPU usage does not exceed `50%`. This is because of the `--cpu 50` flag in our run command. If you are accessing the website from a non-localhost address (e.g. http://raspberrypi.local:8443) you will also see a new established connection. To get the stats collected by the daemon process, simply run:

```bash
ctrz status demo --no-header -o -
```

`-` will print the output to standard out. Optionally you can pass in a filename. If the file you give as an argument already exists an error will be thrown. If you want to override or append to an existing file, use `-o - > someFile` or `-o - >> someFile`.

To get more information about the container, run `ctrz info demo`. This will print an output similar to:

```json
{
  "pid": 2434,
  "containerSpec": {
    "name": "demo",
    "cpu": "50000 100000",
    "command": [
      "webserver"
    ],
    "remove": false,
    "detached": true,
    "ports": [
      "8443:8443"
    ]
  },
  "startTime": 234924,
  "started": 1788286399,
  "cgroup": "/sys/fs/cgroup/ctrz-2434",
  "network": {
    "ports": [
      {
        "hostPort": 8443,
        "containerPort": 8443
      }
    ],
    "IP": "10.200.1.19"
  },
  "process": {
    "pid": 2434,
    "comm": "exe",
    "state": 83,
    "ppid": 2428,
    "pgrp": 2428,
    "session": 2427,
    "tty_nr": 34817,
    "tpgid": 2428,
    "flags": 4194560,
    "minflt": 208,
    "cminflt": 0,
    "majflt": 0,
    "cmajflt": 0,
    "utime": 2,
    "stime": 0,
    "cutime": 0,
    "cstime": 0,
    "priority": 20,
    "nice": 0,
    "num_threads": 5,
    "itrealvalue": 0,
    "starttime": 234924,
    "vsize": 1260421120,
    "rss": 361,
    "rsslim": 18446744073709551615,
    "startcode": 65536,
    "endcode": 2082932,
    "startstack": 140737439575664,
    "kstkesp": 0,
    "kstkeip": 0,
    "signal": 0,
    "blocked": 0,
    "sigignore": 0,
    "sigcatch": 2143420159,
    "wchan": 0,
    "nswap": 0,
    "cnswap": 0,
    "exit_signal": 17,
    "processor": 2,
    "rt_priority": 0,
    "policy": 0,
    "delayacct_blkio_ticks": 0,
    "guest_time": 0,
    "cguest_time": 0,
    "start_data": 4587520,
    "end_data": 5462880,
    "start_brk": 917323776,
    "arg_start": 140737439577844,
    "arg_end": 140737439577898,
    "env_start": 140737439577898,
    "env_end": 140737439580137,
    "exit_code": 0
  }
}

```

You can also enter the container by using `ctrz enter demo`. This will enter an interactive shell in which you can move around/manipulate the container's file system. If you don't want to enter an interactive shell but just need to execute a command in the container, run `ctrz enter demo -c "command"`. If the command contains a whitespace (e.g. `ls -lah`), make sure to use quotes around it.

Lastly you can stop the container by running `ctrz stop demo`. When rerunning the `ps` command, you will notice the container has been stopped. To restart the container, simply run `ctrz start demo`.

````
Name    PID     IP             Created                Command        Status
demo    2434    10.200.1.19    01/09/2026 19:13:19    [webserver]    stopped (19)
````


The container can be removed and stopped by running `ctrz rm --name demo`. This will stop the webserver and cleanup the container. To also stop the daemon process, run `sudo ctrz daemon stop`.

Notice how all assets for the `webserver` are embedded in the binary. This is not always possible and sometimes supporting files need to be added. The only way to support this currently is by passing a `tar.gz` archive as:

```bash
ctrz run --name demo -p 8443:8443 --cpu 50 -d -- webserver.tar.gz
```

The archive should follow a structure like this, where the executable is named `bin`:

```bash
.
├── bin       <-- Executable
├── someFile
└── someFolder
    └── someConfig
```

The archive can be created with `tar -czf archive.tar.gz archive`. The archive will then be copied into the container as is. Note, that the archive cannot contain hardlinks, devices, fifos, etc. Only files, directories and symlinks are supported. 
