#!/bin/bash

for x in $(kns | egrep "group|main-"); do echo $x ; k delete namespace --grace-period=0 --force $x ; done
