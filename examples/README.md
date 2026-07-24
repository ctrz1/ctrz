This folder provides an example webserver that can be used to test/showcase `ctrz`. Download the `webserver` binary and run:

```bash
sudo ctrz run --name demo -p 8443:8443 --cpu 50 -d -- webserver
```

This command will start the webserver in an isolated container. You can run `sudo ctrz ps` to see all active and inactive containers. You should see output similar to:

```bash
Name    PID     IP              Created                Command           Status
demo    2185    10.200.1.106    23/07/2026 22:15:56    webserver         active
```

You can now visit `localhost:8443` to see the homepage. 

Run `sudo ctrz daemon start` to start the ctrz daemon process. This is a simple daemon process that will collect stats of all running containers. To see live stats, run `sudo ctrz status demo`. In a different window, visit (or curl) `localhost:8443/burn`. This will start burning CPU for 60 seconds (the webserver also supports query parameters to indicate how long to burn CPU and how many workers to use `/burn?seconds=30&workers=3`). On the `ctrz status` tab, notice how the CPU usage does not exceed `50%`. This is because of the `--cpu 50` flag in our run command. If you are accessing the website from a non-localhost address (e.g. http://raspberrypi.local:8443) you will also see a new established connection. To get the stats collected by the daemon process, simply run:

```bash
sudo ctrz status demo --no-header -o -
```

`-` will print the output to standard out. Optionally you can pass in a filename. If the file you give as an argument already exists an error will be thrown. If you want to override or append to an existing file, use `-o - > someFile` or `-o - >> someFile`.

To get more information about the container, run `sudo ctrz info demo`. This will print an output similar to:

```json
{
  "pid": 2413,
  "name": "demo",
  "startTime": 1784841648,
  "command": "webserver",
  "cgroup": "/sys/fs/cgroup/ctrz-2413",
  "ContainerIP": "10.200.1.42",
  "containerPort": [
    8443
  ],
  "hostPort": [
    8443
  ]
}
```

You can also enter the container by using `sudo ctrz enter demo`. This will enter an interactive shell in which you can move around/manipulate the container's file system. If you don't want to enter an interactive shell but just need to execute a command in the container, run `sudo ctrz enter demo -c "command"`. If the command contains a whitespace (e.g. `ls -lah`), make sure to use quotes around it.

Lastly you can remove and stop the container by running `sudo ctrz rm --name demo`. This will stop the webserver and cleanup the container. Dedicated `start` and `stop` commands have yet to be implemented. To also stop the daemon process, run `sudo ctrz daemon stop`.

Notice how all assets for the `webserver` are embedded in the binary. This is not always possible and sometimes supporting files need to be added. The only way to support this currently is by passing a `tar.gz` archive as:

```bash
sudo ctrz run --name demo -p 8443:8443 --cpu 50 -d -- webserver.tar.gz
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
