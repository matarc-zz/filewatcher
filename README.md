# Filewatcher
* [Introduction](#introduction)
* [Installation](#installation)
* [Building](#building)
* [Running](#building)
* [Configuration](#configuration)
* [API choices](#api-choices)
* [Limits](#limits)
* [Known Issues](#known-issues)

## Introduction
Filewatcher is a project composed of three disctint elements :
* **Masterserver** : A server that is in charge of handling HTTP request through a REST API that will send you the list of all files watched by the Nodewatchers.
* **Storage** : A server that is in charge of storing the list of all Nodewatchers.
* **Nodewatcher** : A client that list all files in a directory and its subdirectories and send the list as well as all updates of that list to Storage.

Filewatcher can currently run on Mac and Linux, although the developement took mostly place on mac, linux wasn't tested as much but it seems to build and run.

This was developed with go version 1.10.3, I do not know if version prior to that will work or not.

## Installation
To get started install filewatcher with the following command :
```
go get -u github.com/matarc/filewatcher/...
```

## Building
To build the three different elements run the following commands :
```
cd cmd
for i in nodewatcher masterserver storage; do go build -tags "$i" -o "$i"; done
```
This will build three executables in the `cmd` folder named nodewatcher, masterserver and storage.

## Running
By default you can already run all executables without any arguments and it will work out of the box.

The file watched by nodewatcher will be your temporary directory on your system.

You can query the masterserver using the following command :
```
curl http://localhost:8080/list
```

## Configuration
You can configure the different executables using a configuration file that you can generate with the executable built from the `config` directory.
```
cd config
go build
```
This will generate an executable called `config` that prints the configuration on `stdout`.

You can run config with `-help` to display the usage.

The configuration can be reloaded by a running instance at any point by sending the `SIGUSR1` signal to the running executable.

### Masterserver
To configure masterserver use the following command :
```
./config -config masterserver -address "localhost:4242" -storageAddress "localhost:8484" > masterserver.conf
```
This will create a configuration file `masterserver.conf` and will make masterserver listen on `localhost:4242` for http requests while telling the masterserver to connect to storage on `localhost:8484`.

### Storage
To configure storage use the following command :
```
./config -config storage -dbPath "/path/to/storage/database.db" -address "localhost:8484" > storage.conf
```
This will create a configuration file `storage.conf` and will make storage listen on `localhost:8484` and create/open the database `/path/to/storage/database.db`.

### Nodewatcher
To configure nodewatcher use the following command :
```
./config -config nodewatcher -storageAddress "localhost:8484" -dir "/directory/to/watch" -id "uniqueid" > nodewatcher.conf
```
This will create a configuration file `nodewatcher.conf` and will make nodewatcher connect to storage at `localhost:8484`, watch the directory `/directory/to/watch` and give this nodewatcher the id `uniqueid`.

Ids should be **unique** for every single nodewatcher, if not they will overwrite each other's list on storage.

## API choices
This project uses [github.com/fsnotify/fsnotify](https://github.com/fsnotify/fsnotify), a cross platform library that can watch files and directories on Windows, Linux, BSD and macOS.

However I used a modified version of this library, because of the way they implemented the wrapper on kevent (the macOS/BSD API to watch files), they open a file for every file in a watched directory.

This is a problem because you quickly run out of file descriptors with the default configuration on mac, and we don't need that anyway for this project.

This choice was made with what your next project is in mind (which I believe is directly linked to this test) as you plan to have a 'dropbox-solution-like' running on your client's computer to be used with your current product.

This library would allow you to use the same code for all platforms and your servers.

If however you plan on building a client in something other than go, and only build the server side with go, then using the inotify Linux API directly would make more sense (less possible bugs).

## Limits
There are native system limits to be aware of when it comes to the number of files that can be watched or open :
* Linux : By default inotify allows each user to watch 8192 files. This can easily be [changed](https://askubuntu.com/questions/154255/how-can-i-tell-if-i-am-out-of-inotify-watches) by modifying /proc/sys/fs/inotify/max_user_watches 
* MacOs : By default OS X allows you to open 10,240 files per process and 12,288 total. This can be [changed](http://krypted.com/mac-os-x/maximum-files-in-mac-os-x/)
* Windows : Not sure.

Furthermore a shell adds another limit on the number of file descriptors you can open per process.

You can see that limit with the command `ulimit -n` and modify the limit with the command `ulimit -n 42` (sets the limit to 42).

## Known issues
* If you change the ID of a nodewatcher, the list of the old ID will remain forever on storage and be sent to masterserver (you'll basically get a duplicated list if you don't change the directory).
* If two nodes have the same ID, they will erase each other's list.

