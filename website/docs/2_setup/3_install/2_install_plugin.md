---
id: 2_install_plugin
title: Install Plugin
sidebar_label: Install Plugin
---
You can install the plugin by copying the [file](https://github.com/Orange-OpenSource/casskop/tree/master/plugins/kubectl-casskop) into your PATH.

For example on a linux/ mac machine:

```console
cp plugins/kubectl-casskop /usr/local/bin
```

Then you can test the plugin:

```console
kubectl casskop

usage: kubectl-casskop <command> [<args>]

The available commands are:
   cleanup
   upgradesstables
   rebuild
   remove
   restart
   pause
   unpause

For more information you can run kubectl-casskop <command> --help
kubectl-casskop: error: the following arguments are required: command
```

Your CassKop plugin is now installed!

