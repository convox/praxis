#!/bin/bash

set -ex

script -q /dev/null cx run -a httpd --rack noah/dev2 web ping -c20 google.com &
script -q /dev/null cx run -a httpd --rack noah/dev2 web ping -c20 google.com &
script -q /dev/null cx run -a httpd --rack noah/dev2 web ping -c20 google.com &
script -q /dev/null cx run -a httpd --rack noah/dev2 web ping -c20 google.com &
script -q /dev/null cx run -a httpd --rack noah/dev2 web ping -c20 google.com &

wait